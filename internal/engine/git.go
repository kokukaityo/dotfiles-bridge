package engine

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type GitRunner struct {
	WorkDir string
	Stdout  io.Writer
	Stderr  io.Writer
}

func (g GitRunner) command(args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.WorkDir
	cmd.Stdout = g.Stdout
	cmd.Stderr = g.Stderr
	return cmd
}

func (g GitRunner) Run(args ...string) error {
	cmd := g.command(args...)
	if cmd.Stdout == nil {
		cmd.Stdout = os.Stdout
	}
	if cmd.Stderr == nil {
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

func (g GitRunner) Output(args ...string) (string, error) {
	cmd := g.command(args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if cmd.Stderr == nil {
		cmd.Stderr = &bytes.Buffer{}
	}
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func (g GitRunner) Success(args ...string) bool {
	cmd := g.command(args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}
