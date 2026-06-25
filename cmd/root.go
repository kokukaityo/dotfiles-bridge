// root.go は Cobra のルートコマンド定義と、全サブコマンドで共有する application 構造体を持つ。
package main

import (
	"io/fs"

	engine "github.com/kokukaityo/dotfiles-bridge/internal"
	"github.com/spf13/cobra"
)

// application は embed.go で宣言した埋め込みリソースを保持し、各サブコマンドへ渡す中継役。
// execute() で生成され、全てのコマンド定義メソッドのレシーバになる。
type application struct {
	templateFS fs.FS
	hookFS     fs.FS
}

func execute(templateFS fs.FS, hookFS fs.FS) error {
	app := &application{
		templateFS: templateFS,
		hookFS:     hookFS,
	}
	return app.rootCommand().Execute()
}

func (a *application) rootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "dotfiles",
		Short:         "dotfiles管理ツール",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.AddCommand(
		a.initCommand(),
		a.installCommand(),
		a.linkCommand(),
		a.pullCommand(),
		a.pushCommand(),
		a.watchCommand(),
		a.deleteCategoryCommand(),
		a.gitignoreCommand(),
		a.statusCommand(),
		a.versionCommand(),
	)
	return root
}

// config は init・version 以外の全サブコマンドが最初に呼ぶ共通処理。
func (a *application) config() (*engine.Config, error) {
	return engine.Resolve()
}
