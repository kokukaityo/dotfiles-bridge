#!/usr/bin/env bash
# lib/link.sh — 各カテゴリの link.yaml を読み、OS に応じた symlink を配置する。
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/conf.sh"

detect_os() {
    case "$OSTYPE" in
        darwin*)          echo "darwin" ;;
        msys*|cygwin*|mingw*) echo "win32" ;;
        linux*)           echo "linux" ;;
        *)                echo "unknown" ;;
    esac
}

link() {
    local src="$1" dst="$2"
    mkdir -p "$(dirname "$dst")"
    if [ -e "$dst" ] || [ -L "$dst" ]; then
        if [ -L "$dst" ] && [ "$(readlink "$dst")" = "$src" ]; then
            echo "  ok (already linked): $dst"
            return
        fi
        mv "$dst" "$dst.bak.$(date +%Y%m%d%H%M%S)"
        echo "  backed up: $dst -> $dst.bak.*"
    fi
    ln -s "$src" "$dst"
    echo "  linked: $dst -> $src"
}

# parse_link_yaml <link.yaml path> <os_key>
parse_link_yaml() {
    local yaml_file="$1"
    local os_key="$2"
    local state="outside"
    local current_source=""

    while IFS= read -r line || [ -n "$line" ]; do
        if [[ -z "$line" || "$line" =~ ^[[:space:]]*$ ]]; then
            if [ "$state" = "in_os" ] || [ "$state" = "in_source" ]; then
                state="outside"
            fi
            continue
        fi

        [[ "$line" =~ ^[[:space:]]*# ]] && continue

        if [[ "$line" =~ ^[a-zA-Z_] ]]; then
            local key="${line%%:*}"
            if [ "$key" = "$os_key" ]; then
                state="in_os"
            else
                state="skip_os"
            fi
            continue
        fi

        case "$state" in
            in_os|in_source)
                if [[ "$line" =~ ^[[:space:]]{4}[a-zA-Z_] ]]; then
                    current_source="$(echo "$line" | sed 's/^[[:space:]]*//' | sed 's/:$//')"
                    state="in_source"
                    continue
                fi
                if [[ "$line" =~ ^[[:space:]]{8}-[[:space:]] ]]; then
                    local target
                    target="$(echo "$line" | sed 's/^[[:space:]]*-[[:space:]]*//')"
                    echo "${current_source}	${target}"
                    continue
                fi
                ;;
            skip_os)
                continue
                ;;
        esac
    done < "$yaml_file"
}

process_category() {
    local category_dir="$1"
    local os_key="$2"
    local yaml_file="$category_dir/link.yaml"
    local category_name
    category_name="$(basename "$category_dir")"

    if [ ! -f "$yaml_file" ]; then
        return
    fi

    if ! grep -qE '^[a-zA-Z]' "$yaml_file" 2>/dev/null; then
        return
    fi

    echo "[$category_name]"

    while IFS=$'\t' read -r source_key target; do
        [ -z "$source_key" ] && continue

        local src_path="${category_dir}/${source_key%/}"

        target="${target/#\~/$HOME}"
        target="${target%/}"

        if [ ! -e "$src_path" ]; then
            echo "  skip (source not found): $src_path"
            continue
        fi

        link "$src_path" "$target"
    done < <(parse_link_yaml "$yaml_file" "$os_key")

    echo ""
}

# --- main ---
main() {
    local os_key
    os_key="$(detect_os)"

    if [ "$os_key" = "unknown" ]; then
        echo "Error: unsupported OS (OSTYPE=$OSTYPE)" >&2
        exit 1
    fi

    echo "=== dotfiles link ($os_key) ==="
    echo ""

    for yaml in "$DOTFILES"/*/link.yaml; do
        [ -f "$yaml" ] || continue
        process_category "$(dirname "$yaml")" "$os_key"
    done

    echo "Done."
    if [ "$os_key" = "win32" ]; then
        echo ""
        echo "Windows: 'Operation not permitted' が出る場合は開発者モードを有効化するか、"
        echo "管理者権限でターミナルを実行してください。"
    fi
}

main "$@"
