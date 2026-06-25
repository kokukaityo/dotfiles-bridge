// link.go は link.toml に基づく symlink の配置を担当する。
// installCommand (install) と linkCommand (link) の両方から呼ばれる。
// ユーザーのファイルを直接操作するため、既存ファイルのバックアップと
// 同一リンク済みのスキップで安全性を担保する。
package engine

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// LinkConfig は link.toml の構造: OSキー → ソースファイル名 → ターゲットパスのリスト。
// 1つのソースに複数ターゲットを指定できる（例: 同じ settings.json を VS Code と Cursor に配置）。
var (
	linkConfigFile = Setting.Path.LinkConfigFile
	backupDir      = Setting.Path.BackupDir
)

type LinkConfig map[string]map[string][]string

func loadLinkConfig(path string) (LinkConfig, error) {
	var config LinkConfig
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, fmt.Errorf("%sを読み込めません: %w", path, err)
	}
	return config, nil
}

// LinkAll は全カテゴリの link.toml を走査し、現在の OS に該当するセクションだけ処理する。
// 他 OS のセクションは無視されるので、1つの link.toml に全 OS 分を書ける。
// backupSubPath は同一ソースの複数ターゲットでバックアップが衝突しないよう、
// 最小限の親ディレクトリ名を付加したサブパスを返す。
func backupSubPath(targets []string) map[string]string {
	result := make(map[string]string, len(targets))

	groups := map[string][]string{}
	for _, t := range targets {
		base := filepath.Base(t)
		groups[base] = append(groups[base], t)
	}

	for base, paths := range groups {
		if len(paths) == 1 {
			result[paths[0]] = base
			continue
		}

		components := make([][]string, len(paths))
		for i, p := range paths {
			dir := filepath.Dir(p)
			var parts []string
			for {
				parent := filepath.Dir(dir)
				if parent == dir {
					break
				}
				parts = append(parts, filepath.Base(dir))
				dir = parent
			}
			components[i] = parts
		}

		maxDepth := 0
		for _, c := range components {
			if len(c) > maxDepth {
				maxDepth = len(c)
			}
		}

		found := false
		for depth := 0; depth < maxDepth; depth++ {
			seen := map[string]bool{}
			allUnique := true
			for i := range paths {
				var key string
				if depth < len(components[i]) {
					key = components[i][depth]
				} else {
					key = paths[i]
				}
				if seen[key] {
					allUnique = false
					break
				}
				seen[key] = true
			}
			if allUnique {
				for i, p := range paths {
					if depth < len(components[i]) {
						result[p] = filepath.Join(components[i][depth], base)
					} else {
						result[p] = filepath.Join(paths[i], base)
					}
				}
				found = true
				break
			}
		}

		if !found {
			suffixes := make([]string, len(paths))
			for i := range suffixes {
				suffixes[i] = base
			}
			for depth := 0; depth < maxDepth; depth++ {
				for i := range paths {
					if depth < len(components[i]) {
						suffixes[i] = filepath.Join(components[i][depth], suffixes[i])
					}
				}
				seen := map[string]bool{}
				allUnique := true
				for _, s := range suffixes {
					if seen[s] {
						allUnique = false
						break
					}
					seen[s] = true
				}
				if allUnique {
					for i, p := range paths {
						result[p] = suffixes[i]
					}
					found = true
					break
				}
			}
		}

	}

	return result
}

func LinkAll(config *Config, stdout io.Writer) error {
	osKey, err := OSKey()
	if err != nil {
		return err
	}

	matches, err := filepath.Glob(filepath.Join(config.DotfilesDir, "*", linkConfigFile))
	if err != nil {
		return fmt.Errorf("link.tomlの検索に失敗しました: %w", err)
	}
	sort.Strings(matches)

	_, _ = fmt.Fprintf(stdout, "=== dotfile link (%s) ===\n", osKey)
	for _, configPath := range matches {
		categoryDir := filepath.Dir(configPath)
		categoryName := filepath.Base(categoryDir)
		linkConfig, err := loadLinkConfig(configPath)
		if err != nil {
			return err
		}
		entries := linkConfig[osKey]
		if len(entries) == 0 {
			continue
		}

		_, _ = fmt.Fprintf(stdout, "[%s]\n", categoryName)
		timestamp := time.Now().Format("20060102150405")
		categoryBackupDir := RepositoryPath(config, backupDir, categoryName+"_"+timestamp)
		sources := make([]string, 0, len(entries))
		for source := range entries {
			sources = append(sources, source)
		}
		sort.Strings(sources)
		for _, source := range sources {
			sourcePath := filepath.Join(categoryDir, filepath.FromSlash(strings.TrimSuffix(source, "/")))
			if _, err := os.Stat(sourcePath); err != nil {
				if os.IsNotExist(err) {
					_, _ = fmt.Fprintf(stdout, "  skip (source not found): %s\n", sourcePath)
					continue
				}
				return fmt.Errorf("リンク元を確認できません: %w", err)
			}
			sourcePath, err = filepath.Abs(sourcePath)
			if err != nil {
				return fmt.Errorf("リンク元を解決できません: %w", err)
			}

			targetPaths := make([]string, 0, len(entries[source]))
			seen := map[string]bool{}
			for _, target := range entries[source] {
				p, err := ExpandPath(strings.TrimSuffix(target, "/"))
				if err != nil {
					return err
				}
				if seen[p] {
					return fmt.Errorf("ターゲットパスが重複しています: %s", p)
				}
				seen[p] = true
				targetPaths = append(targetPaths, p)
			}
			subPaths := backupSubPath(targetPaths)
			for _, targetPath := range targetPaths {
				bkPath := filepath.Join(categoryBackupDir, subPaths[targetPath])
				if err := createLink(sourcePath, targetPath, bkPath, stdout); err != nil {
					return err
				}
			}
		}
	}
	_, _ = fmt.Fprintln(stdout, "Done.")
	return nil
}

func createLink(source, target, backupPath string, stdout io.Writer) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("リンク先ディレクトリを作成できません: %w", err)
	}

	if info, err := os.Lstat(target); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			current, readErr := os.Readlink(target)
			if readErr == nil && filepath.Clean(current) == filepath.Clean(source) {
				_, _ = fmt.Fprintf(stdout, "  ok (already linked): %s\n", target)
				return nil
			}
		}

		if err := os.MkdirAll(filepath.Dir(backupPath), 0o755); err != nil {
			return fmt.Errorf("バックアップディレクトリを作成できません: %w", err)
		}
		if err := os.Rename(target, backupPath); err != nil {
			return fmt.Errorf("既存ファイルをバックアップできません (%s): %w", target, err)
		}
		_, _ = fmt.Fprintf(stdout, "  backed up: %s -> %s\n", target, backupPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("リンク先を確認できません (%s): %w", target, err)
	}

	if err := os.Symlink(source, target); err != nil {
		if osKey, _ := OSKey(); osKey == "win32" {
			return fmt.Errorf("symlinkを作成できません。Windowsの開発者モードを有効化してください (%s): %w", target, err)
		}
		return fmt.Errorf("symlinkを作成できません (%s): %w", target, err)
	}
	_, _ = fmt.Fprintf(stdout, "  linked: %s -> %s\n", target, source)
	return nil
}
