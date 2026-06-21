package main

import (
	"fmt"
	"os"

	dotfile "github.com/kokukaityo/dotfile"
	"github.com/kokukaityo/dotfile/internal/engine"
)

func main() {
	if err := engine.Execute(dotfile.TemplateFS, dotfile.Version, dotfile.HookFS); err != nil {
		fmt.Fprintf(os.Stderr, "[dotfile] Error: %v\n", err)
		os.Exit(1)
	}
}
