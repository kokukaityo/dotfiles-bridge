# dotfiles

複数マシン間で dotfiles（設定ファイル）を同期するためのエンジン。

- **カテゴリ別管理**: `ai-agent/`, `editor/`, `shell/` などカテゴリごとに設定を分類
- **OS 別 symlink**: `link.yaml` で OS ごとの配置先を宣言的に定義
- **自動同期**: シェル起動時に pull、ファイル保存時に auto-commit & push
- **コンフリクト安全**: 分岐検知時は自動でブランチ退避、データを壊さない
- **セキュリティガード**: push 前にシークレット混入を検知

## インストール

```bash
git clone https://github.com/kokukaityo/dotfile.git ~/.local/share/dotfiles
```

PATH に追加（`~/.bashrc` or `~/.zshrc`）:

```bash
export PATH="$HOME/.local/share/dotfiles/bin:$PATH"
```

## クイックスタート

### 新規ユーザー

```bash
# データリポジトリを作成
dotfiles init ~/dotfiles

# GitHub に private リポジトリを作成後:
cd ~/dotfiles
git remote add origin git@github.com:<user>/<repo>.git
git push -u origin main
```

### 2台目以降のマシン

```bash
# エンジンをインストール（上記と同じ）
git clone https://github.com/kokukaityo/dotfile.git ~/.local/share/dotfiles

# 自分のデータリポジトリをクローン
git clone git@github.com:<user>/<repo>.git ~/dotfiles

# 初期設定
dotfiles setup
```

## 自動同期の設定

`~/.bashrc` or `~/.zshrc` に追加:

```bash
export DOTFILES_DIR="$HOME/dotfiles"
export PATH="$HOME/.local/share/dotfiles/bin:$PATH"
command -v dotfiles >/dev/null && dotfiles pull
command -v dotfiles >/dev/null && dotfiles status
```

## コマンド

| コマンド | 説明 |
|---|---|
| `dotfiles init [path]` | データリポジトリを新規作成（デフォルト: `~/dotfiles`） |
| `dotfiles setup` | hooks 設定・gitignore 生成・symlink 配置 |
| `dotfiles link` | symlink を配置 |
| `dotfiles pull` | リモートから同期 |
| `dotfiles push` | auto カテゴリの変更を commit & push |
| `dotfiles delete-category <name>` | カテゴリを削除 |
| `dotfiles gitignore` | .gitignore を再生成 |
| `dotfiles status` | コンフリクト状態を表示 |
| `dotfiles version` | バージョン情報を表示 |

## データリポジトリの構成

```
~/dotfiles/
├── ai-agent/           # AI エージェント設定
│   └── link.yaml       # symlink 定義
├── editor/             # エディタ設定
│   └── link.yaml
├── shell/              # シェル設定
│   └── link.yaml
├── sync.conf           # 同期モード定義
└── .infra-version      # 互換エンジンバージョン
```

### sync.conf

カテゴリの同期モードを定義する:

```bash
SYNC_AUTO=(ai-agent editor shell)   # 自動 commit & push
SYNC_MANUAL=()                       # git 追跡のみ、push は手動
SYNC_IGNORE=(backup raw)            # git 追跡しない
```

### link.yaml

OS ごとに symlink の配置先を定義する:

```yaml
darwin:
    .zshrc:
        - ~/.zshrc
    settings.json:
        - ~/Library/Application Support/Code/User/settings.json

linux:
    .bashrc:
        - ~/.bashrc
    settings.json:
        - ~/.config/Code/User/settings.json
```

## コンフリクト解消

同期時に分岐（divergence）を検知すると、ローカルの変更を `conflict/<hostname>/<timestamp>` ブランチに退避し、main をリモートに合わせる。

解消手順:

```bash
cd ~/dotfiles
git log --oneline --graph --all     # 状態を確認
git cherry-pick <commit>            # 必要な変更を取り込む
git branch -d conflict/...          # 退避ブランチを削除
```

## ライセンス

MIT
