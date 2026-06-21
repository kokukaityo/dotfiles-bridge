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
		Setting.Path.TemplateDir + "/" + Setting.Path.SyncConfigFile: {
			Data: []byte("default_branch = \"develop\"\nauto = []\nmanual = []\nignore = []\n"),
		},
	}
	hookFS := make(fstest.MapFS)
	for _, source := range Setting.Hook.Sources {
		hookFS[source] = &fstest.MapFile{Data: []byte("#!/usr/bin/env bash\nexit 0\n")}
	}
	target := filepath.Join(t.TempDir(), "repository")
	var stdout bytes.Buffer

	EngineVersion = "1.0.0"
	if err := InitializeRepository(target, templateFS, hookFS, &stdout); err != nil {
		t.Fatal(err)
	}
	if branch := runGit(t, target, "branch", "--show-current"); branch != "develop" {
		t.Fatalf("branch = %q", branch)
	}
	if status := runGit(t, target, "status", "--porcelain"); status != "" {
		t.Fatalf("worktree is dirty:\n%s", status)
	}
	hookCheckPath := Setting.Path.HookDir + "/pre-push"
	if ignored := runGit(t, target, "check-ignore", hookCheckPath); ignored != hookCheckPath {
		t.Fatalf("hook is not ignored: %q", ignored)
	}
	if runtime.GOOS != "windows" {
		info, err := fs.Stat(os.DirFS(target), Setting.Path.HookDir+"/pre-push")
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
