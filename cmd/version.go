package main

import (
	"fmt"

	engine "github.com/kokukaityo/dotfile/internal"
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
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "dotfile engine v%s\n", a.engineVersion)
			config, err := engine.Resolve(a.engineVersion)
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
