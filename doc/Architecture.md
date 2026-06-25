# アーキテクチャ

## 全体構成

`dotfiles-bridge` は `dotfiles` コマンドとして動作するGoのシングルバイナリで、ユーザー固有データは別のGitリポジトリで管理する。

```text
dotfiles-bridge engine
├── cmd/
│   ├── main.go          エントリポイントと終了コード
│   ├── root.go          Cobraルートコマンド、application struct
│   ├── init.go          initサブコマンド
│   ├── install.go       installサブコマンド
│   ├── link.go          linkサブコマンド
│   ├── sync.go          pull / pushサブコマンド
│   ├── status.go        status / delete-category / gitignoreサブコマンド
│   ├── watch.go         watchサブコマンド
│   └── version.go       versionサブコマンド
├── internal/
│   ├── conf.go          リポジトリ解決とTOML設定
│   ├── conf.toml        エンジン内部定数（パス名、Gitキー、hookソース等）
│   ├── git.go           gitコマンド実行
│   ├── link.go          symlink配置
│   ├── setup.go         initとinstall
│   ├── sync.go          pull、push、削除、gitignore、status
│   ├── watch.go         fsnotifyによるファイル監視とdebounce付き自動push
│   ├── service.go       OS別のログイン時自動起動サービス登録・解除
│   ├── tool.go          OS名変換、ホーム展開、ファイル置換等のユーティリティ
│   └── hook/
│       ├── pre-push     Git hook（シェルスクリプト）
│       └── post-merge   Git hook（シェルスクリプト）
├── template/            init展開用テンプレート
├── embed.go             VERSION、template、hookの埋め込み
└── VERSION

user data repository
├── .infra-version
├── .backup/            バックアップ（中身はGit追跡対象外）
├── sync.toml
└── <category>/
    └── link.toml
```

エンジンには個人データを含めない。`template/` は `dotfiles init` で展開する初期雛形だけを持つ。

## 境界と責務

### CLI（cmd/）

Cobraコマンド定義を `cmd/` にサブコマンドごとのファイルで配置する。
各 `RunE` はエラーを返し、プロセス終了は `cmd/main.go` だけが担当する。
`application` structが埋め込みリソースを保持し、`internal` パッケージのロジックへ委譲する。

### ビジネスロジック（internal/）

import path: `github.com/kokukaityo/dotfiles-bridge/internal`（package名 `engine`）。
コマンド定義を持たず、純粋なロジックだけを公開する。

### 設定

データリポジトリは `DOTFILES_DIR`、現在のGitルート、`~/dotfiles` の順に探索する。
`sync.toml` の `mode` はデフォルト `"local"`。`"remote"` を明示した場合のみ origin との同期を行う。
`sync.toml` の `default_branch` は空値を禁止し、`git check-ref-format --branch` で検証する。
`auto` / `ignore` のカテゴリ名は読み込み時に検証し、重複、相互衝突、パス区切り、絶対パス、内部予約名を拒否する。
エンジンとデータの `.infra-version` はメジャーバージョンを比較し、不一致時は警告する。

### Git

go-gitは使わず、ユーザーのSSH・credential・hook設定を利用するため `os/exec` でGit CLIを呼ぶ。
自動pushは `auto` カテゴリのpathspecだけをstage・commitする。既定ブランチ以外では実行しない。

pull時に分岐を検出した場合は自動mergeせず、ローカル履歴を `conflict/<hostname>/<timestamp>` へ退避する。

### symlink

各カテゴリの `link.toml` を読み、`runtime.GOOS` に対応するセクションだけを処理する。
リンク元はデータリポジトリ内、リンク先はユーザー環境に置く。既存のリンク先はデータリポジトリの `.backup/<category>/<timestamp>/` にバックアップしてから置換する。

### ファイル監視（watch）

`dotfiles watch` は fsnotify で `auto` カテゴリのディレクトリを再帰的に監視し、
変更を検知すると debounce（3秒）後に `Push()` を呼ぶ。
エディタの保存操作は複数のファイルシステムイベントを連打するため、即時pushではなく debounce で束ねる。

多重起動は PID ファイル（`.dotfiles-watch.pid`）で防ぐ。
Push 失敗（ネットワーク不通等）は stderr にログして監視を継続する。

### サービス登録（service）

`dotfiles install` の末尾で OS のログイン時自動起動を登録する。
`dotfiles watch` をバックグラウンドサービスとして起動させる。

- Linux: systemd user service（`~/.config/systemd/user/dotfiles-watch.service`）
- macOS: launchd user agent（`~/Library/LaunchAgents/com.dotfiles.watch.plist`）
- Windows: スタートアップフォルダに VBScript（コンソールウィンドウ非表示）

サービス登録失敗は install 全体を失敗させない（WARNING にとどめる）。
環境によっては systemd がなかったり権限が足りなかったりするため。

### hooks

hook本体は互換性のためシェルスクリプトを維持する。バイナリに埋め込み、install時にデータリポジトリの
`.dotfiles-hook/` へ書き出して `core.hooksPath` を設定する。このディレクトリはGit追跡対象外。

hookスクリプトは `internal/hook/` に配置し、ルートの `embed.go` から埋め込む。

### 埋め込み（embed.go）

ルート直下の `embed.go`（package dotfiles）が以下を公開する:

- `TemplateFS`: `template/` ディレクトリ全体
- `Version`: `VERSION` ファイルの内容
- `HookFS`: `internal/hook/` 配下のシェルスクリプト

`//go:embed` はソースファイルからの相対パスでしか参照できないため、`embed.go` はルートに置く必要がある。

## 設定形式

`sync.toml`:

```toml
mode = "local"
default_branch = "main"
auto = ["editor"]
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
