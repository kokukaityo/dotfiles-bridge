package engine

import (
	"bytes"
	"strings"
	"testing"
	"testing/fstest"
)

func TestVersionCommandWithoutDataRepository(t *testing.T) {
	app := &application{templateFS: fstest.MapFS{}, hookFS: fstest.MapFS{}, engineVersion: "1.2.3"}
	command := app.rootCommand()
	var stdout bytes.Buffer
	command.SetArgs([]string{"version"})
	command.SetOut(&stdout)
	command.SetErr(&bytes.Buffer{})

	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "dotfile engine v1.2.3") {
		t.Fatalf("unexpected output: %s", stdout.String())
	}
}
