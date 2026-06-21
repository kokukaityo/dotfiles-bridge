package engine

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type application struct {
	templateFS    fs.FS
	hookFS        fs.FS
	engineVersion string
}

func (a *application) initCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init [path]",
		Short: "データリポジトリを新規作成",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "~/dotfiles"
			if len(args) == 1 {
				target = args[0]
			}
			return initializeRepository(target, a, cmd.OutOrStdout())
		},
	}
}

func (a *application) setupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "データリポジトリを初期設定",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config, err := a.config()
			if err != nil {
				return err
			}
			return setupRepository(config, a.hookFS, cmd.OutOrStdout())
		},
	}
}

func (a *application) linkCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "link",
		Short: "symlinkを配置",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config, err := a.config()
			if err != nil {
				return err
			}
			return linkAll(config, cmd.OutOrStdout())
		},
	}
}

func (a *application) pullCommand() *cobra.Command {
	return configCommand(a, "pull", "リモートから同期", pull)
}

func (a *application) pushCommand() *cobra.Command {
	return configCommand(a, "push", "変更をcommitしてpush", push)
}

func (a *application) gitignoreCommand() *cobra.Command {
	return configCommand(a, "gitignore", ".gitignoreを再生成", func(config *Config, _ *cobra.Command) error {
		return generateGitignore(config)
	})
}

func (a *application) statusCommand() *cobra.Command {
	return configCommand(a, "status", "コンフリクト状態を表示", status)
}

func (a *application) deleteCategoryCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-category <name>",
		Short: "自動同期カテゴリを削除",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := a.config()
			if err != nil {
				return err
			}
			return deleteCategory(config, args[0], cmd)
		},
	}
}

func configCommand(
	app *application,
	use string,
	short string,
	run func(*Config, *cobra.Command) error,
) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config, err := app.config()
			if err != nil {
				return err
			}
			return run(config, cmd)
		},
	}
}

func Execute(templateFS fs.FS, engineVersion string, hookFS fs.FS) error {
	app := &application{
		templateFS:    templateFS,
		hookFS:        hookFS,
		engineVersion: strings.TrimSpace(engineVersion),
	}
	return app.rootCommand().Execute()
}

func (a *application) rootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "dotfile",
		Short:         "dotfiles同期エンジン",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.AddCommand(
		a.initCommand(),
		a.setupCommand(),
		a.linkCommand(),
		a.pullCommand(),
		a.pushCommand(),
		a.deleteCategoryCommand(),
		a.gitignoreCommand(),
		a.statusCommand(),
		a.versionCommand(),
	)
	return root
}

func (a *application) versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "バージョン情報を表示",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "dotfile engine v%s\n", a.engineVersion)
			config, err := Resolve(a.engineVersion)
			if err != nil {
				return nil
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  data:         %s\n", config.DotfilesDir)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  data version: %s\n", config.DataVersion)
			if config.VersionMismatch() {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[dotfile] WARNING: エンジンとデータのメジャーバージョンが異なります\n")
			}
			return nil
		},
	}
}

func (a *application) config() (*Config, error) {
	config, err := Resolve(a.engineVersion)
	if err != nil {
		return nil, err
	}
	if config.VersionMismatch() {
		_, _ = fmt.Fprintf(os.Stderr, "[dotfile] WARNING: バージョン不整合 (engine=%s, data=%s)\n", config.EngineVersion, config.DataVersion)
	}
	return config, nil
}

func repositoryPath(config *Config, names ...string) string {
	parts := append([]string{config.DotfilesDir}, names...)
	return filepath.Join(parts...)
}
