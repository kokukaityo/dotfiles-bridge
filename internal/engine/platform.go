package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func OSKey() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return "win32", nil
	case "darwin", "linux":
		return runtime.GOOS, nil
	default:
		return "", fmt.Errorf("未対応のOSです: %s", runtime.GOOS)
	}
}

func ExpandHome(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") && !strings.HasPrefix(path, `~\`) {
		return filepath.Clean(path), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("ホームディレクトリを取得できません: %w", err)
	}
	if path == "~" {
		return home, nil
	}
	return filepath.Join(home, path[2:]), nil
}
