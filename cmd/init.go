package main

import (
	engine "github.com/kokukaityo/dotfile/internal"
	"github.com/spf13/cobra"
)

// initCommand はデータリポジトリをゼロから新規作成する。
// テンプレート展開 → git init → setup → 初回コミットまで一括実行。
// 既存リポジトリへの適用は setupCommand の役割。
func (a *application) initCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init [path]",
		Short: "データリポジトリを新規作成",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := engine.DefaultDir
			if len(args) == 1 {
				target = args[0]
			}
			return engine.InitializeRepository(target, a.templateFS, a.hookFS, cmd.OutOrStdout())
		},
	}
}
