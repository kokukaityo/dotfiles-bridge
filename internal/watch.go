package engine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

const watchPIDFile = ".dotfiles-watch.pid"

var (
	watchDebounce = 3 * time.Second
	pushForWatch  = Push
)

func Watch(config *Config, stdout, stderr io.Writer) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return watchWithContext(ctx, config, stdout, stderr)
}

func watchWithContext(ctx context.Context, config *Config, stdout, stderr io.Writer) error {
	release, err := acquireWatchPID(config)
	if err != nil {
		return err
	}
	defer release()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("ファイル監視を開始できません: %w", err)
	}
	defer func() {
		_ = watcher.Close()
	}()

	for _, category := range config.Sync.Auto {
		root := RepositoryPath(config, category)
		if err := addWatchTree(watcher, root); err != nil {
			return err
		}
	}
	_, _ = fmt.Fprintln(stdout, "[watch] 自動同期カテゴリの監視を開始しました。")

	var mu sync.Mutex
	var timer *time.Timer
	defer func() {
		mu.Lock()
		if timer != nil {
			timer.Stop()
		}
		mu.Unlock()
	}()

	schedulePush := func() {
		mu.Lock()
		defer mu.Unlock()
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(watchDebounce, func() {
			if err := pushForWatch(config, stdout, stderr); err != nil {
				_, _ = fmt.Fprintf(stderr, "[watch] WARNING: pushに失敗しました: %v\n", err)
			}
		})
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Has(fsnotify.Create) {
				if info, statErr := os.Stat(event.Name); statErr == nil && info.IsDir() {
					if err := addWatchTree(watcher, event.Name); err != nil {
						_, _ = fmt.Fprintf(stderr, "[watch] WARNING: 監視対象を追加できません: %v\n", err)
					}
				}
			}
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
				schedulePush()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			_, _ = fmt.Fprintf(stderr, "[watch] WARNING: 監視エラー: %v\n", err)
		}
	}
}

func addWatchTree(watcher *fsnotify.Watcher, root string) error {
	info, err := os.Stat(root)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("監視対象を確認できません (%s): %w", root, err)
	}
	if !info.IsDir() {
		return nil
	}
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}
		if err := watcher.Add(path); err != nil {
			return fmt.Errorf("監視対象を追加できません (%s): %w", path, err)
		}
		return nil
	})
}

func acquireWatchPID(config *Config) (func(), error) {
	path := RepositoryPath(config, watchPIDFile)
	if data, err := os.ReadFile(path); err == nil {
		pid, parseErr := strconv.Atoi(strings.TrimSpace(string(data)))
		if parseErr == nil && isProcessRunning(pid) {
			return nil, fmt.Errorf("dotfiles watchは既に稼働中です (pid=%d)", pid)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("PIDファイルを読めません: %w", err)
	}
	if err := os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())+"\n"), 0o644); err != nil {
		return nil, fmt.Errorf("PIDファイルを書き込めません: %w", err)
	}
	return func() {
		_ = os.Remove(path)
	}, nil
}

func isProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	if pid == os.Getpid() {
		return true
	}
	if runtime.GOOS == "windows" {
		output, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH").Output()
		return err == nil && strings.Contains(string(output), strconv.Itoa(pid))
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil || errors.Is(err, syscall.EPERM)
}
