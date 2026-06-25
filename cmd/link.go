package main

import (
	engine "github.com/kokukaityo/dotfiles-bridge/internal"
	"github.com/spf13/cobra"
)

// linkCommand は symlink の配置だけを単独で実行する。
// install にも含まれるが、link.toml を編集した後にリンクだけ貼り直したいときに使う。
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
			return engine.LinkAll(config, cmd.OutOrStdout())
		},
	}
}
