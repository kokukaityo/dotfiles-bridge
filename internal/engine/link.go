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

type LinkConfig map[string]map[string][]string

func loadLinkConfig(path string) (LinkConfig, error) {
	var config LinkConfig
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, fmt.Errorf("%sを読み込めません: %w", path, err)
	}
	return config, nil
}

func linkAll(config *Config, stdout io.Writer) error {
	osKey, err := OSKey()
	if err != nil {
		return err
	}

	matches, err := filepath.Glob(filepath.Join(config.DotfilesDir, "*", "link.toml"))
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
			for _, target := range entries[source] {
				targetPath, err := ExpandHome(strings.TrimSuffix(target, "/"))
				if err != nil {
					return err
				}
				if err := createLink(sourcePath, targetPath, stdout); err != nil {
					return err
				}
			}
		}
	}
	_, _ = fmt.Fprintln(stdout, "Done.")
	return nil
}

func createLink(source, target string, stdout io.Writer) error {
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

		backup := target + ".bak." + time.Now().Format("20060102150405")
		if err := os.Rename(target, backup); err != nil {
			return fmt.Errorf("既存ファイルをバックアップできません (%s): %w", target, err)
		}
		_, _ = fmt.Fprintf(stdout, "  backed up: %s -> %s\n", target, backup)
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
