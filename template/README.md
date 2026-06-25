# My Dotfiles

[dotfiles-bridge](https://github.com/kokukaityo/dotfiles-bridge) で管理する個人設定リポジトリ。

## 使い方

- 設定ファイルを追加: カテゴリディレクトリにファイルを置き、`link.toml` に symlink 定義を追加
- 同期: `dotfiles push` / `dotfiles pull`
- symlink 再配置: `dotfiles link`

## リポジトリ構成

```text
~/dotfiles/
├── sync.toml             # 同期設定
└── <category>/           # カテゴリディレクトリ（自由に追加）
    ├── link.toml         # symlink 定義
    └── (設定ファイル群)
```

カテゴリは用途ごとに自由に作れます（例: `shell/`, `vscode/`, `ai-agent/` など）。
カテゴリ名には英数字・`_`・`.`・`-` が使え、先頭は英数字または `_` にしてください。

## sync.toml

リポジトリ全体の同期モードとカテゴリの振り分けを定義します。

```toml
mode = "local"
default_branch = "main"
auto = ["ai-agent", "shell"]
ignore = ["backup", "raw", "vscode"]
```

> `vscode` カテゴリは VS Code に標準の Settings Sync 機能があるため、デフォルトでは `ignore`（Git 追跡外）です。dotfiles-bridge で Git 管理したい場合は `ignore` から外して `auto` に移動してください。

### キー

| キー             | 必須 | 既定値    | 説明                                                                |
| ---------------- | ---- | --------- | ------------------------------------------------------------------- |
| `mode`           | いいえ | `"local"` | `"local"`: コミットのみ。`"remote"`: origin との push/pull も行う |
| `default_branch` | はい | —         | pull・push・カテゴリ削除で使用するブランチ                          |
| `auto`           | いいえ | `[]`      | `dotfiles push` / `dotfiles watch` の自動 commit 対象カテゴリ      |
| `ignore`         | いいえ | `[]`      | `.gitignore` に追加され、Git 追跡から除外するカテゴリ              |

### カテゴリの同期モード

各カテゴリは `auto`・`ignore`・`manual` のいずれかに分類されます。

| モード   | 定義方法                          | 動作                                                   |
| -------- | --------------------------------- | ------------------------------------------------------ |
| `auto`   | `auto` に記載                     | `dotfiles push` で自動 commit・push される             |
| `ignore` | `ignore` に記載                   | `.gitignore` に追加され、Git で追跡しない              |
| `manual` | どちらにも記載しない              | Git で追跡するが、自動 push はしない（手動 commit 用） |

- 同じカテゴリを `auto` と `ignore` の両方に書くことはできません。
- カテゴリ名の重複もエラーになります。
- `dotfiles delete-category <name>` を実行すると、該当カテゴリが `auto`/`ignore` から自動で削除されます。

## link.toml

各カテゴリディレクトリに配置し、設定ファイルの symlink 先を OS ごとに定義します。

```toml
[darwin]
"settings.json" = ["~/Library/Application Support/Code/User/settings.json"]

[linux]
"settings.json" = ["~/.config/Code/User/settings.json"]

[win32]
"settings.json" = ["~/AppData/Roaming/Code/User/settings.json"]
```

### OS キー

| キー     | OS      | 由来                                         |
| -------- | ------- | -------------------------------------------- |
| `darwin` | macOS   | macOS の基盤カーネル名（Go の `runtime.GOOS`）|
| `linux`  | Linux   | Go の `runtime.GOOS`                         |
| `win32`  | Windows | Node.js の `process.platform` に準拠         |

現在の OS に該当するセクションだけが処理されます。他の OS のセクションは無視されるので、1つの `link.toml` に全 OS 分をまとめて書けます。

### 書式

```toml
[<os>]
"<カテゴリ内のファイルパス>" = ["<symlink 先のパス>"]
```

- キー（左辺）: カテゴリディレクトリからの相対パス
- 値（右辺）: symlink を配置するパスの配列。`~` はホームディレクトリに展開されます
- 1つのファイルに複数の symlink 先を指定できます（例: 同じ `settings.json` を VS Code と Cursor に配置）

```toml
[win32]
"settings.json" = [
    "~/AppData/Roaming/Code/User/settings.json",
    "~/AppData/Roaming/Cursor/User/settings.json",
]
```

### バックアップ

`dotfiles link` や `dotfiles install` を実行すると、symlink 先に既存のファイルがある場合は `.backup/<category>_<timestamp>/` に自動退避されます。既に同じリンクが張られている場合はスキップされます。

詳細は [dotfiles-bridge](https://github.com/kokukaityo/dotfiles-bridge) を参照。
