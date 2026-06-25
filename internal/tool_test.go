package engine

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ExpandPath("~/example")
	if err != nil {
		t.Fatal(err)
	}
	expected := filepath.Join(home, "example")
	if got != expected {
		t.Fatalf("ExpandHome() = %q, want %q", got, expected)
	}
}

func TestOSKey(t *testing.T) {
	t.Parallel()

	got, err := OSKey()
	if err != nil {
		t.Fatal(err)
	}
	expected := runtime.GOOS
	if runtime.GOOS == "windows" {
		expected = "win32"
	}
	if got != expected {
		t.Fatalf("OSKey() = %q, want %q", got, expected)
	}
}
