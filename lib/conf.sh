#!/usr/bin/env bash
# lib/conf.sh — エンジン共通設定ローダー。source して使う。

# エンジン自身のパス（ランチャーから export されていなければ BASH_SOURCE から導出）
if [ -z "${DOTFILES_ENGINE_LIB:-}" ]; then
    DOTFILES_ENGINE_LIB="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    DOTFILES_ENGINE_DIR="$(dirname "$DOTFILES_ENGINE_LIB")"
fi

resolve_dotfiles_dir() {
    if [ -n "${DOTFILES_DIR:-}" ] && [ -d "$DOTFILES_DIR" ]; then
        echo "$DOTFILES_DIR"
        return
    fi

    local git_root
    git_root="$(git rev-parse --show-toplevel 2>/dev/null || true)"
    if [ -n "$git_root" ] && [ -f "$git_root/.infra-version" ]; then
        echo "$git_root"
        return
    fi

    if [ -d "$HOME/dotfiles" ] && [ -f "$HOME/dotfiles/.infra-version" ]; then
        echo "$HOME/dotfiles"
        return
    fi

    echo "[dotfiles] Error: データリポジトリが見つかりません。" >&2
    echo "  DOTFILES_DIR 環境変数を設定するか、データリポジトリ内で実行してください。" >&2
    return 1
}

DOTFILES="$(resolve_dotfiles_dir)" || exit 1

# データリポジトリの sync.conf を読み込む
if [ -f "$DOTFILES/sync.conf" ]; then
    source "$DOTFILES/sync.conf"
else
    SYNC_AUTO=()
    SYNC_MANUAL=()
    SYNC_IGNORE=()
fi

check_version_compat() {
    local engine_version data_version
    engine_version="$(cat "$DOTFILES_ENGINE_DIR/VERSION" 2>/dev/null || echo "0.0.0")"
    data_version="$(cat "$DOTFILES/.infra-version" 2>/dev/null || echo "0.0.0")"

    local engine_major="${engine_version%%.*}"
    local data_major="${data_version%%.*}"

    if [ "$engine_major" != "$data_major" ]; then
        echo "[dotfiles] WARNING: バージョン不整合" >&2
        echo "  エンジン: v${engine_version}" >&2
        echo "  データ:   v${data_version}" >&2
        echo "  メジャーバージョンが異なります。" >&2
    fi
}
