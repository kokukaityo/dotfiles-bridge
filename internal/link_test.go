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

func TestBackupSubPath(t *testing.T) {
	tests := []struct {
		name    string
		targets []string
		want    map[string]string
	}{
		{
			name:    "no collision",
			targets: []string{"/home/.bashrc", "/home/.config/Code/User/settings.json"},
			want: map[string]string{
				"/home/.bashrc":                         ".bashrc",
				"/home/.config/Code/User/settings.json": "settings.json",
			},
		},
		{
			name: "one level diff",
			targets: []string{
				"/home/.config/Code/User/settings.json",
				"/home/.config/Cursor/User/settings.json",
			},
			want: map[string]string{
				"/home/.config/Code/User/settings.json":   filepath.Join("Code", "settings.json"),
				"/home/.config/Cursor/User/settings.json": filepath.Join("Cursor", "settings.json"),
			},
		},
		{
			name: "multi level diff",
			targets: []string{
				"/home/.config/Code/User/settings.json",
				"/home/.config/Code/Backup/settings.json",
				"/home/.config/Cursor/User/settings.json",
			},
			want: map[string]string{
				"/home/.config/Code/User/settings.json":   filepath.Join("Code", "User", "settings.json"),
				"/home/.config/Code/Backup/settings.json": filepath.Join("Code", "Backup", "settings.json"),
				"/home/.config/Cursor/User/settings.json": filepath.Join("Cursor", "User", "settings.json"),
			},
		},
		{
			name:    "single target",
			targets: []string{"/home/.config/Code/User/settings.json"},
			want: map[string]string{
				"/home/.config/Code/User/settings.json": "settings.json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := backupSubPath(tt.targets)
			for target, wantSub := range tt.want {
				if gotSub, ok := got[target]; !ok {
					t.Errorf("missing key %q", target)
				} else if gotSub != wantSub {
					t.Errorf("backupSubPath[%q] = %q, want %q", target, gotSub, wantSub)
				}
			}
		})
	}
}

func TestCreateLinkBackup(t *testing.T) {
	if !canSymlink(t) {
		t.Skip("symlink creation not available")
	}
	dir := t.TempDir()
	backupPath := filepath.Join(dir, ".backup", "testcat_20260621120000", "target.txt")
	source := filepath.Join(dir, "source.txt")
	target := filepath.Join(dir, "target.txt")

	writeTestFile(t, source, "new content")
	writeTestFile(t, target, "old content")

	var buf bytes.Buffer
	if err := createLink(source, target, backupPath, &buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, "backed up:") {
		t.Fatalf("expected backup message, got: %s", output)
	}
	if !strings.Contains(output, "linked:") {
		t.Fatalf("expected linked message, got: %s", output)
	}

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
	backupPath := filepath.Join(dir, ".backup", "testcat_20260621120000", "targetdir")
	source := filepath.Join(dir, "srcdir")
	target := filepath.Join(dir, "targetdir")

	writeTestFile(t, filepath.Join(source, "a.txt"), "source a")
	writeTestFile(t, filepath.Join(target, "b.txt"), "old b")

	var buf bytes.Buffer
	if err := createLink(source, target, backupPath, &buf); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(backupPath, "b.txt"))
	if err != nil {
		t.Fatalf("backup dir content not found: %v", err)
	}
	if string(data) != "old b" {
		t.Fatalf("backup content = %q, want %q", string(data), "old b")
	}
}

func TestCreateLinkBackupNameCollision(t *testing.T) {
	if !canSymlink(t) {
		t.Skip("symlink creation not available")
	}
	dir := t.TempDir()
	categoryBackupDir := filepath.Join(dir, ".backup", "editor_20260623120000")
	source := filepath.Join(dir, "settings.json")

	target1 := filepath.Join(dir, "home", ".config", "Code", "User", "settings.json")
	target2 := filepath.Join(dir, "home", ".config", "Cursor", "User", "settings.json")

	writeTestFile(t, source, "new content")
	writeTestFile(t, target1, "vscode content")
	writeTestFile(t, target2, "cursor content")

	subPaths := backupSubPath([]string{target1, target2})

	var buf bytes.Buffer
	bk1 := filepath.Join(categoryBackupDir, subPaths[target1])
	if err := createLink(source, target1, bk1, &buf); err != nil {
		t.Fatal(err)
	}
	bk2 := filepath.Join(categoryBackupDir, subPaths[target2])
	if err := createLink(source, target2, bk2, &buf); err != nil {
		t.Fatal(err)
	}

	data1, err := os.ReadFile(bk1)
	if err != nil {
		t.Fatalf("backup1 not found: %v", err)
	}
	if string(data1) != "vscode content" {
		t.Fatalf("backup1 = %q, want %q", string(data1), "vscode content")
	}

	data2, err := os.ReadFile(bk2)
	if err != nil {
		t.Fatalf("backup2 not found: %v", err)
	}
	if string(data2) != "cursor content" {
		t.Fatalf("backup2 = %q, want %q", string(data2), "cursor content")
	}
}
