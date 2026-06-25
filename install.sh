#!/usr/bin/env bash
set -euo pipefail

REPO="kokukaityo/dotfiles-bridge"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)        ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "未対応のアーキテクチャ: $ARCH" >&2; exit 1 ;;
esac

case "$OS" in mingw*|msys*)
  echo "Windows では GitHub Releases から zip を直接ダウンロードしてください:" >&2
  echo "  https://github.com/${REPO}/releases" >&2
  exit 1
  ;; esac

VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v(.*)".*/\1/')

if [ -z "$VERSION" ]; then
  echo "最新バージョンの取得に失敗しました" >&2
  exit 1
fi

ARCHIVE="dotfiles_${VERSION}_${OS}_${ARCH}.tar.gz"
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "dotfiles v${VERSION} をダウンロード中..."
curl -fsSL -o "${TMPDIR}/${ARCHIVE}" "https://github.com/${REPO}/releases/download/v${VERSION}/${ARCHIVE}"

echo "展開中..."
tar xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR" dotfiles

echo "${INSTALL_DIR} にインストール中..."
install -m 755 "${TMPDIR}/dotfiles" "${INSTALL_DIR}/dotfiles"

echo "dotfiles v${VERSION} のインストールが完了しました"
dotfiles version 2>/dev/null || true
