package engine

import (
	"os"
	"path/filepath"
	"strings"
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

func TestLoadSyncConfigRejectsInvalidCategoryNames(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "path separator in auto",
			content: `auto = ["../escape"]`,
		},
		{
			name:    "absolute path in ignore",
			content: `ignore = ["/tmp/secret"]`,
		},
		{
			name:    "dot category",
			content: `auto = ["."]`,
		},
		{
			name:    "reserved file",
			content: `auto = ["sync.toml"]`,
		},
		{
			name:    "leading dot internal directory",
			content: `ignore = [".backup"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "sync.toml")
			writeTestFile(t, path, tt.content)
			if _, err := loadSyncConfig(path); err == nil {
				t.Fatal("不正なカテゴリ名が受理された")
			}
		})
	}
}

func TestLoadSyncConfigRejectsDuplicateCategories(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "duplicate auto",
			content: `auto = ["editor", "editor"]`,
		},
		{
			name:    "duplicate ignore",
			content: `ignore = ["raw", "raw"]`,
		},
		{
			name:    "auto ignore conflict",
			content: "auto = [\"editor\"]\nignore = [\"editor\"]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "sync.toml")
			writeTestFile(t, path, tt.content)
			if _, err := loadSyncConfig(path); err == nil || !strings.Contains(err.Error(), "カテゴリ") {
				t.Fatal("重複カテゴリが受理された")
			}
		})
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
