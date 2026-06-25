// main.go はエントリポイント。os.Exit を呼ぶ唯一のファイル。
// ロジックは持たず、execute() → Cobra → internal パッケージへ全て委譲する。
package main

import (
	"fmt"
	"os"

	dotfiles "github.com/kokukaityo/dotfiles-bridge"
	engine "github.com/kokukaityo/dotfiles-bridge/internal"
)

func main() {
	engine.EngineVersion = dotfiles.Version
	if err := execute(dotfiles.TemplateFS, dotfiles.HookFS); err != nil {
		fmt.Fprintf(os.Stderr, "[dotfiles] Error: %v\n", err)
		os.Exit(1)
	}
}
