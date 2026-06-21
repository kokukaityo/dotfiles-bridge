package main

import (
	engine "github.com/kokukaityo/dotfile/internal"
	"github.com/spf13/cobra"
)

// setupCommand は clone 済みの既存データリポジトリに hooks・gitignore・symlink を適用する。
// init が「新規作成」なのに対し、setup は「別マシンで既存リポジトリを使い始める」ときに使う。
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
			return engine.SetupRepository(config, a.hookFS, cmd.OutOrStdout())
		},
	}
}
