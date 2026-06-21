package main

import (
	engine "github.com/kokukaityo/dotfile/internal"
	"github.com/spf13/cobra"
)

// pullCommand はリモートの変更をローカルに反映する。
// コンフリクト時は自動 merge せず退避ブランチに逃がす安全設計。
func (a *application) pullCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "pull",
		Short: "リモートから同期",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config, err := a.config()
			if err != nil {
				return err
			}
			return engine.Pull(config, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
}

// pushCommand は auto カテゴリの変更だけを commit して push する。
// manual カテゴリの変更は含めない。pull とは非対称な設計。
func (a *application) pushCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Short: "変更をcommitしてpush",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config, err := a.config()
			if err != nil {
				return err
			}
			return engine.Push(config, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
}
