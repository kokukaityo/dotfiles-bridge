# シェル起動時の自動pull（任意）
#
# hookは「pull/merge後」と「push前」にしか走らない。
# 「セッション開始時に最新を取り込む」には、シェル設定かエージェントの起動フックを使う。
# 下記を ~/.bashrc / ~/.zshrc に追記すると、新しいシェルを開くたびに dotfiles を ff-only で更新する。
# --ff-only なので、端末間で分岐している場合は merge せず失敗し、手動解決を促す（正本を汚さない）。

dotfiles_sync() {
  local d="${DOTFILES:-$HOME/dotfiles}"
  [ -d "$d/.git" ] || return 0
  # ローカルに未コミット変更がある時は触らない（編集中を尊重）
  if [ -n "$(git -C "$d" status --porcelain)" ]; then
    echo "[dotfiles] ローカルに未コミット変更あり。自動pullをスキップ。"
    return 0
  fi
  git -C "$d" fetch --quiet origin
  if ! git -C "$d" merge --ff-only --quiet origin/main 2>/dev/null; then
    echo "[dotfiles] 分岐検出：自動pullを中止。'cd $d && git rebase origin/main' で解決してください。"
  fi
}

# 毎回うるさい場合は、1日1回などに絞ってもよい。
dotfiles_sync

# --- Claude Code の SessionStart hook で自動pullしたい場合（代替）---
# プロジェクト or グローバルの settings に SessionStart hook を追加し、
# 上記 dotfiles_sync 相当のスクリプトを呼ぶ。シェルに常駐させたくない場合はこちら。
