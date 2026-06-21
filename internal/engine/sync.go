package engine

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

const (
	gitignoreMarkerStart = "# --- auto-generated from sync.toml (do not edit below) ---"
	gitignoreMarkerEnd   = "# --- end auto-generated ---"
)

var categoryNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_.-]*$`)

func generateGitignore(config *Config) error {
	path := repositoryPath(config, ".gitignore")
	manualSection, err := readManualGitignore(path)
	if err != nil {
		return err
	}

	var output strings.Builder
	if manualSection != "" {
		output.WriteString(strings.TrimRight(manualSection, "\r\n"))
		output.WriteString("\n")
	}
	output.WriteString(gitignoreMarkerStart + "\n\n")
	output.WriteString("# Security exclusions\n")
	output.WriteString("*auth*\n*.key\n*.pem\n.env*\n\n")
	output.WriteString("# Ignored categories (ignore in sync.toml)\n")
	for _, category := range config.Sync.Ignore {
		output.WriteString(category + "/\n")
	}
	output.WriteString("\n.conflict-pending\n.dotfile-hook/\n")
	output.WriteString(gitignoreMarkerEnd + "\n")

	if err := os.WriteFile(path, []byte(output.String()), 0o644); err != nil {
		return fmt.Errorf(".gitignoreを書き込めません: %w", err)
	}
	return nil
}

func readManualGitignore(path string) (string, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf(".gitignoreを開けません: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Text() == gitignoreMarkerStart {
			break
		}
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf(".gitignoreを読み込めません: %w", err)
	}
	return strings.Join(lines, "\n"), nil
}

func status(config *Config, cmd *cobra.Command) error {
	if _, err := os.Stat(repositoryPath(config, ".conflict-pending")); err == nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "")
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "========================================")
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  [dotfile] CONFLICT PENDING")
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Run: cd %s && git log --oneline --graph --all\n", config.DotfilesDir)
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "========================================")
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "")
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("コンフリクト状態を確認できません: %w", err)
	}
	return nil
}

func pull(config *Config, cmd *cobra.Command) error {
	git := GitRunner{WorkDir: config.DotfilesDir, Stdout: cmd.OutOrStdout(), Stderr: cmd.ErrOrStderr()}
	if err := clearResolvedConflictMarker(config, git, cmd.OutOrStdout()); err != nil {
		return err
	}
	if err := git.Run("fetch", "--quiet", "origin"); err != nil {
		return err
	}

	remoteRef := "origin/" + config.Sync.DefaultBranch
	remoteHead, err := git.Output("rev-parse", "--verify", remoteRef)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[sync] リモートブランチがありません: %s\n", remoteRef)
		return nil
	}
	localHead, err := git.Output("rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if localHead == remoteHead {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "[sync] Already up to date.")
		return nil
	}

	mergeBase, err := git.Output("merge-base", "HEAD", remoteRef)
	if err != nil {
		return err
	}
	if localHead == mergeBase {
		if err := git.Run("merge", "--ff-only", remoteRef); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[sync] Fast-forwarded to %s.\n", remoteRef)
		return nil
	}
	if remoteHead == mergeBase {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "[sync] ローカルがリモートより先行しています。pullをスキップします。")
		return nil
	}

	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}
	conflictBranch := fmt.Sprintf("conflict/%s/%s", sanitizeBranchPart(host), time.Now().Format("20060102-150405"))
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[sync] 分岐を検出しました。退避ブランチを作成: %s\n", conflictBranch)
	if err := git.Run("checkout", "-b", conflictBranch); err != nil {
		return err
	}
	dirty, err := git.Output("status", "--porcelain")
	if err != nil {
		return err
	}
	if dirty != "" {
		if err := git.Run("add", "-A"); err != nil {
			return err
		}
		if err := git.Run("commit", "-m", "auto-save: uncommitted changes before conflict resolution"); err != nil {
			return err
		}
	}
	if err := git.Run("checkout", config.Sync.DefaultBranch); err != nil {
		return err
	}
	if err := git.Run("reset", "--hard", remoteRef); err != nil {
		return err
	}
	if err := os.WriteFile(repositoryPath(config, ".conflict-pending"), nil, 0o644); err != nil {
		return fmt.Errorf(".conflict-pendingを作成できません: %w", err)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[sync] ローカル変更は%sへ退避し、%sを%sへ戻しました。\n", conflictBranch, config.Sync.DefaultBranch, remoteRef)
	return nil
}

func clearResolvedConflictMarker(config *Config, git GitRunner, stdout io.Writer) error {
	marker := repositoryPath(config, ".conflict-pending")
	if _, err := os.Stat(marker); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	branches, err := git.Output("branch", "--list", "conflict/*")
	if err != nil {
		return err
	}
	if branches == "" {
		if err := os.Remove(marker); err != nil {
			return fmt.Errorf(".conflict-pendingを削除できません: %w", err)
		}
		_, _ = fmt.Fprintln(stdout, "[sync] コンフリクト解消を確認し、マーカーを削除しました。")
	}
	return nil
}

func sanitizeBranchPart(value string) string {
	replacer := strings.NewReplacer(" ", "-", "~", "-", "^", "-", ":", "-", "?", "-", "*", "-", "[", "-", "\\", "-", "..", "-")
	return strings.Trim(replacer.Replace(value), "./")
}

func push(config *Config, cmd *cobra.Command) error {
	git := GitRunner{WorkDir: config.DotfilesDir, Stdout: cmd.OutOrStdout(), Stderr: cmd.ErrOrStderr()}
	currentBranch, err := git.Output("branch", "--show-current")
	if err != nil {
		return err
	}
	if currentBranch != config.Sync.DefaultBranch {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[sync] %sブランチではありません (%s)。自動pushをスキップします。\n", config.Sync.DefaultBranch, currentBranch)
		return nil
	}

	var autoPaths []string
	var missing []string
	for _, category := range config.Sync.Auto {
		if info, statErr := os.Stat(repositoryPath(config, category)); statErr == nil && info.IsDir() {
			autoPaths = append(autoPaths, category)
			continue
		}
		if trackedCategory(git, category) {
			missing = append(missing, category)
		}
	}
	if len(missing) > 0 {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "[sync] WARNING: 追跡済みの自動同期カテゴリが見つかりません:")
		for _, category := range missing {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  - %s\n", category)
		}
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "誤削除ならgit restore、恒久削除ならdotfile delete-categoryを使用してください。")
	}

	for _, category := range autoPaths {
		if err := git.Run("add", "--", category+"/"); err != nil {
			return err
		}
	}
	if len(autoPaths) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "[sync] 自動同期カテゴリに変更はありません。")
		return nil
	}
	diffArgs := append([]string{"diff", "--cached", "--name-only", "--"}, autoPaths...)
	changed, err := git.Output(diffArgs...)
	if err != nil {
		return err
	}
	if changed == "" {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "[sync] 自動同期カテゴリに変更はありません。")
		return nil
	}

	message, err := generateCommitMsg(git, autoPaths)
	if err != nil {
		return err
	}
	commitArgs := append([]string{"commit", "--only", "-m", message, "--"}, autoPaths...)
	if err := git.Run(commitArgs...); err != nil {
		return err
	}
	if err := git.Run("push", "origin", config.Sync.DefaultBranch); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[sync] Pushed: %s\n", message)
	return nil
}

func trackedCategory(git GitRunner, category string) bool {
	if output, err := git.Output("ls-files", "--", category); err == nil && output != "" {
		return true
	}
	output, err := git.Output("ls-tree", "-r", "--name-only", "HEAD", "--", category)
	return err == nil && output != ""
}

func generateCommitMsg(git GitRunner, paths []string) (string, error) {
	type change struct {
		filter string
		prefix string
	}
	changes := []change{
		{filter: "A", prefix: "add"},
		{filter: "M", prefix: "update"},
		{filter: "D", prefix: "delete"},
	}
	var parts []string
	for _, item := range changes {
		args := []string{"diff", "--cached", "--diff-filter=" + item.filter, "--name-only", "--"}
		args = append(args, paths...)
		output, err := git.Output(args...)
		if err != nil {
			return "", err
		}
		names := uniqueBaseNames(output)
		if len(names) > 0 {
			parts = append(parts, fmt.Sprintf("%s: %s", item.prefix, strings.Join(names, ", ")))
		}
	}
	if len(parts) == 0 {
		return "sync: no changes", nil
	}
	return strings.Join(parts, "; "), nil
}

func uniqueBaseNames(output string) []string {
	seen := make(map[string]struct{})
	for path := range strings.SplitSeq(output, "\n") {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		seen[filepath.Base(filepath.FromSlash(path))] = struct{}{}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func deleteCategory(config *Config, category string, cmd *cobra.Command) error {
	if !categoryNamePattern.MatchString(category) || category == "." || category == ".." {
		return fmt.Errorf("不正なカテゴリ名です: %s", category)
	}
	git := GitRunner{WorkDir: config.DotfilesDir, Stdout: cmd.OutOrStdout(), Stderr: cmd.ErrOrStderr()}
	currentBranch, err := git.Output("branch", "--show-current")
	if err != nil {
		return err
	}
	if currentBranch != config.Sync.DefaultBranch {
		return fmt.Errorf("カテゴリ削除は%sブランチでのみ実行できます (current: %s)", config.Sync.DefaultBranch, currentBranch)
	}
	if !contains(config.Sync.Auto, category) {
		return fmt.Errorf("自動同期カテゴリではありません: %s", category)
	}
	if !git.Success("diff", "--quiet", "--", "sync.toml") ||
		!git.Success("diff", "--cached", "--quiet", "--", "sync.toml") {
		return fmt.Errorf("sync.tomlに未コミットの変更があります")
	}

	hadTracked := trackedCategory(git, category)
	config.Sync.Auto = without(config.Sync.Auto, category)
	if err := writeSyncConfig(repositoryPath(config, "sync.toml"), config.Sync); err != nil {
		return err
	}
	if err := git.Run("reset", "-q", "HEAD", "--", category); err != nil && hadTracked {
		return err
	}
	if err := os.RemoveAll(repositoryPath(config, category)); err != nil {
		return fmt.Errorf("カテゴリを削除できません: %w", err)
	}
	if err := git.Run("add", "--", "sync.toml"); err != nil {
		return err
	}
	commitPaths := []string{"sync.toml"}
	if hadTracked {
		if err := git.Run("add", "-A", "--", category); err != nil {
			return err
		}
		commitPaths = append(commitPaths, category)
	}
	args := []string{"commit", "--only", "-m", "delete: category " + category, "--"}
	args = append(args, commitPaths...)
	if err := git.Run(args...); err != nil {
		return err
	}
	if err := git.Run("push", "origin", config.Sync.DefaultBranch); err != nil {
		return fmt.Errorf("pushに失敗しました。削除commitはローカルに残っています: %w", err)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[sync] カテゴリを削除してpushしました: %s\n", category)
	return nil
}

func writeSyncConfig(path string, config SyncConfig) error {
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, ".sync.toml-*")
	if err != nil {
		return fmt.Errorf("一時ファイルを作成できません: %w", err)
	}
	tempPath := temp.Name()
	cleanup := func() {
		_ = temp.Close()
		_ = os.Remove(tempPath)
	}
	defer cleanup()

	if err := toml.NewEncoder(temp).Encode(config); err != nil {
		return fmt.Errorf("sync.tomlをエンコードできません: %w", err)
	}
	if err := temp.Sync(); err != nil {
		return fmt.Errorf("一時ファイルを同期できません: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("一時ファイルを閉じられません: %w", err)
	}
	if err := replaceFile(path, tempPath); err != nil {
		return err
	}
	return nil
}

func replaceFile(destination, source string) error {
	if runtime.GOOS != "windows" {
		if err := os.Rename(source, destination); err != nil {
			return fmt.Errorf("sync.tomlを置換できません: %w", err)
		}
		return nil
	}

	backup := destination + ".backup"
	_ = os.Remove(backup)
	if err := os.Rename(destination, backup); err != nil {
		return fmt.Errorf("既存のsync.tomlを退避できません: %w", err)
	}
	if err := os.Rename(source, destination); err != nil {
		restoreErr := os.Rename(backup, destination)
		if restoreErr != nil {
			return fmt.Errorf("sync.tomlの置換に失敗し、復元にも失敗しました: replace=%v restore=%w", err, restoreErr)
		}
		return fmt.Errorf("sync.tomlを置換できません: %w", err)
	}
	if err := os.Remove(backup); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("sync.tomlの退避ファイルを削除できません: %w", err)
	}
	return nil
}

func contains(items []string, expected string) bool {
	for _, item := range items {
		if item == expected {
			return true
		}
	}
	return false
}

func without(items []string, removed string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item != removed {
			result = append(result, item)
		}
	}
	return result
}
