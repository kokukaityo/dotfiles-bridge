package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type SyncConfig struct {
	DefaultBranch string   `toml:"default_branch"`
	Auto          []string `toml:"auto"`
	Manual        []string `toml:"manual"`
	Ignore        []string `toml:"ignore"`
}

type Config struct {
	EngineVersion string
	DotfilesDir   string
	DataVersion   string
	Sync          SyncConfig
}

func Resolve(engineVersion string) (*Config, error) {
	dir, err := resolveDotfilesDir()
	if err != nil {
		return nil, err
	}
	return loadConfig(dir, engineVersion)
}

func resolveDotfilesDir() (string, error) {
	if envDir := os.Getenv("DOTFILES_DIR"); envDir != "" {
		dir, err := filepath.Abs(envDir)
		if err != nil {
			return "", fmt.Errorf("DOTFILES_DIRを解決できません: %w", err)
		}
		if isDataRepository(dir) {
			return dir, nil
		}
	}

	if root, err := (GitRunner{}).Output("rev-parse", "--show-toplevel"); err == nil && isDataRepository(root) {
		return filepath.Clean(root), nil
	}

	home, err := os.UserHomeDir()
	if err == nil {
		dir := filepath.Join(home, "dotfiles")
		if isDataRepository(dir) {
			return dir, nil
		}
	}

	return "", fmt.Errorf("データリポジトリが見つかりません。DOTFILES_DIRを設定するか、データリポジトリ内で実行してください")
}

func isDataRepository(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".infra-version"))
	return err == nil && !info.IsDir()
}

func loadConfig(dir, engineVersion string) (*Config, error) {
	syncConfig, err := loadSyncConfig(filepath.Join(dir, "sync.toml"))
	if err != nil {
		return nil, fmt.Errorf("sync.tomlを読み込めません: %w", err)
	}
	if err := validateDefaultBranch(dir, syncConfig.DefaultBranch); err != nil {
		return nil, err
	}

	versionBytes, err := os.ReadFile(filepath.Join(dir, ".infra-version"))
	if err != nil {
		return nil, fmt.Errorf(".infra-versionを読み込めません: %w", err)
	}

	config := &Config{
		EngineVersion: strings.TrimSpace(engineVersion),
		DotfilesDir:   filepath.Clean(dir),
		DataVersion:   strings.TrimSpace(string(versionBytes)),
		Sync:          syncConfig,
	}
	return config, nil
}

func loadSyncConfig(path string) (SyncConfig, error) {
	var config SyncConfig
	_, err := toml.DecodeFile(path, &config)
	return config, err
}

func validateDefaultBranch(workDir, branch string) error {
	if strings.TrimSpace(branch) == "" {
		return fmt.Errorf("sync.tomlのdefault_branchは空にできません")
	}
	if !(GitRunner{WorkDir: workDir}).Success("check-ref-format", "--branch", branch) {
		return fmt.Errorf("不正なdefault_branchです: %s", branch)
	}
	return nil
}

func majorVersion(version string) string {
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	major, _, _ := strings.Cut(version, ".")
	return major
}

func (c *Config) VersionMismatch() bool {
	return majorVersion(c.EngineVersion) != majorVersion(c.DataVersion)
}
