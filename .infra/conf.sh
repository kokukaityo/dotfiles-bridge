#!/usr/bin/env bash
# dotfiles インフラ共通設定 — source して使う

INFRA="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILE="$INFRA/.."

SYNC_AUTO=(ai-agent editor shell)
SYNC_MANUAL=(.infra)
SYNC_IGNORE=(backup raw)