# アーキテクチャ

## 全体構成

`dotfile` はGoのシングルバイナリとして動作し、ユーザー固有データは別のGitリポジトリで管理する。

```text
dotfile engine
├── cmd/
│   ├── main.go          エントリポイントと終了コード
│   ├── root.go          Cobraルートコマンド、application struct
│   ├── init.go          initサブコマンド
│   ├── setup.go         setupサブコマンド
│   ├── link.go          linkサブコマンド
│   ├── sync.go          pull / pushサブコマンド
│   ├── status.go        status / delete-category / gitignoreサブコマンド
│   └── version.go       versionサブコマンド
├── internal/
│   ├── conf.go          リポジトリ解決とTOML設定
│   ├── conf.toml        エンジン内部定数（パス名、Gitキー、hookソース等）
│   ├── git.go           gitコマンド実行
│   ├── link.go          symlink配置
│   ├── setup.go         initとsetup
│   ├── sync.go          pull、push、削除、gitignore、status
│   ├── tool.go          OS名変換、ホーム展開、ファイル置換等のユーティリティ
│   └── hook/
│       ├── pre-push     Git hook（シェルスクリプト）
│       └── post-merge   Git hook（シェルスクリプト）
├── template/            init展開用テンプレート
├── embed.go             VERSION、template、hookの埋め込み
└── VERSION

user data repository
├── .infra-version
├── sync.toml
└── <category>/
    └── link.toml
```

エンジンには個人データを含めない。`template/` は `dotfile init` で展開する初期雛形だけを持つ。

## 境界と責務

### CLI（cmd/）

Cobraコマンド定義を `cmd/` にサブコマンドごとのファイルで配置する。
各 `RunE` はエラーを返し、プロセス終了は `cmd/main.go` だけが担当する。
`application` structが埋め込みリソースを保持し、`internal` パッケージのロジックへ委譲する。

### ビジネスロジック（internal/）

import path: `github.com/kokukaityo/dotfile/internal`（package名 `engine`）。
コマンド定義を持たず、純粋なロジックだけを公開する。

### 設定

データリポジトリは `DOTFILES_DIR`、現在のGitルート、`~/dotfiles` の順に探索する。
`sync.toml` の `default_branch` は空値を禁止し、`git check-ref-format --branch` で検証する。
エンジンとデータの `.infra-version` はメジャーバージョンを比較し、不一致時は警告する。

### Git

go-gitは使わず、ユーザーのSSH・credential・hook設定を利用するため `os/exec` でGit CLIを呼ぶ。
自動pushは `auto` カテゴリのpathspecだけをstage・commitする。既定ブランチ以外では実行しない。

pull時に分岐を検出した場合は自動mergeせず、ローカル履歴を `conflict/<hostname>/<timestamp>` へ退避する。

### symlink

各カテゴリの `link.toml` を読み、`runtime.GOOS` に対応するセクションだけを処理する。
リンク元はデータリポジトリ内、リンク先はユーザー環境に置く。既存のリンク先はバックアップしてから置換する。

### hooks

hook本体は互換性のためシェルスクリプトを維持する。バイナリに埋め込み、setup時にデータリポジトリの
`.dotfile-hook/` へ書き出して `core.hooksPath` を設定する。このディレクトリはGit追跡対象外。

hookスクリプトは `internal/hook/` に配置し、ルートの `embed.go` から埋め込む。

### 埋め込み（embed.go）

ルート直下の `embed.go`（package dotfile）が以下を公開する:

- `TemplateFS`: `template/` ディレクトリ全体
- `Version`: `VERSION` ファイルの内容
- `HookFS`: `internal/hook/` 配下のシェルスクリプト

`//go:embed` はソースファイルからの相対パスでしか参照できないため、`embed.go` はルートに置く必要がある。

## 設定形式

`sync.toml`:

```toml
default_branch = "main"
auto = ["editor"]
manual = ["shell"]
ignore = ["backup"]
```

`link.toml`:

```toml
[darwin]
"settings.json" = ["~/Library/Application Support/Code/User/settings.json"]

[win32]
"settings.json" = ["~/AppData/Roaming/Code/User/settings.json"]
```

旧 `sync.conf` と `link.yaml` の互換・移行処理は持たない。

## ファイル更新の安全性

`delete-category` は既存の `sync.toml` に未コミット変更がないことを確認する。
更新内容を同一ディレクトリの一時ファイルへ書き、Unixではrename、Windowsでは退避・置換・失敗時復元を行う。
カテゴリ削除と設定更新は1つのcommitにまとめる。
