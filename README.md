# dotfile

複数マシン間で設定ファイルを同期する、Go製のdotfilesエンジン。

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

生成されるバイナリは `dist/dotfile`。

## クイックスタート

新しいデータリポジトリを作成する:

```bash
dotfile init ~/dotfiles
cd ~/dotfiles
git remote add origin git@github.com:<user>/<repo>.git
git push -u origin main
```

既存のデータリポジトリを使う:

```bash
git clone git@github.com:<user>/<repo>.git ~/dotfiles
export DOTFILES_DIR="$HOME/dotfiles"
dotfile setup
```

シェル起動時に同期する場合:

```bash
export DOTFILES_DIR="$HOME/dotfiles"
command -v dotfile >/dev/null && dotfile pull
command -v dotfile >/dev/null && dotfile status
```

## コマンド

| コマンド                         | 説明                                          |
| -------------------------------- | --------------------------------------------- |
| `dotfile init [path]`            | データリポジトリを作成。既定値は `~/dotfiles` |
| `dotfile setup`                  | hooks、gitignore、symlinkを設定               |
| `dotfile link`                   | OSに応じたsymlinkを配置                       |
| `dotfile pull`                   | リモートから同期                              |
| `dotfile push`                   | autoカテゴリの変更をcommitしてpush            |
| `dotfile delete-category <name>` | カテゴリを設定とGit履歴から削除               |
| `dotfile gitignore`              | `.gitignore` の自動生成部分を更新             |
| `dotfile status`                 | コンフリクト退避状態を表示                    |
| `dotfile version`                | バージョン情報を表示                          |

データリポジトリは次の順で解決する。

1. `DOTFILES_DIR`
2. 現在のGitルート（`.infra-version` がある場合）
3. `~/dotfiles`

## データリポジトリ

```text
~/dotfiles/
├── .infra-version
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
default_branch = "main"
auto = ["ai-agent", "editor", "shell"]
manual = []
ignore = ["backup", "raw"]
```

- `default_branch`: pull、push、カテゴリ削除で使用するブランチ
- `auto`: `dotfile push` の対象
- `manual`: Gitで追跡するが自動pushしないカテゴリ
- `ignore`: 自動生成される `.gitignore` に追加するカテゴリ

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
未解決状態では `.conflict-pending` が作成され、`dotfile status` が警告を表示する。

## ライセンス

MIT
