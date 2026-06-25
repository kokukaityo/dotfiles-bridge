package main

import (
	"fmt"

	engine "github.com/kokukaityo/dotfiles-bridge/internal"
	"github.com/spf13/cobra"
)

// versionCommand はエンジンとデータリポジトリのバージョンを表示する。
// Resolve のエラーは握り潰す。データリポジトリがなくてもエンジンバージョンだけは表示したいため。
func (a *application) versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "バージョン情報を表示",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "dotfiles engine v%s\n", engine.EngineVersion)
			config, err := engine.Resolve()
			if err != nil {
				return nil
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  data:         %s\n", config.DotfilesDir)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  data version: %s\n", config.DataVersion)
			if config.VersionMismatch() {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[dotfiles] WARNING: エンジンとデータのメジャーバージョンが異なります\n")
			}
			return nil
		},
	}
}
