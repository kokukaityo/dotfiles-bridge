// sync.go は Git 同期（pull/push）、カテゴリ削除、.gitignore 生成、ステータス表示を担当する。
// このファイルが最も多くの機能を持ち、データリポジトリへの書き込み操作の大部分を担う。
package engine

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// gitignore のマーカー行。この行より上はユーザーの手書き、下は自動生成。
// GenerateGitignore はマーカー上を保持し、マーカー下だけを再生成する。
var (
	gitignoreMarkerStart = Setting.Gitignore.MarkerStart
	gitignoreMarkerEnd   = Setting.Gitignore.MarkerEnd
	securityPatterns     = Setting.Gitignore.SecurityPatterns
	conflictMarkerFile   = Setting.Path.ConflictMarkerFile
	syncConfigFile       = Setting.Path.SyncConfigFile
)

var categoryNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_.-]*$`)

// GenerateGitignore は .gitignore を再生成する。
// マーカー行より上のユーザー手書き部分を保持し、マーカー以下を
// sync.toml の ignore カテゴリやセキュリティ除外パターンから再生成する。
func GenerateGitignore(config *Config) error {
	path := RepositoryPath(config, ".gitignore")
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
	for _, pattern := range securityPatterns {
		output.WriteString(pattern + "\n")
	}
	output.WriteString("\n")
	output.WriteString("# Ignored categories (ignore in sync.toml)\n")
	for _, category := range config.Sync.Ignore {
		output.WriteString(category + "/\n")
	}
	output.WriteString("\n" + conflictMarkerFile + "\n" + hookDir + "/\n")
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

// Status は .conflict-pending マーカーの有無を確認し、未解決コンフリクトがあれば警告する。
// シェル起動時に dotfile status を呼ぶ運用を想定した軽量チェック。
func Status(config *Config, stdout io.Writer) error {
	if _, err := os.Stat(RepositoryPath(config, conflictMarkerFile)); err == nil {
		_, _ = fmt.Fprintln(stdout, "")
		_, _ = fmt.Fprintln(stdout, "========================================")
		_, _ = fmt.Fprintln(stdout, "  [dotfile] CONFLICT PENDING")
		_, _ = fmt.Fprintf(stdout, "  Run: cd %s && git log --oneline --graph --all\n", config.DotfilesDir)
		_, _ = fmt.Fprintln(stdout, "========================================")
		_, _ = fmt.Fprintln(stdout, "")
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("コンフリクト状態を確認できません: %w", err)
	}
	return nil
}

// Pull は fetch 後にローカルとリモートの関係を判定し、4通りに分岐する。
//  1. 最新: 何もしない
//  2. ローカルが遅れている: fast-forward merge
//  3. ローカルが先行: pull をスキップ（push を促す）
//  4. 分岐: ローカルを conflict/<host>/<timestamp> ブランチに退避し、
//     デフォルトブランチをリモートに合わせる。自動 merge はしない安全設計。
func Pull(config *Config, stdout, stderr io.Writer) error {
	git := GitRunner{WorkDir: config.DotfilesDir, Stdout: stdout, Stderr: stderr}
	if err := clearResolvedConflictMarker(config, git, stdout); err != nil {
		return err
	}
	if err := git.Run("fetch", "--quiet", "origin"); err != nil {
		return err
	}

	remoteRef := "origin/" + config.Sync.DefaultBranch
	remoteHead, err := git.Output("rev-parse", "--verify", remoteRef)
	if err != nil {
		_, _ = fmt.Fprintf(stdout, "[sync] リモートブランチがありません: %s\n", remoteRef)
		return nil
	}
	localHead, err := git.Output("rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if localHead == remoteHead {
		_, _ = fmt.Fprintln(stdout, "[sync] Already up to date.")
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
		_, _ = fmt.Fprintf(stdout, "[sync] Fast-forwarded to %s.\n", remoteRef)
		return nil
	}
	if remoteHead == mergeBase {
		_, _ = fmt.Fprintln(stdout, "[sync] ローカルがリモートより先行しています。pullをスキップします。")
		return nil
	}

	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}
	conflictBranch := fmt.Sprintf("conflict/%s/%s", sanitizeBranchPart(host), time.Now().Format("20060102-150405"))
	_, _ = fmt.Fprintf(stdout, "[sync] 分岐を検出しました。退避ブランチを作成: %s\n", conflictBranch)
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
	if err := os.WriteFile(RepositoryPath(config, conflictMarkerFile), nil, 0o644); err != nil {
		return fmt.Errorf("%sを作成できません: %w", conflictMarkerFile, err)
	}
	_, _ = fmt.Fprintf(stdout, "[sync] ローカル変更は%sへ退避し、%sを%sへ戻しました。\n", conflictBranch, config.Sync.DefaultBranch, remoteRef)
	return nil
}

// clearResolvedConflictMarker は Pull の冒頭で呼ばれる自動クリーンアップ。
// ユーザーが conflict ブランチを全て削除していれば、.conflict-pending マーカーも消す。
// ブランチがまだ残っていれば何もしない。
func clearResolvedConflictMarker(config *Config, git GitRunner, stdout io.Writer) error {
	marker := RepositoryPath(config, conflictMarkerFile)
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
			return fmt.Errorf("%sを削除できません: %w", conflictMarkerFile, err)
		}
		_, _ = fmt.Fprintln(stdout, "[sync] コンフリクト解消を確認し、マーカーを削除しました。")
	}
	return nil
}

// Push は auto カテゴリの変更だけを stage → commit → push する。
// デフォルトブランチ以外では実行しない安全弁付き。
// manual カテゴリや ignore カテゴリの変更は意図的に対象外。
func Push(config *Config, stdout, stderr io.Writer) error {
	git := GitRunner{WorkDir: config.DotfilesDir, Stdout: stdout, Stderr: stderr}
	currentBranch, err := git.Output("branch", "--show-current")
	if err != nil {
		return err
	}
	if currentBranch != config.Sync.DefaultBranch {
		_, _ = fmt.Fprintf(stdout, "[sync] %sブランチではありません (%s)。自動pushをスキップします。\n", config.Sync.DefaultBranch, currentBranch)
		return nil
	}

	var autoPaths []string
	var missing []string
	for _, category := range config.Sync.Auto {
		if info, statErr := os.Stat(RepositoryPath(config, category)); statErr == nil && info.IsDir() {
			autoPaths = append(autoPaths, category)
			continue
		}
		if trackedCategory(git, category) {
			missing = append(missing, category)
		}
	}
	if len(missing) > 0 {
		_, _ = fmt.Fprintln(stderr, "[sync] WARNING: 追跡済みの自動同期カテゴリが見つかりません:")
		for _, category := range missing {
			_, _ = fmt.Fprintf(stderr, "  - %s\n", category)
		}
		_, _ = fmt.Fprintln(stderr, "誤削除ならgit restore、恒久削除ならdotfile delete-categoryを使用してください。")
	}

	for _, category := range autoPaths {
		if err := git.Run("add", "--", category+"/"); err != nil {
			return err
		}
	}
	if len(autoPaths) == 0 {
		_, _ = fmt.Fprintln(stdout, "[sync] 自動同期カテゴリに変更はありません。")
		return nil
	}
	diffArgs := append([]string{"diff", "--cached", "--name-only", "--"}, autoPaths...)
	changed, err := git.Output(diffArgs...)
	if err != nil {
		return err
	}
	if changed == "" {
		_, _ = fmt.Fprintln(stdout, "[sync] 自動同期カテゴリに変更はありません。")
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
	_, _ = fmt.Fprintf(stdout, "[sync] Pushed: %s\n", message)
	return nil
}

// trackedCategory はカテゴリが Git に追跡されているかを2段階で確認する。
// ls-files（ワークツリー）と ls-tree（HEAD）の両方を見ることで、
// ディレクトリが消えていても追跡履歴があれば検出できる。
func trackedCategory(git GitRunner, category string) bool {
	if output, err := git.Output("ls-files", "--", category); err == nil && output != "" {
		return true
	}
	output, err := git.Output("ls-tree", "-r", "--name-only", "HEAD", "--", category)
	return err == nil && output != ""
}

// generateCommitMsg は staged の差分を add/update/delete に分類して
// "add: file1; update: file2" のような自動コミットメッセージを生成する。
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

// DeleteCategory は auto カテゴリを sync.toml・ファイルシステム・Git 履歴から一括削除する。
// sync.toml 更新 + カテゴリ削除 + push を1コミットにまとめるトランザクション的な処理。
// デフォルトブランチ以外や sync.toml に未コミット変更がある場合は拒否する。
func DeleteCategory(config *Config, category string, stdout, stderr io.Writer) error {
	if !categoryNamePattern.MatchString(category) || category == "." || category == ".." {
		return fmt.Errorf("不正なカテゴリ名です: %s", category)
	}
	git := GitRunner{WorkDir: config.DotfilesDir, Stdout: stdout, Stderr: stderr}
	currentBranch, err := git.Output("branch", "--show-current")
	if err != nil {
		return err
	}
	if currentBranch != config.Sync.DefaultBranch {
		return fmt.Errorf("カテゴリ削除は%sブランチでのみ実行できます (current: %s)", config.Sync.DefaultBranch, currentBranch)
	}
	if !slices.Contains(config.Sync.Auto, category) {
		return fmt.Errorf("自動同期カテゴリではありません: %s", category)
	}
	if !git.Success("diff", "--quiet", "--", syncConfigFile) ||
		!git.Success("diff", "--cached", "--quiet", "--", syncConfigFile) {
		return fmt.Errorf("%sに未コミットの変更があります", syncConfigFile)
	}

	hadTracked := trackedCategory(git, category)
	config.Sync.Auto = slices.DeleteFunc(config.Sync.Auto, func(s string) bool { return s == category })
	if err := writeSyncConfig(RepositoryPath(config, syncConfigFile), config.Sync); err != nil {
		return err
	}
	if err := git.Run("reset", "-q", "HEAD", "--", category); err != nil && hadTracked {
		return err
	}
	if err := os.RemoveAll(RepositoryPath(config, category)); err != nil {
		return fmt.Errorf("カテゴリを削除できません: %w", err)
	}
	if err := git.Run("add", "--", syncConfigFile); err != nil {
		return err
	}
	commitPaths := []string{syncConfigFile}
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
	_, _ = fmt.Fprintf(stdout, "[sync] カテゴリを削除してpushしました: %s\n", category)
	return nil
}

// writeSyncConfig は sync.toml をアトミックに書き換える。
// 一時ファイルに書いてから rename することで、書き込み途中のクラッシュでファイルが壊れるのを防ぐ。
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

