package engine

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegisterAndUnregisterServiceFiles(t *testing.T) {
	for _, goos := range []string{"linux", "darwin", "windows"} {
		t.Run(goos, func(t *testing.T) {
			home := t.TempDir()
			executable := "/opt/dotfiles/bin/dotfiles"
			if goos == "windows" {
				executable = `C:\Program Files\dotfiles\dotfiles.exe`
			}
			var commands []string
			withServiceTestHooks(t, goos, home, executable, func(name string, args ...string) error {
				commands = append(commands, name+" "+strings.Join(args, " "))
				return nil
			})

			var stdout bytes.Buffer
			config := &Config{DotfilesDir: t.TempDir()}
			if err := RegisterService(config, &stdout); err != nil {
				t.Fatal(err)
			}
			path, err := servicePath(goos)
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			content := string(data)
			assertServiceContent(t, goos, content, executable)

			if err := UnregisterService(config, &stdout); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("service file still exists: %v", err)
			}
			assertServiceCommands(t, goos, commands, path)
		})
	}
}

func withServiceTestHooks(t *testing.T, goos, home, executable string, run func(string, ...string) error) {
	t.Helper()
	originalGOOS := serviceGOOS
	originalExecutable := serviceExecutable
	originalHomeDir := serviceHomeDir
	originalRunCommand := serviceRunCommand
	serviceGOOS = goos
	serviceExecutable = func() (string, error) {
		return executable, nil
	}
	serviceHomeDir = func() (string, error) {
		return home, nil
	}
	serviceRunCommand = run
	t.Cleanup(func() {
		serviceGOOS = originalGOOS
		serviceExecutable = originalExecutable
		serviceHomeDir = originalHomeDir
		serviceRunCommand = originalRunCommand
	})
}

func assertServiceContent(t *testing.T, goos, content, executable string) {
	t.Helper()
	if !strings.Contains(content, "watch") {
		t.Fatalf("%s service content does not contain %q:\n%s", goos, "watch", content)
	}
	if !strings.Contains(content, filepath.Base(executable)) {
		t.Fatalf("%s service content does not contain executable base %q:\n%s", goos, filepath.Base(executable), content)
	}
	switch goos {
	case "linux":
		for _, expected := range []string{"[Unit]", "ExecStart=", "Restart=on-failure", "WantedBy=default.target"} {
			if !strings.Contains(content, expected) {
				t.Fatalf("linux service content does not contain %q:\n%s", expected, content)
			}
		}
	case "darwin":
		for _, expected := range []string{"<key>Label</key>", launchdLabel, "<key>ProgramArguments</key>", "watch.log"} {
			if !strings.Contains(content, expected) {
				t.Fatalf("darwin service content does not contain %q:\n%s", expected, content)
			}
		}
	case "windows":
		for _, expected := range []string{"WScript.Shell", "\"\" watch", "False"} {
			if !strings.Contains(content, expected) {
				t.Fatalf("windows service content does not contain %q:\n%s", expected, content)
			}
		}
	}
}

func assertServiceCommands(t *testing.T, goos string, commands []string, path string) {
	t.Helper()
	switch goos {
	case "linux":
		expected := []string{
			"systemctl --user daemon-reload",
			"systemctl --user enable --now " + systemdUnitName,
			"systemctl --user disable --now " + systemdUnitName,
			"systemctl --user daemon-reload",
		}
		if strings.Join(commands, "\n") != strings.Join(expected, "\n") {
			t.Fatalf("commands = %#v", commands)
		}
	case "darwin":
		expected := []string{
			"launchctl load -w " + path,
			"launchctl unload -w " + path,
		}
		if strings.Join(commands, "\n") != strings.Join(expected, "\n") {
			t.Fatalf("commands = %#v", commands)
		}
	case "windows":
		if len(commands) != 0 {
			t.Fatalf("windows commands = %#v", commands)
		}
	}
}
