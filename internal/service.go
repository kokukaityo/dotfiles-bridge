package engine

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	systemdUnitName = "dotfiles-watch.service"
	launchdLabel    = "com.dotfiles.watch"
	startupVBSName  = "dotfiles-watch.vbs"
)

var (
	serviceGOOS       = runtime.GOOS
	serviceExecutable = os.Executable
	serviceHomeDir    = os.UserHomeDir
	serviceRunCommand = runServiceCommand
)

func RegisterService(config *Config, stdout io.Writer) error {
	executable, err := serviceExecutable()
	if err != nil {
		return fmt.Errorf("実行ファイルのパスを取得できません: %w", err)
	}
	path, content, err := serviceFile(serviceGOOS, executable)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("サービス設定ディレクトリを作成できません: %w", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("サービス設定を書き込めません: %w", err)
	}
	if err := enableService(serviceGOOS, path); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "[dotfiles] watchサービスを登録しました: %s\n", path)
	_ = config
	return nil
}

func UnregisterService(config *Config, stdout io.Writer) error {
	path, err := servicePath(serviceGOOS)
	if err != nil {
		return err
	}
	if err := disableService(serviceGOOS, path); err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("サービス設定を削除できません: %w", err)
	}
	_, _ = fmt.Fprintf(stdout, "[dotfiles] watchサービスを解除しました: %s\n", path)
	_ = config
	return nil
}

func serviceFile(goos, executable string) (string, string, error) {
	path, err := servicePath(goos)
	if err != nil {
		return "", "", err
	}
	switch goos {
	case "linux":
		return path, fmt.Sprintf(`[Unit]
Description=dotfiles watch - auto push on file change

[Service]
ExecStart=%s watch
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
`, systemdEscapePath(executable)), nil
	case "darwin":
		logPath, err := launchdLogPath()
		if err != nil {
			return "", "", err
		}
		return path, fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>%s</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>watch</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>%s</string>
    <key>StandardErrorPath</key>
    <string>%s</string>
</dict>
</plist>
`, launchdLabel, xmlEscape(executable), xmlEscape(logPath), xmlEscape(logPath)), nil
	case "windows":
		return path, fmt.Sprintf("CreateObject(\"WScript.Shell\").Run \"\"\"%s\"\" watch\", 0, False\r\n", strings.ReplaceAll(executable, `"`, `""`)), nil
	default:
		return "", "", fmt.Errorf("watchサービス登録はこのOSに対応していません: %s", goos)
	}
}

func servicePath(goos string) (string, error) {
	home, err := serviceHomeDir()
	if err != nil {
		return "", fmt.Errorf("ホームディレクトリを取得できません: %w", err)
	}
	switch goos {
	case "linux":
		return filepath.Join(home, ".config", "systemd", "user", systemdUnitName), nil
	case "darwin":
		return filepath.Join(home, "Library", "LaunchAgents", launchdLabel+".plist"), nil
	case "windows":
		return filepath.Join(home, "AppData", "Roaming", "Microsoft", "Windows", "Start Menu", "Programs", "Startup", startupVBSName), nil
	default:
		return "", fmt.Errorf("watchサービス登録はこのOSに対応していません: %s", goos)
	}
}

func launchdLogPath() (string, error) {
	home, err := serviceHomeDir()
	if err != nil {
		return "", fmt.Errorf("ホームディレクトリを取得できません: %w", err)
	}
	return filepath.Join(home, ".local", "state", "dotfiles", "watch.log"), nil
}

func enableService(goos, path string) error {
	switch goos {
	case "linux":
		if err := serviceRunCommand("systemctl", "--user", "daemon-reload"); err != nil {
			return fmt.Errorf("systemd設定の再読み込みに失敗しました: %w", err)
		}
		if err := serviceRunCommand("systemctl", "--user", "enable", "--now", systemdUnitName); err != nil {
			return fmt.Errorf("systemdサービスの有効化に失敗しました: %w", err)
		}
	case "darwin":
		if err := serviceRunCommand("launchctl", "load", "-w", path); err != nil {
			return fmt.Errorf("launchdエージェントの読み込みに失敗しました: %w", err)
		}
	case "windows":
		return nil
	}
	return nil
}

func disableService(goos, path string) error {
	switch goos {
	case "linux":
		if err := serviceRunCommand("systemctl", "--user", "disable", "--now", systemdUnitName); err != nil {
			return fmt.Errorf("systemdサービスの無効化に失敗しました: %w", err)
		}
		if err := serviceRunCommand("systemctl", "--user", "daemon-reload"); err != nil {
			return fmt.Errorf("systemd設定の再読み込みに失敗しました: %w", err)
		}
	case "darwin":
		if err := serviceRunCommand("launchctl", "unload", "-w", path); err != nil {
			return fmt.Errorf("launchdエージェントの解除に失敗しました: %w", err)
		}
	case "windows":
		return nil
	}
	return nil
}

func runServiceCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

func systemdEscapePath(path string) string {
	if !strings.ContainsAny(path, " \t\n\"'\\") {
		return path
	}
	return `"` + strings.ReplaceAll(strings.ReplaceAll(path, `\`, `\\`), `"`, `\"`) + `"`
}

func xmlEscape(s string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;")
	return replacer.Replace(s)
}
