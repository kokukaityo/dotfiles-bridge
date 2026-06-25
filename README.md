# dotfiles-bridge

複数マシン間で設定ファイルを同期する、Go製のdotfiles管理ツール。
CLI コマンド名は `dotfiles`。

- カテゴリ単位の同期モード管理
- Windows / macOS / Linux別のsymlink配置
- Gitによる自動commit・push・pull
- 分岐時のローカル変更退避
- push前のコンフリクト・シークレット検査

## 前提条件

| 環境    | 要件                                                                |
| ------- | ------------------------------------------------------------------- |
| 共通    | Go 1.26以降、Git 2.x                                                |
| Windows | 開発者モードの有効化（symlink作成に必要。設定→システム→開発者向け） |

## ビルド

```bash
make build
```

生成されるバイナリは `dist/dotfiles`。

## クイックスタート

新しいデータリポジトリを作成する:

```bash
dotfiles init ~/dotfiles
cd ~/dotfiles
git remote add origin git@github.com:<user>/<repo>.git
git push -u origin main
```

既存のデータリポジトリを使う:

```bash
git clone git@github.com:<user>/<repo>.git ~/dotfiles
export DOTFILES_DIR="$HOME/dotfiles"
dotfiles install
```

シェル起動時に同期する場合:

```bash
export DOTFILES_DIR="$HOME/dotfiles"
command -v dotfiles >/dev/null && dotfiles pull
command -v dotfiles >/dev/null && dotfiles status
```

`dotfiles install` を実行すると、ファイル監視サービスが OS のログイン時自動起動に登録される。以降は `dotfiles push` を手動で叩く必要はない。手動で監視プロセスを起動する場合は `dotfiles watch` を使う。

## コマンド

| コマンド                           | 説明                                          |
| ---------------------------------- | --------------------------------------------- |
| `dotfiles init [path]`             | データリポジトリを作成。既定値は `~/dotfiles` |
| `dotfiles install`                 | hooks、gitignore設定とsymlink配置             |
| `dotfiles link`                    | OSに応じたsymlinkを配置                       |
| `dotfiles pull`                    | リモートから同期                              |
| `dotfiles push`                    | autoカテゴリの変更をcommitしてpush            |
| `dotfiles watch`                   | ファイル変更を監視して自動push                |
| `dotfiles delete-category <name>`  | カテゴリを設定とGit履歴から削除               |
| `dotfiles gitignore`               | `.gitignore` の自動生成部分を更新             |
| `dotfiles status`                  | コンフリクト退避状態を表示                    |
| `dotfiles version`                 | バージョン情報を表示                          |

データリポジトリは次の順で解決する。

1. `DOTFILES_DIR`
2. 現在のGitルート（`sync.toml` がある場合）
3. `~/dotfiles`

## データリポジトリ

```text
~/dotfiles/
├── sync.toml
├── ai-agent/
│   └── link.toml
├── editor/
│   └── link.toml
└── shell/
    └── link.toml
```

### sync.toml

```toml
mode = "local"
default_branch = "main"
auto = ["ai-agent", "editor", "shell"]
ignore = ["backup", "raw"]
```

- `mode`: `"local"`（デフォルト）または `"remote"`。local はコミットのみ、remote は origin との同期も行う
- `default_branch`: pull、push、カテゴリ削除で使用するブランチ
- `auto`: `dotfiles push` の対象
- `ignore`: 自動生成される `.gitignore` に追加するカテゴリ
- どちらにも属さないカテゴリは manual 扱い（Git で追跡するが自動 push しない）

`auto` / `ignore` のカテゴリ名は英数字、`_`、`.`、`-` のみ使用でき、先頭は英数字または `_` にする必要がある。
パス区切り、絶対パス、`.`、`..`、本体内部で使う予約名（`sync.toml` など）は使用できない。

### link.toml

OSキーの下に、カテゴリ内のリンク元とリンク先一覧を定義する。
WindowsのOSキーは `win32`。

```toml
[darwin]
"settings.json" = ["~/Library/Application Support/Code/User/settings.json"]

[linux]
"settings.json" = ["~/.config/Code/User/settings.json"]

[win32]
"settings.json" = ["~/AppData/Roaming/Code/User/settings.json"]
```

既存のリンク先はデータリポジトリの `.backup/<category>/<timestamp>/` に退避する。Windowsでsymlink作成に失敗する場合は、[前提条件](#前提条件)を確認する。

## コンフリクト

pull時にローカルとリモートが分岐していた場合、ローカル側を
`conflict/<hostname>/<timestamp>` ブランチへ退避し、既定ブランチをリモートへ戻す。
未解決状態では `.conflict-pending` が作成され、`dotfiles status` が警告を表示する。

## ライセンス

MIT
