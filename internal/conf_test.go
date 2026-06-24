package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, Setting.Path.InfraVersionFile), "1.2.0\n")
	writeTestFile(t, filepath.Join(dir, Setting.Path.SyncConfigFile), `
default_branch = "develop"
auto = ["editor"]
ignore = ["raw"]
`)

	EngineVersion = "1.3.0"
	config, err := loadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if config.Sync.DefaultBranch != "develop" {
		t.Fatalf("DefaultBranch = %q", config.Sync.DefaultBranch)
	}
	if config.VersionMismatch() {
		t.Fatal("同じメジャーバージョンを不整合と判定した")
	}
}

func TestLoadConfigRejectsInvalidBranch(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, Setting.Path.InfraVersionFile), "1.0.0")
	writeTestFile(t, filepath.Join(dir, Setting.Path.SyncConfigFile), `default_branch = "bad branch"`)

	EngineVersion = "1.0.0"
	if _, err := loadConfig(dir); err == nil {
		t.Fatal("不正なブランチ名が受理された")
	}
}

func TestVersionMismatch(t *testing.T) {
	t.Parallel()

	config := Config{EngineVersion: "2.0.0", DataVersion: "1.9.0"}
	if !config.VersionMismatch() {
		t.Fatal("異なるメジャーバージョンを一致と判定した")
	}
}

func TestLoadSyncConfigDefaultsToLocalMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sync.toml")
	writeTestFile(t, path, `default_branch = "main"`)
	config, err := loadSyncConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if config.Mode != "local" {
		t.Fatalf("Mode = %q, want %q", config.Mode, "local")
	}
}

func TestLoadSyncConfigRejectsInvalidMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sync.toml")
	writeTestFile(t, path, `mode = "invalid"`)
	if _, err := loadSyncConfig(path); err == nil {
		t.Fatal("invalid mode should be rejected")
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
