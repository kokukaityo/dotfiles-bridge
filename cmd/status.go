package main

import (
	engine "github.com/kokukaityo/dotfiles-bridge/internal"
	"github.com/spf13/cobra"
)

// statusCommand は未解決コンフリクトの有無を表示する。
// シェルの起動スクリプトに組み込んで、毎回チェックする運用を想定。
func (a *application) statusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "コンフリクト状態を表示",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config, err := a.config()
			if err != nil {
				return err
			}
			return engine.Status(config, cmd.OutOrStdout())
		},
	}
}

// deleteCategoryCommand はカテゴリをまるごと削除する。
// sync.toml の更新・ファイル削除・Git 履歴の整理・push までを1コミットで行う。
func (a *application) deleteCategoryCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-category <name>",
		Short: "カテゴリを削除",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := a.config()
			if err != nil {
				return err
			}
			return engine.DeleteCategory(config, args[0], cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
}

// gitignoreCommand は .gitignore の自動生成部分だけを再生成する。
// sync.toml の ignore リストを変更した後に使う。手書き部分は保持される。
func (a *application) gitignoreCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "gitignore",
		Short: ".gitignoreを再生成",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config, err := a.config()
			if err != nil {
				return err
			}
			return engine.GenerateGitignore(config)
		},
	}
}
