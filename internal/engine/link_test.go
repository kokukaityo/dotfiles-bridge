package engine

import (
	"path/filepath"
	"testing"
)

func TestLoadLinkConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "link.toml")
	writeTestFile(t, path, `
[darwin]
"AGENTS.md" = ["~/.codex/AGENTS.md", "~/.claude/CLAUDE.md"]
`)

	config, err := loadLinkConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	targets := config["darwin"]["AGENTS.md"]
	if len(targets) != 2 || targets[0] != "~/.codex/AGENTS.md" {
		t.Fatalf("unexpected targets: %#v", targets)
	}
}
