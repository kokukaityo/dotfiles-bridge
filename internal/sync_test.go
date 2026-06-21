package engine

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestGenerateCommitMsg(t *testing.T) {
	dir := newTestRepository(t, "main")
	writeTestFile(t, filepath.Join(dir, "editor", "updated.txt"), "before")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-m", "initial")

	writeTestFile(t, filepath.Join(dir, "editor", "updated.txt"), "after")
	writeTestFile(t, filepath.Join(dir, "editor", "new.txt"), "new")
	runGit(t, dir, "add", "--", "editor")

	message, err := generateCommitMsg(GitRunner{WorkDir: dir}, []string{"editor"})
	if err != nil {
		t.Fatal(err)
	}
	if message != "add: new.txt; update: updated.txt" {
		t.Fatalf("message = %q", message)
	}
}

func TestGenerateGitignorePreservesManualSection(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, ".gitignore"), "manual/\n"+Setting.Gitignore.MarkerStart+"\nold\n")
	config := &Config{
		DotfilesDir: dir,
		Sync:        SyncConfig{Ignore: []string{"raw"}},
	}

	if err := GenerateGitignore(config); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, expected := range []string{"manual/", "raw/", Setting.Path.HookDir + "/", Setting.Gitignore.MarkerEnd} {
		if !strings.Contains(content, expected) {
			t.Fatalf(".gitignoreに%qがない:\n%s", expected, content)
		}
	}
}

func TestWriteSyncConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sync.toml")
	writeTestFile(t, path, `default_branch = "main"`)

	expected := SyncConfig{
		DefaultBranch: "develop",
		Auto:          []string{"editor"},
		Manual:        []string{},
		Ignore:        []string{"raw"},
	}
	if err := writeSyncConfig(path, expected); err != nil {
		t.Fatal(err)
	}

	config, err := loadSyncConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if config.DefaultBranch != "develop" || len(config.Auto) != 1 || config.Auto[0] != "editor" {
		t.Fatalf("unexpected config: %#v", config)
	}
	matches, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".sync.toml-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("一時ファイルが残っている: %#v", matches)
	}
	if _, err := os.Stat(path + ".backup"); !os.IsNotExist(err) {
		t.Fatalf("退避ファイルが残っている: %v", err)
	}
}

func TestDefaultBranchIntegration(t *testing.T) {
	dir := newTestRepository(t, "develop")
	current := strings.TrimSpace(runGit(t, dir, "branch", "--show-current"))
	if current != "develop" {
		t.Fatalf("current branch = %q", current)
	}
}

func TestPushPullAndDeleteCategory(t *testing.T) {
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	runGit(t, root, "init", "--bare", remote)

	first := filepath.Join(root, "first")
	if err := os.Mkdir(first, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, first, "init", "-b", "develop")
	runGit(t, first, "config", "user.name", "dotfile test")
	runGit(t, first, "config", "user.email", "dotfile@example.invalid")
	runGit(t, first, "remote", "add", "origin", remote)
	writeTestFile(t, filepath.Join(first, Setting.Path.InfraVersionFile), "1.0.0")
	writeTestFile(t, filepath.Join(first, Setting.Path.SyncConfigFile), "default_branch = \"develop\"\nauto = [\"editor\"]\nmanual = []\nignore = []\n")
	writeTestFile(t, filepath.Join(first, "editor", "settings.json"), "one")
	runGit(t, first, "add", "-A")
	runGit(t, first, "commit", "-m", "initial")
	runGit(t, first, "push", "-u", "origin", "develop")

	EngineVersion = "1.0.0"
	config, err := loadConfig(first)
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	writeTestFile(t, filepath.Join(first, "editor", "settings.json"), "two")
	if err := Push(config, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}

	second := filepath.Join(root, "second")
	runGit(t, root, "clone", "--branch", "develop", remote, second)
	runGit(t, second, "config", "user.name", "dotfile test")
	runGit(t, second, "config", "user.email", "dotfile@example.invalid")
	writeTestFile(t, filepath.Join(second, "editor", "remote.txt"), "remote")
	runGit(t, second, "add", "-A")
	runGit(t, second, "commit", "-m", "remote update")
	runGit(t, second, "push", "origin", "develop")

	stdout.Reset()
	stderr.Reset()
	if err := Pull(config, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(first, "editor", "remote.txt")); err != nil {
		t.Fatalf("pull did not update worktree: %v", err)
	}

	stdout.Reset()
	stderr.Reset()
	if err := DeleteCategory(config, "editor", &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(first, "editor")); !os.IsNotExist(err) {
		t.Fatalf("category still exists: %v", err)
	}
	updated, err := loadSyncConfig(filepath.Join(first, Setting.Path.SyncConfigFile))
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(updated.Auto, "editor") {
		t.Fatal("deleted category remains in sync.toml")
	}
}

func newTestRepository(t *testing.T, branch string) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-b", branch)
	runGit(t, dir, "config", "user.name", "dotfile test")
	runGit(t, dir, "config", "user.email", "dotfile@example.invalid")
	return dir
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	output, err := (GitRunner{WorkDir: dir}).Output(args...)
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return output
}
