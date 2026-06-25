package main

import (
	"fmt"

	engine "github.com/kokukaityo/dotfiles-bridge/internal"
	"github.com/spf13/cobra"
)

// versionCommand は本体とデータリポジトリの情報を表示する。
// Resolve のエラーは握り潰す。データリポジトリがなくても本体バージョンだけは表示したいため。
func (a *application) versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "バージョン情報を表示",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "dotfiles v%s\n", engine.EngineVersion)
			config, err := engine.Resolve()
			if err != nil {
				return nil
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  data: %s\n", config.DotfilesDir)
			return nil
		},
	}
}
