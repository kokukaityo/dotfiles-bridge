#!/usr/bin/env bash
# .infra/sync.sh — dotfiles の同期エンジン。
# サブコマンド: pull / push / gitignore
set -euo pipefail

DOTFILES="${DOTFILES:-$HOME/dotfiles}"
INFRA="$DOTFILES/.infra"
SYNC_YAML="$INFRA/sync.yaml"

# --- sync.yaml パーサー ---

# parse_sync_yaml: "category\tmode" 形式で出力
parse_sync_yaml() {
    local current_category=""
    while IFS= read -r line || [ -n "$line" ]; do
        [[ -z "$line" || "$line" =~ ^[[:space:]]*$ ]] && continue
        [[ "$line" =~ ^[[:space:]]*# ]] && continue

        if [[ "$line" =~ ^[a-zA-Z._] ]]; then
            current_category="${line%%:*}"
            continue
        fi

        if [[ "$line" =~ sync:[[:space:]]*([a-z]+) ]]; then
            echo "${current_category}	${BASH_REMATCH[1]}"
        fi
    done < "$SYNC_YAML"
}

get_categories_by_mode() {
    local mode="$1"
    parse_sync_yaml | while IFS=$'\t' read -r cat m; do
        [ "$m" = "$mode" ] && echo "$cat"
    done
}

# --- commit message 自動生成 ---

generate_commit_msg() {
    local added modified deleted
    added="$(git diff --cached --diff-filter=A --name-only 2>/dev/null || true)"
    modified="$(git diff --cached --diff-filter=M --name-only 2>/dev/null || true)"
    deleted="$(git diff --cached --diff-filter=D --name-only 2>/dev/null || true)"

    local parts=()

    if [ -n "$added" ]; then
        local names
        names="$(echo "$added" | xargs -I{} basename {} | sort -u | paste -sd ", ")"
        parts+=("add: $names")
    fi
    if [ -n "$modified" ]; then
        local names
        names="$(echo "$modified" | xargs -I{} basename {} | sort -u | paste -sd ", ")"
        parts+=("update: $names")
    fi
    if [ -n "$deleted" ]; then
        local names
        names="$(echo "$deleted" | xargs -I{} basename {} | sort -u | paste -sd ", ")"
        parts+=("delete: $names")
    fi

    if [ ${#parts[@]} -eq 0 ]; then
        echo "sync: no changes"
        return
    fi

    local IFS="; "
    echo "${parts[*]}"
}

# --- .gitignore 生成 ---

cmd_gitignore() {
    local gitignore="$DOTFILES/.gitignore"
    local marker_start="# --- auto-generated from sync.yaml (do not edit below) ---"
    local marker_end="# --- end auto-generated ---"

    # 手動セクションを保持
    local manual_section=""
    if [ -f "$gitignore" ]; then
        manual_section="$(sed "/$marker_start/,\$d" "$gitignore" 2>/dev/null || true)"
    fi

    {
        if [ -n "$manual_section" ]; then
            echo "$manual_section"
        fi
        echo "$marker_start"
        echo ""
        echo "# Security exclusions"
        echo "*auth*"
        echo "*.key"
        echo "*.pem"
        echo ".env*"
        echo ""
        echo "# Ignored categories (sync: ignore in sync.yaml)"
        for cat in $(get_categories_by_mode "ignore"); do
            echo "${cat}/"
        done
        echo ""
        echo ".conflict-pending"
        echo "$marker_end"
    } > "$gitignore"

    echo "[sync] .gitignore generated."
}

# --- pull ---

cmd_pull() {
    cd "$DOTFILES"

    # conflict branch が全て削除済みなら .conflict-pending を掃除
    if [ -f .conflict-pending ]; then
        if ! git branch | grep -q 'conflict/'; then
            rm -f .conflict-pending
            echo "[sync] Conflict resolved. Removed .conflict-pending."
        fi
    fi

    git fetch --quiet origin

    local local_head remote_head merge_base
    local_head="$(git rev-parse HEAD)"
    remote_head="$(git rev-parse --verify origin/main 2>/dev/null || echo "")"

    if [ -z "$remote_head" ]; then
        echo "[sync] No remote main branch found. Skipping pull."
        return 0
    fi

    if [ "$local_head" = "$remote_head" ]; then
        echo "[sync] Already up to date."
        return 0
    fi

    merge_base="$(git merge-base HEAD origin/main)"

    if [ "$local_head" = "$merge_base" ]; then
        git merge --ff-only origin/main
        echo "[sync] Fast-forwarded to origin/main."
        return 0
    fi

    # 分岐検知 → conflict branch に退避
    local hostname_str
    hostname_str="$(hostname)"
    local timestamp
    timestamp="$(date +%Y%m%d-%H%M%S)"
    local conflict_branch="conflict/${hostname_str}/${timestamp}"

    echo "[sync] Divergence detected. Creating conflict branch: $conflict_branch"

    git checkout -b "$conflict_branch"

    if [ -n "$(git status --porcelain)" ]; then
        git add -A
        git commit -m "auto-save: uncommitted changes before conflict resolution"
    fi

    git checkout main
    git reset --hard origin/main

    touch .conflict-pending

    echo ""
    echo "========================================"
    echo "  CONFLICT DETECTED"
    echo "  Local changes saved to: $conflict_branch"
    echo "  main has been reset to origin/main."
    echo "  See README.md for resolution steps."
    echo "========================================"
    echo ""
}

# --- push ---

cmd_push() {
    cd "$DOTFILES"

    local current_branch
    current_branch="$(git branch --show-current)"
    if [ "$current_branch" != "main" ]; then
        echo "[sync] Not on main branch ($current_branch). Skipping auto-push."
        return 0
    fi

    local has_changes=false
    for cat in $(get_categories_by_mode "auto"); do
        if [ -d "$DOTFILES/$cat" ]; then
            git add "$DOTFILES/$cat/"
        fi
    done

    if [ -z "$(git diff --cached --name-only)" ]; then
        echo "[sync] No changes in auto-sync categories."
        return 0
    fi

    local msg
    msg="$(generate_commit_msg)"
    git commit -m "$msg"
    git push origin main
    echo "[sync] Pushed: $msg"
}

# --- conflict 通知バナー（シェル起動時に呼ぶ用） ---

cmd_status() {
    if [ -f "$DOTFILES/.conflict-pending" ]; then
        echo ""
        echo "========================================"
        echo "  [dotfiles] CONFLICT PENDING"
        echo "  Run: cd $DOTFILES && git log --oneline --graph --all"
        echo "  See README.md for resolution steps."
        echo "========================================"
        echo ""
    fi
}

# --- main ---

case "${1:-help}" in
    pull)      cmd_pull ;;
    push)      cmd_push ;;
    gitignore) cmd_gitignore ;;
    status)    cmd_status ;;
    *)
        echo "Usage: sync.sh {pull|push|gitignore|status}"
        echo "  pull      — fetch and merge (or create conflict branch)"
        echo "  push      — auto-commit and push (auto categories only)"
        echo "  gitignore — regenerate .gitignore from sync.yaml"
        echo "  status    — show conflict warning if pending"
        exit 1
        ;;
esac
