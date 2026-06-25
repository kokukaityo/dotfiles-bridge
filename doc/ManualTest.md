# dotfile エンジン 手動動作確認手順書

## Context

v0.1 の機能実装が完了したので、自分の PC（Windows）で全コマンドが正しく動作するか一通り確認する。コンフリクト処理やバックアップ挙動を含めた E2E レベルの検証。

## 構成

テストは段階的に進める：

1. **Phase 1 — ローカル完結**（リモート不要）: init, link, status, version, gitignore, watch
2. **Phase 2 — リモート追加**: push, pull, 2台目（clone → install）
3. **Phase 3 — 異常系**: コンフリクト発生〜解消、pre-push hook

Phase 2・3 はリモート同期を検証するため、開始前に `sync.toml` の `mode` を `"remote"` に変更する。

---

## 事前準備

### バックアップ

`ai-agent/link.toml` が以下のパスに symlink を張る。事前にバックアップ：

```bash
cp -r ~/.claude ~/.claude.bak
cp -r ~/.codex ~/.codex.bak
cp -r ~/.agents ~/.agents.bak
```

### 実行方法

WDAC 環境ではビルド済みバイナリの直接実行がブロックされるため、`make exe-<subcommand>` で `go run` 経由で実行する。

```bash
cd /d/work/dotfile/work

# 例: init
make exe-init ~/dotfiles-test

# 例: link
make exe-link

# デフォルトで DOTFILES_DIR=~/dotfiles-test が設定される。
# 別のパスを使う場合は明示指定：
# make exe-link DOTFILES_DIR=~/my-dotfiles
```

### Windows 開発者モード

設定 → システム → 開発者向け → 開発者モード ON（symlink 作成に必須）

---

## Phase 1 — ローカル完結（リモート不要）

### A. init

```bash
make exe-init ~/dotfiles-test
```

**確認：**

- [ ] `~/dotfiles-test/` が作成される
- [ ] `.infra-version` にエンジンバージョンが書かれている
- [ ] `sync.toml` が存在し、`mode`, `default_branch`, `auto`, `ignore` が設定されている
- [ ] `ai-agent/link.toml`, `editor/link.toml`, `shell/link.toml` が存在
- [ ] `.dotfile-hook/pre-push`, `.dotfile-hook/post-merge` が存在
- [ ] `.gitattributes` に `* -text` が書かれている
- [ ] `.gitignore` にマーカー行と自動生成セクションがある
- [ ] `git log` → `feat: initial dotfiles setup` の初回コミット
- [ ] `[dotfile] watchサービスを登録しました` と表示される（`go run` 経由のため VBScript のパスは無効だが、登録自体は成功する）

### B. link

```bash
make exe-link
```

**確認：**

- [ ] `~/.claude/CLAUDE.md` → `~/dotfiles-test/ai-agent/AGENTS.md` の symlink
- [ ] `~/.codex/AGENTS.md` → 同上
- [ ] `~/.claude/commands/` → `~/dotfiles-test/ai-agent/commands/` の symlink
- [ ] `~/.agents/skills/` → `~/dotfiles-test/ai-agent/skills/` の symlink
- [ ] 元のファイルが `.backup/ai-agent_<timestamp>/` に退避されている
- [ ] 2回目の `dotfile link` → `ok (already linked)` と表示（冪等）

### C. status

```bash
make exe-status
```

**確認：**

- [ ] コンフリクトがない状態 → `[status] No conflicts.` と表示

### D. version

```bash
make exe-version
```

**確認：**

- [ ] `dotfile engine v<version>` と表示される

### E. gitignore

```bash
# ~/dotfiles-test/sync.toml を編集して ignore に追加
# ignore = ["backup", "raw", "secret-stuff"]
make exe-gitignore
```

**確認：**

- [ ] `.gitignore` の自動生成部分に `secret-stuff/` が追加される
- [ ] マーカー上の手書き部分は保持されたまま

### F. watch（ファイル監視）

ターミナルを2つ使う。

```bash
# ターミナル1: watch を起動
make exe-watch
```

**確認：**

- [ ] `[watch] 自動同期カテゴリの監視を開始しました。` と表示

```bash
# ターミナル2: auto カテゴリのファイルを変更
echo "watch test" >> ~/dotfiles-test/ai-agent/AGENTS.md
```

**確認（約3秒後）：**

- [ ] debounce 後に自動 commit が実行される（local モードのため push はされない）
- [ ] `cd ~/dotfiles-test && git log --oneline -1` でコミットメッセージに `update: AGENTS.md` が含まれる

```bash
# ターミナル2: 多重起動防止を確認
make exe-watch
```

**確認：**

- [ ] `dotfile watchは既に稼働中です (pid=...)` とエラー表示

```bash
# ターミナル1: Ctrl+C で停止
```

**確認：**

- [ ] 正常終了し、`~/dotfiles-test/.dotfile-watch.pid` が削除されている

---

## Phase 2 — リモート追加

### 事前準備

Phase 1 完了後、リモート同期を有効にする：

```bash
# ~/dotfiles-test/sync.toml の mode を "remote" に変更
# mode = "remote"
```

### G. リモート接続 & push

```bash
# テスト用プライベートリポジトリ作成
gh repo create dotfiles-test --private

cd ~/dotfiles-test
git remote add origin git@github.com:kokukaityo/dotfiles-test.git
git push -u origin main
cd /d/work/dotfile/work

# ファイルを変更して push
echo "test change" >> ~/dotfiles-test/ai-agent/AGENTS.md
make exe-push
```

**確認：**

- [ ] auto カテゴリ（ai-agent）の変更が commit される
- [ ] コミットメッセージに `update: AGENTS.md` のような分類
- [ ] GitHub 上に反映される

### G-2. manual カテゴリは push されない

```bash
mkdir -p ~/dotfiles-test/experiment
echo "test" > ~/dotfiles-test/experiment/note.txt
make exe-push
```

**確認：**

- [ ] `experiment/` は commit されない（auto に含まれてないため）

### H. pull（別の場所から変更を push → 元で pull）

```bash
# 別ディレクトリに clone
cd ~
git clone git@github.com:kokukaityo/dotfiles-test.git dotfiles-test-clone

# clone 側で変更して push
cd ~/dotfiles-test-clone
echo "remote change" >> ai-agent/AGENTS.md
git add -A && git commit -m "remote update" && git push

# 元のリポジトリで pull
cd /d/work/dotfile/work
make exe-pull
```

**確認：**

- [ ] `[sync] Fast-forwarded to origin/main.` と表示
- [ ] ローカルに変更が反映されている

### I. 2台目セットアップ（clone → install）

```bash
make exe-install DOTFILES_DIR=~/dotfiles-test-clone
```

> link のターゲットが1台目と同じパスを指すので、symlink が上書きされる。
> 実運用でも「新しいマシンで install」はこの動作。

**確認：**

- [ ] `.dotfile-hook/` に hook が展開される
- [ ] `git config core.hooksPath` → `.dotfile-hook`
- [ ] symlink が配置される（1台目の link 結果を上書き）
- [ ] `[dotfile] watchサービスを登録しました` と表示される

---

## Phase 3 — 異常系

> Phase 2 の事前準備で `mode = "remote"` に変更済みであること。

### J. コンフリクト発生

```bash
# 1. clone 側で変更を push
cd ~/dotfiles-test-clone
git pull
echo "conflict from remote" >> ai-agent/AGENTS.md
git add -A && git commit -m "conflict remote" && git push

# 2. 元のリポジトリでもローカル commit を作る（fetch せずに）
cd ~/dotfiles-test
echo "conflict from local" >> ai-agent/AGENTS.md
git add -A && git commit -m "conflict local"

# 3. pull → コンフリクト処理が発動
cd /d/work/dotfile/work
make exe-pull
```

**確認：**

- [ ] `conflict/<hostname>/<timestamp>` ブランチが作成される
- [ ] ローカルの変更がそのブランチに退避される
- [ ] main がリモートの HEAD にリセットされる（リモート側の内容になる）
- [ ] `.conflict-pending` ファイルが作成される
- [ ] `dotfile status` → `CONFLICT PENDING` 警告バナー

**設計思想：** 設定ファイルの auto-merge は危険なので行わない。ローカル変更を別ブランチに保護した上で、main をリモートに合わせる。ユーザーが明示的に解消する。

### J-2. コンフリクト解消

```bash
cd ~/dotfiles-test

# 退避された内容を確認
git log --oneline --graph --all
# → conflict/<hostname>/<timestamp> ブランチにローカル変更がある

# 必要な変更を cherry-pick（不要ならスキップ）
git cherry-pick <conflict-branch の commit hash>
# コンフリクトが出たら手動で解消 → git add → git cherry-pick --continue

# 解消したら conflict ブランチを削除
git branch -D conflict/<hostname>/<timestamp>

# 次の pull でマーカーが自動で消える
cd /d/work/dotfile/work
make exe-pull
```

**確認：**

- [ ] conflict ブランチ削除後の pull → `.conflict-pending` が消える
- [ ] `コンフリクト解消を確認し、マーカーを削除しました` と表示
- [ ] `status` が `[status] No conflicts.` に戻る

### K. pre-push hook — コンフリクトマーカー検出

```bash
cd ~/dotfiles-test
echo "<<<<<<< HEAD" >> ai-agent/AGENTS.md
git add -A && git commit -m "bad marker"
git push origin main
# ※ pre-push hook のテストなので git push を直接使う
```

**確認：**

- [ ] push がブロックされる
- [ ] `未解決のコンフリクトマーカーが残っています` と表示

```bash
# 後始末
git reset --soft HEAD~1
git checkout -- ai-agent/AGENTS.md
```

### L. pre-push hook — シークレット検出

```bash
cd ~/dotfiles-test
echo 'api_key = "sk-abc123def456"' >> ai-agent/AGENTS.md
git add -A && git commit -m "secret leak"
git push origin main
# ※ pre-push hook のテストなので git push を直接使う
```

**確認：**

- [ ] push がブロックされる
- [ ] `シークレットらしき記述を検出` と表示

**検出後の対処：**

1. シークレットを除去 → 新しい commit → 再 push
2. 誤検知なら `SKIP_SECRET_SCAN=1 git push` で回避

```bash
# 後始末
git reset --soft HEAD~1
git checkout -- ai-agent/AGENTS.md
```

### M. delete-category

```bash
# テスト用カテゴリを作成
mkdir -p ~/dotfiles-test/throwaway
echo "test" > ~/dotfiles-test/throwaway/file.txt

# sync.toml の auto 配列に "throwaway" を追加
# auto = ["ai-agent", "editor", "shell", "throwaway"]
cd ~/dotfiles-test
git add sync.toml throwaway/ && git commit -m "chore: add throwaway category"
cd /d/work/dotfile/work

# コミット済み変更も含めて push
make exe-push

# 削除
make exe-delete-category throwaway
```

**確認：**

- [ ] `throwaway/` ディレクトリが消える
- [ ] `sync.toml` から `throwaway` が消える
- [ ] 1コミットで commit + push される
- [ ] GitHub 上でも反映

---

## 後片付け

```bash
# テスト用ディレクトリ削除
rm -rf ~/dotfiles-test
rm -rf ~/dotfiles-test-clone

# GitHub テストリポジトリ削除
gh repo delete kokukaityo/dotfiles-test --yes

# バックアップを復元
rm -rf ~/.claude
cp -r ~/.claude.bak ~/.claude
rm -rf ~/.claude.bak

rm -rf ~/.codex
cp -r ~/.codex.bak ~/.codex
rm -rf ~/.codex.bak

rm -rf ~/.agents
cp -r ~/.agents.bak ~/.agents
rm -rf ~/.agents.bak

# watch サービス（スタートアップ VBScript）を削除
rm -f "$APPDATA/Microsoft/Windows/Start Menu/Programs/Startup/dotfile-watch.vbs"
```
