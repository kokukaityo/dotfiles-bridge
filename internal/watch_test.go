package engine

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestWatchPushesAfterDebounce(t *testing.T) {
	dir := newTestRepository(t, "main")
	writeTestFile(t, filepath.Join(dir, "editor", ".keep"), "")
	config := &Config{
		DotfilesDir: dir,
		Sync:        SyncConfig{DefaultBranch: "main", Auto: []string{"editor"}},
	}

	originalDebounce := watchDebounce
	originalPush := pushForWatch
	watchDebounce = 50 * time.Millisecond
	pushed := make(chan struct{}, 1)
	pushForWatch = func(*Config, io.Writer, io.Writer) error {
		pushed <- struct{}{}
		return nil
	}
	t.Cleanup(func() {
		watchDebounce = originalDebounce
		pushForWatch = originalPush
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var stdout, stderr bytes.Buffer
	errCh := make(chan error, 1)
	go func() {
		errCh <- watchWithContext(ctx, config, &stdout, &stderr)
	}()
	waitForOutput(t, &stdout, "監視を開始")

	writeTestFile(t, filepath.Join(dir, "editor", "settings.json"), "changed")

	select {
	case <-pushed:
	case <-time.After(3 * time.Second):
		t.Fatalf("Push was not called; stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("watch did not stop")
	}
}

func TestWatchPIDGuard(t *testing.T) {
	dir := t.TempDir()
	config := &Config{DotfilesDir: dir}
	writeTestFile(t, filepath.Join(dir, watchPIDFile), strconv.Itoa(os.Getpid())+"\n")

	_, err := acquireWatchPID(config)
	if err == nil {
		t.Fatal("expected duplicate watch error")
	}
	if !strings.Contains(err.Error(), "既に稼働中") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func waitForOutput(t *testing.T, buffer *bytes.Buffer, substring string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(buffer.String(), substring) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("output did not contain %q: %s", substring, buffer.String())
}
