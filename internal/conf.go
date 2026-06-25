// conf.go はエンジンの設定を担当する。
// 内部設定値を conf.toml から読み込み、データリポジトリの探索と設定の読み込みを行う。
// cmd/ 層の各サブコマンドは Resolve() を経由してここから Config を受け取り、
// 以降のロジック（sync, link, setup）に渡す。
package engine

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// --- エンジン内部設定（conf.toml） ---

//go:embed conf.toml
var settingData string

var Setting = mustParseSetting(settingData)

func mustParseSetting(data string) EngineSetting {
	var s EngineSetting
	if _, err := toml.Decode(data, &s); err != nil {
		panic("conf.toml: " + err.Error())
	}
	return s
}

type EngineSetting struct {
	Path      PathSetting      `toml:"path"`
	Git       GitSetting       `toml:"git"`
	Hook      HookSetting      `toml:"hook"`
	Gitignore GitignoreSetting `toml:"gitignore"`
}

type PathSetting struct {
	DefaultDir         string `toml:"default_dir"`
	TemplateDir        string `toml:"template_dir"`
	HookDir            string `toml:"hook_dir"`
	BackupDir          string `toml:"backup_dir"`
	InfraVersionFile   string `toml:"infra_version_file"`
	SyncConfigFile     string `toml:"sync_config_file"`
	LinkConfigFile     string `toml:"link_config_file"`
	ConflictMarkerFile string `toml:"conflict_marker_file"`
}

type GitSetting struct {
	HooksPath         string `toml:"hooks_path"`
	Symlinks          string `toml:"symlinks"`
	GitattributesLine string `toml:"gitattributes_line"`
}

type HookSetting struct {
	Sources []string `toml:"sources"`
}

type GitignoreSetting struct {
	MarkerStart      string   `toml:"marker_start"`
	MarkerEnd        string   `toml:"marker_end"`
	SecurityPatterns []string `toml:"security_patterns"`
}

// --- ランタイム設定（データリポジトリ） ---

var (
	defaultDir       = Setting.Path.DefaultDir
	infraVersionFile = Setting.Path.InfraVersionFile

	DefaultDir = "~/" + defaultDir

	// EngineVersion はビルド時に embed された VERSION 文字列。
	// main.go が起動時にセットする。
	EngineVersion string
)

// SyncConfig は sync.toml をそのまま構造体にしたもの。
// カテゴリの同期モード（auto/ignore、どちらにも属さないカテゴリは manual 扱い）とブランチ設定を保持する。
type SyncConfig struct {
	Mode          string   `toml:"mode"`
	DefaultBranch string   `toml:"default_branch"`
	Auto          []string `toml:"auto"`
	Ignore        []string `toml:"ignore"`
}

// Config はランタイムで組み立てる実行時設定。
// SyncConfig（ファイル由来）に加えて、エンジンバージョンやリポジトリの絶対パスなど
// 実行環境から決まる情報を束ねる。ほぼ全ての internal 関数がこれを受け取る。
type Config struct {
	EngineVersion string
	DotfilesDir   string
	DataVersion   string
	Sync          SyncConfig
}

// Resolve は cmd 層から呼ばれる公開エントリポイント。
// 「データリポジトリを探す → 設定を読む」を一括で行い、Config を返す。
func Resolve() (*Config, error) {
	dir, err := resolveDotfilesDir()
	if err != nil {
		return nil, err
	}
	return loadConfig(dir)
}

// resolveDotfilesDir は3段フォールバックでデータリポジトリを探す。
//  1. DOTFILES_DIR 環境変数（明示指定を最優先）
//  2. カレントディレクトリの Git ルート（データリポジトリ内で作業中のケース）
//  3. ~/dotfiles（規約ベースのデフォルト）
//
// 各候補は isDataRepository で .infra-version の存在を確認してから採用する。
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
		dir := filepath.Join(home, defaultDir)
		if isDataRepository(dir) {
			return dir, nil
		}
	}

	return "", fmt.Errorf("データリポジトリが見つかりません。DOTFILES_DIRを設定するか、データリポジトリ内で実行してください")
}

// isDataRepository は .infra-version ファイルの存在でデータリポジトリかどうかを判定する。
// ただの Git リポジトリと区別するためのマーカー。ディレクトリの場合は false。
func isDataRepository(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, infraVersionFile))
	return err == nil && !info.IsDir()
}

// loadConfig は確定済みのディレクトリから sync.toml と .infra-version を読み、
// エンジンバージョンと合わせて Config を組み立てる。
func loadConfig(dir string) (*Config, error) {
	syncConfig, err := loadSyncConfig(filepath.Join(dir, syncConfigFile))
	if err != nil {
		return nil, fmt.Errorf("%sを読み込めません: %w", syncConfigFile, err)
	}
	if err := validateDefaultBranch(dir, syncConfig.DefaultBranch); err != nil {
		return nil, err
	}

	versionBytes, err := os.ReadFile(filepath.Join(dir, infraVersionFile))
	if err != nil {
		return nil, fmt.Errorf("%sを読み込めません: %w", infraVersionFile, err)
	}

	config := &Config{
		EngineVersion: strings.TrimSpace(EngineVersion),
		DotfilesDir:   filepath.Clean(dir),
		DataVersion:   strings.TrimSpace(string(versionBytes)),
		Sync:          syncConfig,
	}
	return config, nil
}

func loadSyncConfig(path string) (SyncConfig, error) {
	var config SyncConfig
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return config, err
	}
	if config.Mode == "" {
		config.Mode = "local"
	}
	if config.Mode != "local" && config.Mode != "remote" {
		return config, fmt.Errorf("sync.tomlのmodeは\"local\"か\"remote\"のみ有効です: %q", config.Mode)
	}
	if err := validateSyncConfigCategories(config); err != nil {
		return config, err
	}
	return config, nil
}

// validateDefaultBranch は git check-ref-format でブランチ名の安全性を検証する。
// sync.toml の値がそのまま git コマンドの引数になるため、インジェクション防止を兼ねる。
func validateDefaultBranch(workDir, branch string) error {
	if strings.TrimSpace(branch) == "" {
		return fmt.Errorf("sync.tomlのdefault_branchは空にできません")
	}
	if !(GitRunner{WorkDir: workDir}).Success("check-ref-format", "--branch", branch) {
		return fmt.Errorf("不正なdefault_branchです: %s", branch)
	}
	return nil
}

// majorVersion はバージョン文字列からメジャー番号だけを取り出す。
// VersionMismatch でメジャー単位の互換性チェックに使う。
// マイナー・パッチの差異は許容する設計。
func majorVersion(version string) string {
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	major, _, _ := strings.Cut(version, ".")
	return major
}

// VersionMismatch はエンジンとデータリポジトリのメジャーバージョンが異なるかを判定する。
// 不一致時は cmd/root.go の config() が stderr に警告を出す。
func (c *Config) VersionMismatch() bool {
	return majorVersion(c.EngineVersion) != majorVersion(c.DataVersion)
}

// RepositoryPath はデータリポジトリ内のパスを組み立てるヘルパー。
// sync.go, setup.go など複数箇所から呼ばれる。
func RepositoryPath(config *Config, names ...string) string {
	parts := append([]string{config.DotfilesDir}, names...)
	return filepath.Join(parts...)
}
