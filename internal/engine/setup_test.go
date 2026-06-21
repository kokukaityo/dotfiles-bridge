package engine

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"
)

func TestInitializeRepository(t *testing.T) {
	t.Setenv("GIT_AUTHOR_NAME", "dotfile test")
	t.Setenv("GIT_AUTHOR_EMAIL", "dotfile@example.invalid")
	t.Setenv("GIT_COMMITTER_NAME", "dotfile test")
	t.Setenv("GIT_COMMITTER_EMAIL", "dotfile@example.invalid")

	templateFS := fstest.MapFS{
		"template/.infra-version": {Data: []byte("1.0.0\n")},
		"template/sync.toml": {
			Data: []byte("default_branch = \"develop\"\nauto = []\nmanual = []\nignore = []\n"),
		},
	}
	hookFS := fstest.MapFS{
		"lib/hooks/pre-push":   {Data: []byte("#!/usr/bin/env bash\nexit 0\n")},
		"lib/hooks/post-merge": {Data: []byte("#!/usr/bin/env bash\nexit 0\n")},
	}
	target := filepath.Join(t.TempDir(), "repository")
	var stdout bytes.Buffer
	app := &application{templateFS: templateFS, hookFS: hookFS, engineVersion: "1.0.0"}

	if err := initializeRepository(target, app, &stdout); err != nil {
		t.Fatal(err)
	}
	if branch := runGit(t, target, "branch", "--show-current"); branch != "develop" {
		t.Fatalf("branch = %q", branch)
	}
	if status := runGit(t, target, "status", "--porcelain"); status != "" {
		t.Fatalf("worktree is dirty:\n%s", status)
	}
	if ignored := runGit(t, target, "check-ignore", ".dotfile-hook/pre-push"); ignored != ".dotfile-hook/pre-push" {
		t.Fatalf("hook is not ignored: %q", ignored)
	}
	if runtime.GOOS != "windows" {
		info, err := fs.Stat(os.DirFS(target), ".dotfile-hook/pre-push")
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != hookFileMode {
			t.Fatalf("hook mode = %o", info.Mode().Perm())
		}
	}
	if !strings.Contains(stdout.String(), "作成が完了") {
		t.Fatalf("unexpected output: %s", stdout.String())
	}
}

func TestRootCommandReturnsErrorWithoutExit(t *testing.T) {
	existing := t.TempDir()
	app := &application{templateFS: fstest.MapFS{}, hookFS: fstest.MapFS{}, engineVersion: "1.0.0"}
	command := app.rootCommand()
	command.SetArgs([]string{"init", existing})
	command.SetOut(&bytes.Buffer{})
	command.SetErr(&bytes.Buffer{})

	if err := command.Execute(); err == nil {
		t.Fatal("existing path did not return an error")
	}
}
