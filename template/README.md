# My Dotfiles

[dotfiles-bridge](https://github.com/kokukaityo/dotfiles-bridge) で管理する個人設定リポジトリ。

## セットアップ

### 1. dotfiles-bridge をインストール

```bash
curl -fsSL https://raw.githubusercontent.com/kokukaityo/dotfiles-bridge/main/install.sh | bash
```

Windows の場合は [GitHub Releases](https://github.com/kokukaityo/dotfiles-bridge/releases) から zip をダウンロードしてください。

### 2. 初期設定

```bash
dotfiles install
```

### 3. シェル起動時の自動同期（任意）

`~/.bashrc` や `~/.zshrc` に追加:

```bash
export DOTFILES_DIR="$HOME/dotfiles"
command -v dotfiles >/dev/null && dotfiles pull
command -v dotfiles >/dev/null && dotfiles status
```

## 使い方

- 設定ファイルを追加: カテゴリディレクトリにファイルを置き、`link.toml` に symlink 定義を追加
- 同期: `dotfiles push` / `dotfiles pull`
- symlink 再配置: `dotfiles link`

## 構成

| ディレクトリ | 用途                |
| ------------ | ------------------- |
| `ai-agent/`  | AI エージェント設定 |
| `editor/`    | エディタ設定        |
| `shell/`     | シェル設定          |

| ファイル         | 用途                 |
| ---------------- | -------------------- |
| `sync.toml`      | 同期モード定義       |
| `.infra-version` | 互換本体バージョン   |

詳細は [dotfiles-bridge](https://github.com/kokukaityo/dotfiles-bridge) を参照。
