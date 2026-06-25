// main.go はエントリポイント。os.Exit を呼ぶ唯一のファイル。
// ロジックは持たず、execute() → Cobra → internal パッケージへ全て委譲する。
package main

import (
	"fmt"
	"os"

	dotfile "github.com/kokukaityo/dotfile"
	engine "github.com/kokukaityo/dotfile/internal"
)

func main() {
	engine.EngineVersion = dotfile.Version
	if err := execute(dotfile.TemplateFS, dotfile.HookFS); err != nil {
		fmt.Fprintf(os.Stderr, "[dotfile] Error: %v\n", err)
		os.Exit(1)
	}
}
