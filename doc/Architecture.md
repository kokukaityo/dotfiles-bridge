# アーキテクチャ

## 全体構成

`dotfile` はGoのシングルバイナリとして動作し、ユーザー固有データは別のGitリポジトリで管理する。

```text
dotfile engine
├── cmd/main.go          CLIエントリポイントと終了コード
├── embed.go             VERSION、template、hooksの埋め込み
└── internal/engine/
    ├── command.go       Cobraコマンド
    ├── config.go        リポジトリ解決とTOML設定
    ├── git.go           gitコマンド実行
    ├── link.go          symlink配置
    ├── setup.go         initとsetup
    ├── sync.go          pull、push、削除、gitignore、status
    └── platform.go      OS名とホーム展開

user data repository
├── .infra-version
├── sync.toml
└── <category>/
    └── link.toml
```

エンジンには個人データを含めない。`template/` は `dotfile init` で展開する初期雛形だけを持つ。

## 境界と責務

### CLI

Cobraは引数検証と出力先を管理する。各 `RunE` はエラーを返し、プロセス終了は `cmd/main.go` だけが担当する。

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
