// platform.go は OS 依存の変換処理を集約する。
package engine

import (
	"fmt"
	"runtime"
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
