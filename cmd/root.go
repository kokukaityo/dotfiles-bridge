// root.go は Cobra のルートコマンド定義と、全サブコマンドで共有する application 構造体を持つ。
package main

import (
	"fmt"
	"io/fs"
	"os"

	engine "github.com/kokukaityo/dotfile/internal"
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
		Use:           "dotfile",
		Short:         "dotfiles同期エンジン",
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
// データリポジトリの解決とバージョン不整合の警告を集約している。
func (a *application) config() (*engine.Config, error) {
	config, err := engine.Resolve()
	if err != nil {
		return nil, err
	}
	if config.VersionMismatch() {
		_, _ = fmt.Fprintf(os.Stderr, "[dotfile] WARNING: バージョン不整合 (engine=%s, data=%s)\n", config.EngineVersion, config.DataVersion)
	}
	return config, nil
}
