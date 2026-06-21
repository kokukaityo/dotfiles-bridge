package engine

import (
	"runtime"
	"testing"
)

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
