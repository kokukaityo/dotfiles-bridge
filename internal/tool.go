package engine

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// ExpandPath はパス先頭の ~ をホームディレクトリに展開する。
func ExpandPath(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") && !strings.HasPrefix(path, `~\`) {
		return filepath.Clean(path), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("ホームディレクトリを取得できません: %w", err)
	}
	if path == "~" {
		return home, nil
	}
	return filepath.Join(home, path[2:]), nil
}

// ensureLine はファイルに指定行がなければ追記する冪等ヘルパー。
func ensureLine(path, expected string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return fmt.Errorf("%sを開けません: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == expected {
			return nil
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%sを読み込めません: %w", path, err)
	}
	info, err := file.Stat()
	if err != nil {
		return err
	}
	prefix := ""
	if info.Size() > 0 {
		prefix = "\n"
	}
	if _, err := file.WriteString(prefix + expected + "\n"); err != nil {
		return fmt.Errorf("%sを更新できません: %w", path, err)
	}
	return nil
}

// replaceFile は OS ごとのアトミックなファイル置換。
// Windows は既存ファイルがあると rename が失敗するため、退避→置換→退避削除のフォールバックを行う。
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

// sanitizeBranchPart は外部由来の文字列を Git ブランチ名に使える形にする。
func sanitizeBranchPart(value string) string {
	replacer := strings.NewReplacer(" ", "-", "~", "-", "^", "-", ":", "-", "?", "-", "*", "-", "[", "-", "\\", "-", "..", "-")
	return strings.Trim(replacer.Replace(value), "./")
}

// uniqueBaseNames は改行区切りのパス一覧からユニークなファイル名を抽出する。
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
