// setup.go はデータリポジトリの初期化（init）と既存リポジトリへの設定適用（setup）を担当する。
// init: テンプレート展開 → git init → SetupRepository → 初回コミット の一連フロー。
// setup: clone 済みリポジトリに hooks・gitattributes・gitignore・symlink を適用するフロー。
package engine

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const hookFileMode fs.FileMode = 0o755

// InitializeRepository は dotfile init のフロー全体を実行する。
// テンプレート展開 → git init → SetupRepository → add + commit まで一括で行う。
// 対象パスが既に存在する場合はエラーにして上書きを防ぐ。
func InitializeRepository(target string, templateFS fs.FS, engineVersion string, hookFS fs.FS, stdout io.Writer) error {
	target, err := ExpandPath(target)
	if err != nil {
		return err
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("初期化先を解決できません: %w", err)
	}
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("既にパスが存在します: %s", target)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("初期化先を確認できません: %w", err)
	}

	_, _ = fmt.Fprintf(stdout, "[dotfile] データリポジトリを作成: %s\n", target)
	if err := extractTemplate(templateFS, target); err != nil {
		return err
	}

	config, err := loadConfig(target, engineVersion)
	if err != nil {
		return err
	}
	git := GitRunner{WorkDir: target, Stdout: stdout}
	if err := git.Run("init", "-b", config.Sync.DefaultBranch); err != nil {
		return err
	}
	if err := SetupRepository(config, hookFS, stdout); err != nil {
		return err
	}
	if err := git.Run("add", "-A"); err != nil {
		return err
	}
	if err := git.Run("commit", "-m", "feat: initial dotfiles setup"); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(stdout, "[dotfile] データリポジトリの作成が完了しました")
	return nil
}

// extractTemplate は埋め込みテンプレート（embed.go の TemplateFS）をディスクに展開する。
// template/ 配下のディレクトリ構造をそのまま再現する。
func extractTemplate(templateFS fs.FS, target string) error {
	return fs.WalkDir(templateFS, Setting.Path.TemplateDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(Setting.Path.TemplateDir, path)
		if err != nil {
			return fmt.Errorf("テンプレートパスを解決できません: %w", err)
		}
		destination := filepath.Join(target, filepath.FromSlash(relative))
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}
		data, err := fs.ReadFile(templateFS, path)
		if err != nil {
			return fmt.Errorf("テンプレートを読み込めません (%s): %w", path, err)
		}
		if err := os.WriteFile(destination, data, 0o644); err != nil {
			return fmt.Errorf("テンプレートを書き出せません (%s): %w", destination, err)
		}
		return nil
	})
}

// SetupRepository は dotfile setup のフロー。InitializeRepository からも呼ばれる共通処理。
// hooks 展開 → core.hooksPath 設定 → gitattributes → gitignore 生成 → symlink 配置 の順。
func SetupRepository(config *Config, hookFS fs.FS, stdout io.Writer) error {
	git := GitRunner{WorkDir: config.DotfilesDir, Stdout: stdout}
	if !git.Success("rev-parse", "--git-dir") {
		return fmt.Errorf("gitリポジトリではありません: %s", config.DotfilesDir)
	}
	if err := installHooks(config.DotfilesDir, hookFS); err != nil {
		return err
	}
	if err := git.Run("config", Setting.Git.HooksPath, Setting.Path.HookDir); err != nil {
		return err
	}
	if err := ensureLine(filepath.Join(config.DotfilesDir, ".gitattributes"), Setting.Git.GitattributesLine); err != nil {
		return err
	}
	if err := git.Run("config", Setting.Git.Symlinks, "true"); err != nil {
		return err
	}
	if err := GenerateGitignore(config); err != nil {
		return err
	}
	if err := LinkAll(config, stdout); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(stdout, "[dotfile] Setup complete.")
	return nil
}

// installHooks は埋め込みの hook スクリプトを .dotfile-hook/ に書き出す。
// .git/hooks/ ではなく core.hooksPath で参照させることで、
// データリポジトリ側に hook を置きつつ Git 追跡対象外にできる。
func installHooks(dotfilesDir string, hookFS fs.FS) error {
	hookDir := filepath.Join(dotfilesDir, Setting.Path.HookDir)
	if err := os.MkdirAll(hookDir, 0o755); err != nil {
		return fmt.Errorf("hookディレクトリを作成できません: %w", err)
	}
	for _, source := range Setting.Hook.Sources {
		data, err := fs.ReadFile(hookFS, source)
		if err != nil {
			return fmt.Errorf("hookを読み込めません (%s): %w", source, err)
		}
		target := filepath.Join(hookDir, filepath.Base(source))
		if err := os.WriteFile(target, data, hookFileMode); err != nil {
			return fmt.Errorf("hookを書き出せません (%s): %w", target, err)
		}
		if err := os.Chmod(target, hookFileMode); err != nil {
			return fmt.Errorf("hookの権限を設定できません (%s): %w", target, err)
		}
	}
	return nil
}
