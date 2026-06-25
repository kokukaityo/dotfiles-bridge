// embed.go はビルド時にバイナリへ埋め込む静的リソースを宣言する。
// 埋め込み対象はソースファイルからの相対パスしか参照できないため、ルートに配置する必要がある。
package dotfiles

import "embed"

// TemplateFS は dotfiles init で展開するデータリポジトリの雛形一式。
//
//go:embed all:template
var TemplateFS embed.FS

// Version はエンジンのバージョン文字列。dotfiles version で表示される。
//
//go:embed VERSION
var Version string

// HookFS は dotfiles install でデータリポジトリの .dotfiles-hook/ へ書き出す Git hook スクリプト。
//
//go:embed internal/hook/pre-push internal/hook/post-merge
var HookFS embed.FS
