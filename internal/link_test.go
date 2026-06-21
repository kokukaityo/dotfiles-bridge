package engine

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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

func canSymlink(t *testing.T) bool {
	t.Helper()
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	link := filepath.Join(dir, "link")
	writeTestFile(t, target, "")
	return os.Symlink(target, link) == nil
}

func TestCreateLinkBackup(t *testing.T) {
	if !canSymlink(t) {
		t.Skip("symlink creation not available")
	}
	dir := t.TempDir()
	backupDir := filepath.Join(dir, ".backup", "testcat", "20260621120000")
	source := filepath.Join(dir, "source.txt")
	target := filepath.Join(dir, "target.txt")

	writeTestFile(t, source, "new content")
	writeTestFile(t, target, "old content")

	var buf bytes.Buffer
	if err := createLink(source, target, backupDir, &buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, "backed up:") {
		t.Fatalf("expected backup message, got: %s", output)
	}
	if !strings.Contains(output, "linked:") {
		t.Fatalf("expected linked message, got: %s", output)
	}

	backupPath := filepath.Join(backupDir, "target.txt")
	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("backup file not found: %v", err)
	}
	if string(data) != "old content" {
		t.Fatalf("backup content = %q, want %q", string(data), "old content")
	}
}

func TestCreateLinkBackupDirectory(t *testing.T) {
	if !canSymlink(t) {
		t.Skip("symlink creation not available")
	}
	dir := t.TempDir()
	backupDir := filepath.Join(dir, ".backup", "testcat", "20260621120000")
	source := filepath.Join(dir, "srcdir")
	target := filepath.Join(dir, "targetdir")

	writeTestFile(t, filepath.Join(source, "a.txt"), "source a")
	writeTestFile(t, filepath.Join(target, "b.txt"), "old b")

	var buf bytes.Buffer
	if err := createLink(source, target, backupDir, &buf); err != nil {
		t.Fatal(err)
	}

	backedUpDir := filepath.Join(backupDir, "targetdir")
	data, err := os.ReadFile(filepath.Join(backedUpDir, "b.txt"))
	if err != nil {
		t.Fatalf("backup dir content not found: %v", err)
	}
	if string(data) != "old b" {
		t.Fatalf("backup content = %q, want %q", string(data), "old b")
	}
}
