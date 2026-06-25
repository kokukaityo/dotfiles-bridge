# dotfiles-bridge

> **Alpha版です。** 実データへの適用前にバックアップを取ることを推奨します。
> フィードバックや不具合報告は [Issues](https://github.com/kokukaityo/dotfiles-bridge/issues) へお願いします。

本来アプリごとに閉じている設定ファイルを、アプリ間・マシン間で安全に橋渡しする Go 製の dotfiles 管理ツール。

CLI コマンド名は `dotfiles`。

## 特徴

- **一元管理** — 散らばった設定ファイルを一箇所に集約
- **OS 別配置** — symlink でアプリが期待するパスに自動配置（Windows / macOS / Linux）
- **バージョン管理** — Git による自動 commit・push・pull、分岐時のローカル変更退避
- **スコープの柔軟性** — 単一マシン内のアプリ間共有も、複数マシン間の同期も同じ仕組みで対応

## インストール

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/kokukaityo/dotfiles-bridge/main/install.sh | bash
```

インストール先を変更する場合:

```bash
INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/kokukaityo/dotfiles-bridge/main/install.sh | bash
```

### Windows

[GitHub Releases](https://github.com/kokukaityo/dotfiles-bridge/releases) から zip をダウンロードし、展開した `dotfiles.exe` を PATH の通ったディレクトリに配置してください。

### ソースからビルド

Go 1.26 以降と Git 2.x が必要です。

```bash
git clone https://github.com/kokukaityo/dotfiles-bridge.git
cd dotfiles-bridge
make build
# dist/dotfiles が生成される
```

## クイックスタート

### 新しく始める

```bash
# データリポジトリを作成
dotfiles init ~/dotfiles
cd ~/dotfiles

# リモートに接続（任意）
git remote add origin git@github.com:<user>/<repo>.git
git push -u origin main
```

### 別のマシンで使う

```bash
git clone git@github.com:<user>/<repo>.git ~/dotfiles
export DOTFILES_DIR="$HOME/dotfiles"
dotfiles install
```

### シェル起動時に自動同期（任意）

`~/.bashrc` や `~/.zshrc` に追加:

```bash
export DOTFILES_DIR="$HOME/dotfiles"
command -v dotfiles >/dev/null && dotfiles pull
command -v dotfiles >/dev/null && dotfiles status
```

`dotfiles install` を実行すると、ファイル監視サービスが OS のログイン時自動起動に登録されます。以降は設定ファイルの変更が自動で commit・push されます。

## コマンド

| コマンド                          | 説明                                       |
| --------------------------------- | ------------------------------------------ |
| `dotfiles init [path]`            | データリポジトリを作成（既定: `~/dotfiles`）|
| `dotfiles install`                | hooks・gitignore 設定と symlink 配置       |
| `dotfiles link`                   | OS に応じた symlink を配置                 |
| `dotfiles pull`                   | リモートから同期                           |
| `dotfiles push`                   | auto カテゴリの変更を commit して push     |
| `dotfiles watch`                  | ファイル変更を監視して自動 push            |
| `dotfiles delete-category <name>` | カテゴリを設定と Git 履歴から削除          |
| `dotfiles gitignore`              | `.gitignore` の自動生成部分を更新          |
| `dotfiles status`                 | コンフリクト退避状態を表示                 |
| `dotfiles version`                | バージョン情報を表示                       |

## データリポジトリ

`dotfiles init` で作成されるデータリポジトリの構成:

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

データリポジトリは次の順で解決されます:

1. 環境変数 `DOTFILES_DIR`
2. 現在の Git ルート（`.infra-version` がある場合）
3. `~/dotfiles`

### sync.toml

```toml
mode = "local"
default_branch = "main"
auto = ["ai-agent", "editor", "shell"]
ignore = ["backup", "raw"]
```

| キー             | 説明                                                                 |
| ---------------- | -------------------------------------------------------------------- |
| `mode`           | `"local"`（コミットのみ）または `"remote"`（origin との同期も行う）  |
| `default_branch` | pull・push・カテゴリ削除で使用するブランチ                           |
| `auto`           | `dotfiles push` の対象カテゴリ                                       |
| `ignore`         | `.gitignore` に追加するカテゴリ                                      |

どちらにも属さないカテゴリは manual 扱い（Git で追跡するが自動 push しない）。

カテゴリ名は英数字・`_`・`.`・`-` が使用でき、先頭は英数字または `_` にする必要があります。

### link.toml

OS キーの下に、リンク元（カテゴリ内のファイル）とリンク先の一覧を定義します。Windows の OS キーは `win32`。

```toml
[darwin]
"settings.json" = ["~/Library/Application Support/Code/User/settings.json"]

[linux]
"settings.json" = ["~/.config/Code/User/settings.json"]

[win32]
"settings.json" = ["~/AppData/Roaming/Code/User/settings.json"]
```

既存のリンク先はデータリポジトリの `.backup/<category>/<timestamp>/` に退避されます。

## コンフリクト

pull 時にローカルとリモートが分岐していた場合、ローカル側を `conflict/<hostname>/<timestamp>` ブランチへ退避し、既定ブランチをリモートへ戻します。

未解決状態では `.conflict-pending` が作成され、`dotfiles status` が警告を表示します。

## 既知の制約

- **Alpha 版です。** 実データへの適用前にバックアップを推奨します。
- 主に Windows で検証しています。macOS / Linux での動作は未検証です。
- watch サービスは `~/dotfiles` 前提の利用を推奨します。別パスでは手動起動か `DOTFILES_DIR` 環境変数の設定が必要です。
- link 先の親ディレクトリが存在しない場合、現状は自動作成します。この挙動は今後変更される可能性があります。
- CI・E2E テスト・installer の堅牢化は今後対応予定です。
- Windows で symlink を作成するには開発者モードの有効化が必要です（設定 → システム → 開発者向け）。

## 開発

開発に参加する場合は [doc/Development.md](doc/Development.md) を参照してください。
設計の詳細は [doc/Architecture.md](doc/Architecture.md) にあります。

## ライセンス

[MIT](LICENSE)
