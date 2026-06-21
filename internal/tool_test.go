package engine

import (
	"os"
	"path/filepath"
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
