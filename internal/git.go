// git.go は Git CLI の実行を抽象化する。
// go-git ではなく os/exec で実 git を呼ぶことで、
// ユーザーの SSH 鍵・credential helper・hook 設定をそのまま使える。
package engine

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// GitRunner は作業ディレクトリと出力先を束ねた Git コマンド実行器。
// WorkDir 未指定時はカレントディレクトリで動く（resolveDotfilesDir の git root 検出で使用）。
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

// Run は副作用のある Git コマンド（commit, push, checkout 等）を実行する。
// stdout/stderr はユーザーに直接表示される。
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

// Output は結果を文字列で取得する Git コマンド（rev-parse, branch --show-current 等）用。
// stdout はキャプチャされ、ユーザーには表示されない。
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

// Success は成否だけを知りたい Git コマンド（check-ref-format, diff --quiet 等）用。
// 出力は全て捨てる。
func (g GitRunner) Success(args ...string) bool {
	cmd := g.command(args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}
