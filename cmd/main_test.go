package main

import (
	"bytes"
	"strings"
	"testing"
	"testing/fstest"

	engine "github.com/kokukaityo/dotfiles-bridge/internal"
)

func TestVersionCommandWithoutDataRepository(t *testing.T) {
	engine.EngineVersion = "1.2.3"
	app := &application{templateFS: fstest.MapFS{}, hookFS: fstest.MapFS{}}
	command := app.rootCommand()
	var stdout bytes.Buffer
	command.SetArgs([]string{"version"})
	command.SetOut(&stdout)
	command.SetErr(&bytes.Buffer{})

	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "dotfiles engine v1.2.3") {
		t.Fatalf("unexpected output: %s", stdout.String())
	}
}

func TestRootCommandReturnsErrorWithoutExit(t *testing.T) {
	existing := t.TempDir()
	engine.EngineVersion = "1.0.0"
	app := &application{templateFS: fstest.MapFS{}, hookFS: fstest.MapFS{}}
	command := app.rootCommand()
	command.SetArgs([]string{"init", existing})
	command.SetOut(&bytes.Buffer{})
	command.SetErr(&bytes.Buffer{})

	if err := command.Execute(); err == nil {
		t.Fatal("existing path did not return an error")
	}
}
