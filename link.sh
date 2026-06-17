#!/usr/bin/env bash
# link.sh — dotfiles の正本を各AIツールの設定パスへ symlink 配置する。
# 既存ファイルは *.bak に退避してからリンクを張る。
# OS分岐：Unix系は ln -s。Windowsは Git Bash 環境なら MSYS の ln が使えるが、
#         権限不足時は管理者権限/開発者モードが必要（後述のフォールバック参照）。

set -euo pipefail

DOTFILES="${DOTFILES:-$HOME/dotfiles}"
AI="$DOTFILES/ai"

link() {
  # link <実体> <リンク先>
  local src="$1" dst="$2"
  mkdir -p "$(dirname "$dst")"
  if [ -e "$dst" ] || [ -L "$dst" ]; then
    if [ -L "$dst" ] && [ "$(readlink "$dst")" = "$src" ]; then
      echo "ok (already linked): $dst"
      return
    fi
    mv "$dst" "$dst.bak.$(date +%Y%m%d%H%M%S)"
    echo "backed up: $dst -> $dst.bak.*"
  fi
  ln -s "$src" "$dst"
  echo "linked: $dst -> $src"
}

# --- 単一ファイル ---
link "$AI/AGENTS.md" "$HOME/.codex/AGENTS.md"
link "$AI/AGENTS.md" "$HOME/.claude/CLAUDE.md"

# --- commands ディレクトリ（中身をまとめてリンク） ---
if [ -d "$AI/commands" ]; then
  mkdir -p "$HOME/.claude/commands"
  for f in "$AI/commands"/*; do
    [ -e "$f" ] || continue
    link "$f" "$HOME/.claude/commands/$(basename "$f")"
  done
fi

# --- skills ディレクトリ（スキル単位でリンク） ---
if [ -d "$AI/skills" ]; then
  mkdir -p "$HOME/.agents/skills"
  for d in "$AI/skills"/*/; do
    [ -d "$d" ] || continue
    link "${d%/}" "$HOME/.agents/skills/$(basename "$d")"
  done
fi

echo
echo "Done. もし 'Operation not permitted' が出る場合（Windows）："
echo "  - VSCode/ターミナルを管理者権限で実行する、または Windows の開発者モードを有効化する。"
echo "  - それも不可な環境では、link.sh の ln を 'cp -r' に置き換えてコピー配置に切り替える"
echo "    （その場合は編集→正本へ反映の運用に注意。symlinkの一元性は失われる）。"
