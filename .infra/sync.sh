#!/usr/bin/env bash
# .infra/sync.sh — dotfiles の同期エンジン。
# サブコマンド: pull / push / delete-category / gitignore
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/conf.sh"

get_categories_by_mode() {
    case "$1" in
        auto)   printf '%s\n' "${SYNC_AUTO[@]}" ;;
        manual) printf '%s\n' "${SYNC_MANUAL[@]}" ;;
        ignore) printf '%s\n' "${SYNC_IGNORE[@]}" ;;
        *)
            echo "Unknown sync mode: $1" >&2
            return 1
            ;;
    esac
}

# --- commit message 自動生成 ---

generate_commit_msg() {
    local added modified deleted
    added="$(git diff --cached --diff-filter=A --name-only -- "$@" 2>/dev/null || true)"
    modified="$(git diff --cached --diff-filter=M --name-only -- "$@" 2>/dev/null || true)"
    deleted="$(git diff --cached --diff-filter=D --name-only -- "$@" 2>/dev/null || true)"

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
    local gitignore="$DOTFILE/.gitignore"
    local marker_start="# --- auto-generated from conf.sh (do not edit below) ---"
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
        echo "# Ignored categories (SYNC_IGNORE in conf.sh)"
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
    cd "$DOTFILE"

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

    if [ "$remote_head" = "$merge_base" ]; then
        echo "[sync] Local main is ahead of origin/main. Skipping pull."
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
    cd "$DOTFILE"

    local current_branch
    current_branch="$(git branch --show-current)"
    if [ "$current_branch" != "main" ]; then
        echo "[sync] Not on main branch ($current_branch). Skipping auto-push."
        return 0
    fi

    local auto_paths=()
    local missing_categories=()
    for cat in $(get_categories_by_mode "auto"); do
        if [ -d "$DOTFILE/$cat" ]; then
            auto_paths+=("$cat")
        elif { [ -n "$(git ls-files -- "$cat")" ] ||
            [ -n "$(git ls-tree -r --name-only HEAD -- "$cat")" ]; }; then
            missing_categories+=("$cat")
        fi
    done

    if [ ${#missing_categories[@]} -gt 0 ]; then
        echo "[sync] WARNING: tracked auto-sync categories are missing; skipping them:" >&2
        printf '  - %s\n' "${missing_categories[@]}" >&2
        echo "[sync] Restore an accidental deletion with:" >&2
        printf '  git restore -- %q\n' "${missing_categories[@]}" >&2
        echo "[sync] Or permanently remove one category with:" >&2
        echo "  bash .infra/sync.sh delete-category <category>" >&2
    fi

    for cat in "${auto_paths[@]}"; do
        git add -- "$cat/"
    done

    if [ ${#auto_paths[@]} -eq 0 ] ||
        [ -z "$(git diff --cached --name-only -- "${auto_paths[@]}")" ]; then
        echo "[sync] No changes in auto-sync categories."
        return 0
    fi

    local msg
    msg="$(generate_commit_msg "${auto_paths[@]}")"
    git commit --only -m "$msg" -- "${auto_paths[@]}"
    git push origin main
    echo "[sync] Pushed: $msg"
}

# --- category deletion ---

cmd_delete_category() {
    if [ "$#" -ne 1 ]; then
        echo "Usage: sync.sh delete-category <category>" >&2
        return 1
    fi

    local category="$1"
    case "$category" in
        ""|"."|".."|*/*|*\\*)
            echo "[sync] Invalid category name: $category" >&2
            return 1
            ;;
    esac
    if [[ ! "$category" =~ ^[[:alnum:]_][[:alnum:]_.-]*$ ]]; then
        echo "[sync] Invalid category name: $category" >&2
        return 1
    fi

    cd "$DOTFILE"

    local current_branch
    current_branch="$(git branch --show-current)"
    if [ "$current_branch" != "main" ]; then
        echo "[sync] Category deletion is only allowed on main (current: $current_branch)." >&2
        return 1
    fi

    local is_auto=false
    local cat
    for cat in "${SYNC_AUTO[@]}"; do
        if [ "$cat" = "$category" ]; then
            is_auto=true
            break
        fi
    done
    if [ "$is_auto" != "true" ]; then
        echo "[sync] Category is not in SYNC_AUTO: $category" >&2
        return 1
    fi

    if ! git diff --quiet -- .infra/conf.sh ||
        ! git diff --cached --quiet -- .infra/conf.sh; then
        echo "[sync] BLOCKED: .infra/conf.sh already has changes." >&2
        echo "[sync] Commit or restore them before deleting a category." >&2
        return 1
    fi

    local had_tracked_in_head=false
    if [ -n "$(git ls-tree -r --name-only HEAD -- "$category")" ]; then
        had_tracked_in_head=true
    fi

    local new_auto=()
    for cat in "${SYNC_AUTO[@]}"; do
        if [ "$cat" != "$category" ]; then
            new_auto+=("$cat")
        fi
    done

    local tmp_conf
    tmp_conf="$(mktemp "${TMPDIR:-/tmp}/dotfiles-conf.XXXXXX")"
    awk -v replacement="SYNC_AUTO=(${new_auto[*]})" '
        /^SYNC_AUTO=\(/ {
            print replacement
            replaced = 1
            next
        }
        { print }
        END {
            if (!replaced) {
                exit 1
            }
        }
    ' .infra/conf.sh > "$tmp_conf" || {
        rm -f -- "$tmp_conf"
        echo "[sync] Could not update SYNC_AUTO in .infra/conf.sh." >&2
        return 1
    }

    git reset -q HEAD -- "$category"
    rm -rf -- "$DOTFILE/$category"
    mv -- "$tmp_conf" .infra/conf.sh

    local commit_paths=(.infra/conf.sh)
    git add -- .infra/conf.sh
    if [ "$had_tracked_in_head" = "true" ]; then
        git add -A -- "$category"
        commit_paths+=("$category")
    fi

    local msg="delete: category $category"
    git commit --only -m "$msg" -- "${commit_paths[@]}"

    if ! git push origin main; then
        echo "[sync] Push failed. The deletion commit remains in local main." >&2
        return 1
    fi
    echo "[sync] Deleted category and pushed: $category"
}

# --- conflict 通知バナー（シェル起動時に呼ぶ用） ---

cmd_status() {
    if [ -f "$DOTFILE/.conflict-pending" ]; then
        echo ""
        echo "========================================"
        echo "  [dotfiles] CONFLICT PENDING"
        echo "  Run: cd $DOTFILE && git log --oneline --graph --all"
        echo "  See README.md for resolution steps."
        echo "========================================"
        echo ""
    fi
}

# --- main ---

case "${1:-help}" in
    pull)      cmd_pull ;;
    push)      cmd_push ;;
    delete-category)
        shift
        cmd_delete_category "$@"
        ;;
    gitignore) cmd_gitignore ;;
    status)    cmd_status ;;
    *)
        echo "Usage: sync.sh {pull|push|delete-category|gitignore|status}"
        echo "  pull      — fetch and merge (or create conflict branch)"
        echo "  push      — auto-commit and push (auto categories only)"
        echo "  delete-category <category>"
        echo "            — permanently delete an auto category and push"
        echo "  gitignore — regenerate .gitignore from conf.sh"
        echo "  status    — show conflict warning if pending"
        exit 1
        ;;
esac
