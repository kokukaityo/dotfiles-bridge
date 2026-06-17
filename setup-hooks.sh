#!/usr/bin/env bash
# setup-hooks.sh — dotfiles リポジトリに Git hook を設置する。
# core.hooksPath を使い、リポジトリ内の hooks/ をそのまま hook ディレクトリにする。
# これにより hook 自体も Git 管理・端末間共有できる。

set -euo pipefail
DOTFILES="${DOTFILES:-$HOME/dotfiles}"
cd "$DOTFILES"

git config core.hooksPath hooks
chmod +x hooks/* 2>/dev/null || true

# symlinkを実体展開させない（端末間でリンクが壊れるのを防ぐ）
if ! grep -q "symlink" .gitattributes 2>/dev/null; then
  echo "* -text" >> .gitattributes
  echo "# symlinks are stored as symlinks, not materialized" >> .gitattributes
fi
git config core.symlinks true

echo "hooks installed (core.hooksPath=hooks)."
echo "post-merge: pull後にsymlinkを張り直す / pre-push: 不正状態でpushを止める"
