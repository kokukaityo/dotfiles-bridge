#!/usr/bin/env bash
# .infra/setup.sh — 新端末の初期セットアップ。
# symlink 配置 + Git hook 設置 + .gitignore 生成を一括で行う。
set -euo pipefail

source "$(cd "$(dirname "$0")" && pwd)/env.sh"

cd "$DOTFILE"

echo "=== dotfiles setup ==="
echo ""

# 1. Git hooks
git config core.hooksPath .infra/hook
chmod +x .infra/hook/* 2>/dev/null || true
echo "[setup] Git hooks configured (core.hooksPath=.infra/hook)"

# 2. symlink をそのまま保存する設定
if ! grep -q '^\* -text' .gitattributes 2>/dev/null; then
    echo '* -text' >> .gitattributes
fi
git config core.symlinks true
echo "[setup] .gitattributes configured for symlinks"

# 3. .gitignore 生成
bash "$INFRA/sync.sh" gitignore

# 4. symlink 配置
bash "$INFRA/link.sh"

echo ""
echo "=== Setup complete ==="
echo ""
echo "シェル起動時の自動同期を設定する場合、~/.bashrc or ~/.zshrc に以下を追加:"
echo '  [ -f ~/dotfiles/.infra/sync.sh ] && bash ~/dotfiles/.infra/sync.sh pull'
echo '  [ -f ~/dotfiles/.infra/sync.sh ] && bash ~/dotfiles/.infra/sync.sh status'
