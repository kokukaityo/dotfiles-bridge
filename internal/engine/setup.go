package engine

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const hookFileMode fs.FileMode = 0o755

func initializeRepository(target string, app *application, stdout io.Writer) error {
	target, err := ExpandHome(target)
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
	if err := extractTemplate(app.templateFS, target); err != nil {
		return err
	}

	config, err := loadConfig(target, app.engineVersion)
	if err != nil {
		return err
	}
	git := GitRunner{WorkDir: target, Stdout: stdout}
	if err := git.Run("init", "-b", config.Sync.DefaultBranch); err != nil {
		return err
	}
	if err := setupRepository(config, app.hookFS, stdout); err != nil {
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

func extractTemplate(templateFS fs.FS, target string) error {
	return fs.WalkDir(templateFS, "template", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel("template", path)
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

func setupRepository(config *Config, hookFS fs.FS, stdout io.Writer) error {
	git := GitRunner{WorkDir: config.DotfilesDir, Stdout: stdout}
	if !git.Success("rev-parse", "--git-dir") {
		return fmt.Errorf("gitリポジトリではありません: %s", config.DotfilesDir)
	}
	if err := installHooks(config.DotfilesDir, hookFS); err != nil {
		return err
	}
	if err := git.Run("config", "core.hooksPath", ".dotfile-hook"); err != nil {
		return err
	}
	if err := ensureLine(filepath.Join(config.DotfilesDir, ".gitattributes"), "* -text"); err != nil {
		return err
	}
	if err := git.Run("config", "core.symlinks", "true"); err != nil {
		return err
	}
	if err := generateGitignore(config); err != nil {
		return err
	}
	if err := linkAll(config, stdout); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(stdout, "[dotfile] Setup complete.")
	return nil
}

func installHooks(dotfilesDir string, hookFS fs.FS) error {
	hookDir := filepath.Join(dotfilesDir, ".dotfile-hook")
	if err := os.MkdirAll(hookDir, 0o755); err != nil {
		return fmt.Errorf("hookディレクトリを作成できません: %w", err)
	}
	for _, source := range []string{"lib/hooks/pre-push", "lib/hooks/post-merge"} {
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
