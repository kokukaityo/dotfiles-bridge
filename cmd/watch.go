package main

import (
	engine "github.com/kokukaityo/dotfile/internal"
	"github.com/spf13/cobra"
)

func (a *application) watchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "watch",
		Short: "ファイル変更を監視して自動push",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config, err := a.config()
			if err != nil {
				return err
			}
			return engine.Watch(config, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
}
