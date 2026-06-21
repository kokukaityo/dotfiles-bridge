// platform.go は OS 依存の変換処理を集約する。
// link.toml のセクション名やパス中の ~ 展開など、プラットフォーム差を吸収する。
package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// OSKey は runtime.GOOS を link.toml のセクションキーに変換する。
// "windows" → "win32" の変換は、link.toml のキーを
// Node.js の process.platform に合わせているため。
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

// ExpandHome はパス先頭の ~ をホームディレクトリに展開する。
// Go 標準ライブラリには ~ 展開がないため自前で実装。
// link.toml のターゲットパスが "~/..." 形式で書かれるので、各所から呼ばれる。
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
