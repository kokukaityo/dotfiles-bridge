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
	stubServiceRegistration(t)
	t.Setenv("GIT_AUTHOR_NAME", "dotfiles test")
	t.Setenv("GIT_AUTHOR_EMAIL", "dotfiles@example.invalid")
	t.Setenv("GIT_COMMITTER_NAME", "dotfiles test")
	t.Setenv("GIT_COMMITTER_EMAIL", "dotfiles@example.invalid")

	templateFS := fstest.MapFS{
		Setting.Path.TemplateDir + "/" + Setting.Path.SyncConfigFile: {
			Data: []byte("default_branch = \"develop\"\nauto = []\nignore = []\n"),
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

func TestMigrateHookDir(t *testing.T) {
	t.Run("旧ディレクトリをリネーム", func(t *testing.T) {
		dir := t.TempDir()
		oldDir := filepath.Join(dir, legacyHookDir)
		if err := os.MkdirAll(oldDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(oldDir, "pre-push"), []byte("#!/bin/bash\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		var stdout bytes.Buffer
		if err := migrateHookDir(dir, &stdout); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
			t.Fatal("旧ディレクトリが残っている")
		}
		newDir := filepath.Join(dir, hookDir)
		if _, err := os.Stat(filepath.Join(newDir, "pre-push")); err != nil {
			t.Fatal("hookファイルがリネーム先にない")
		}
		if !strings.Contains(stdout.String(), "リネームしました") {
			t.Fatalf("unexpected output: %s", stdout.String())
		}
	})

	t.Run("両方存在する場合は旧ディレクトリを削除", func(t *testing.T) {
		dir := t.TempDir()
		oldDir := filepath.Join(dir, legacyHookDir)
		newDir := filepath.Join(dir, hookDir)
		if err := os.MkdirAll(oldDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(newDir, 0o755); err != nil {
			t.Fatal(err)
		}
		var stdout bytes.Buffer
		if err := migrateHookDir(dir, &stdout); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
			t.Fatal("旧ディレクトリが残っている")
		}
		if !strings.Contains(stdout.String(), "削除しました") {
			t.Fatalf("unexpected output: %s", stdout.String())
		}
	})

	t.Run("旧ディレクトリがなければ何もしない", func(t *testing.T) {
		dir := t.TempDir()
		var stdout bytes.Buffer
		if err := migrateHookDir(dir, &stdout); err != nil {
			t.Fatal(err)
		}
		if stdout.String() != "" {
			t.Fatalf("unexpected output: %s", stdout.String())
		}
	})
}

func stubServiceRegistration(t *testing.T) {
	t.Helper()
	originalGOOS := serviceGOOS
	originalExecutable := serviceExecutable
	originalHomeDir := serviceHomeDir
	originalRunCommand := serviceRunCommand
	serviceGOOS = "windows"
	serviceExecutable = func() (string, error) {
		return filepath.Join(t.TempDir(), "dotfiles.exe"), nil
	}
	serviceHomeDir = func() (string, error) {
		return t.TempDir(), nil
	}
	serviceRunCommand = func(string, ...string) error {
		return nil
	}
	t.Cleanup(func() {
		serviceGOOS = originalGOOS
		serviceExecutable = originalExecutable
		serviceHomeDir = originalHomeDir
		serviceRunCommand = originalRunCommand
	})
}
