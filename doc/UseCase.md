# ユースケース

dotfiles-bridge の具体的な活用例です。各ユースケースで `link.toml` の記述例を示します。

カテゴリの追加手順やデータリポジトリの構成は [template/README.md](../template/README.md) を参照してください。複数マシン間の同期フローは [README.md のクイックスタート](../README.md#クイックスタート) を参照してください。

## AI エージェント設定の一元管理

Claude Code、Codex、Gemini CLI などの AI エージェント設定を、1箇所で管理して各ツールに配置します。

### 管理対象

- **AGENTS.md**: 全エージェント共通のグローバルルール（応答言語、コーディング規約など）
- **skills/**: 育てたスキル定義
- **commands/**: カスタムスラッシュコマンド
- **ai-setup-local.md**: リポジトリ別の AI 設定を生成するためのテンプレート。エージェントに読み込ませると、そのリポジトリに合った AGENTS.md と CLAUDE.md を生成できます

### link.toml の例

`dotfiles init` で生成されるテンプレートにはこの設定が含まれています。

```toml
[darwin]
"AGENTS.md" = ["~/.codex/AGENTS.md", "~/.claude/CLAUDE.md"]
"commands/" = ["~/.claude/commands/"]
"skills/" = ["~/.claude/skills/", "~/.codex/skills/"]

[win32]
"AGENTS.md" = ["~/.codex/AGENTS.md", "~/.claude/CLAUDE.md"]
"commands/" = ["~/.claude/commands/"]
"skills/" = ["~/.claude/skills/", "~/.codex/skills/"]
```

1つの `AGENTS.md` を Claude Code（`~/.claude/CLAUDE.md`）と Codex（`~/.codex/AGENTS.md`）の両方に配置できます。片方の PC で AGENTS.md を編集すれば、symlink 経由で両方のツールに即座に反映されます。

### このユースケースのポイント

- AI エージェントの設定は使い込むほど洗練されていきます。dotfiles-bridge で管理すれば「育てた」設定がどの端末でもすぐに使えます
- Git で管理するため、設定の変更履歴が残ります。「あの設定をいつ変えたか」「前の状態に戻したい」に対応できます

## VS Code 系エディタ間の設定共有

VS Code、Cursor、Windsurf、VSCodium など、同じ設定形式を使うエディタ間で設定を共有します。1つの設定ファイルを複数エディタに同時配置できるため、エディタを乗り換えたり併用したりする際に設定を個別に管理する必要がありません。

> **注意**: VS Code 単体の同期だけが目的なら、標準の [Settings Sync](https://code.visualstudio.com/docs/editor/settings-sync) で十分です。dotfiles-bridge が活きるのは、複数の VS Code 系エディタに同じ設定を配りたい場合とGit によって標準機能よりも自由度の高いバージョン管理が可能となります。

### 標準同期との併用に関する注意

dotfiles-bridge と VS Code の標準 Settings Sync を同時に使うと、互いに設定を上書きし合うコンフリクトが発生します。dotfiles-bridge の同期機能を使う場合は、VS Code 側の Settings Sync を無効にしてください。

### 初期状態

`dotfiles init` で生成される `vscode` カテゴリはデフォルトで `ignore`（Git 追跡外）です。dotfiles-bridge で管理する場合は、`sync.toml` の `ignore` から `vscode` を外し、`auto` に追加してください。

```toml
auto = ["ai-agent", "shell", "vscode"]
ignore = []
```

### link.toml の例

```toml
[darwin]
"settings.json" = [
    "~/Library/Application Support/Code/User/settings.json",
    "~/Library/Application Support/Cursor/User/settings.json",
]
"keybindings.json" = [
    "~/Library/Application Support/Code/User/keybindings.json",
    "~/Library/Application Support/Cursor/User/keybindings.json",
]

[win32]
"settings.json" = [
    "~/AppData/Roaming/Code/User/settings.json",
    "~/AppData/Roaming/Cursor/User/settings.json",
]
"keybindings.json" = [
    "~/AppData/Roaming/Code/User/keybindings.json",
    "~/AppData/Roaming/Cursor/User/keybindings.json",
]
```

## シェル設定の共有

`.bashrc`, `.zshrc`, `.profile` などのシェル設定を OS 別に管理します。

### link.toml の例

```toml
[darwin]
".zshrc" = ["~/.zshrc"]
".zprofile" = ["~/.zprofile"]

[linux]
".bashrc" = ["~/.bashrc"]
".profile" = ["~/.profile"]
```

macOS では zsh、Linux では bash がデフォルトシェルのため、OS ごとに異なるファイルを配置できます。

## 単一マシン内のアプリ間共有

dotfiles-bridge はリモート同期ツールではありません。1台の PC 内で、異なるアプリ間の設定を橋渡しするだけでも価値があります。

`sync.toml` の `mode` はデフォルト `"local"` です。この状態ではリモートリポジトリへの push / pull は行わず、ローカルの Git コミットのみで動作します。

### 例: AGENTS.md をローカルの複数ツールに配置

```toml
[win32]
"AGENTS.md" = ["~/.codex/AGENTS.md", "~/.claude/CLAUDE.md"]
```

リモートリポジトリを一切使わずに、1つのファイルを複数のアプリに同時配置できます。設定の変更履歴もローカルの Git リポジトリに残ります。
