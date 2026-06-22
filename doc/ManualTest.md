# dotfile エンジン 手動動作確認手順書

## Context

v0.1 の機能実装が完了したので、自分の PC（Windows）で全コマンドが正しく動作するか一通り確認する。コンフリクト処理やバックアップ挙動を含めた E2E レベルの検証。

## 構成

テストは段階的に進める：

1. **Phase 1 — ローカル完結**（リモート不要）: init, link, status, version, gitignore
2. **Phase 2 — リモート追加**: push, pull, 2台目（clone → setup）
3. **Phase 3 — 異常系**: コンフリクト発生〜解消、pre-push hook

---

## 事前準備

### バックアップ

`ai-agent/link.toml` が以下のパスに symlink を張る。事前にバックアップ：

```bash
cp -r ~/.claude ~/.claude.bak
cp -r ~/.codex ~/.codex.bak
cp -r ~/.agents ~/.agents.bak
```

### ビルド & PATH 設定

```bash
cd /d/work/dotfile/work
make build
# → dist/dotfile.exe が生成される

# テスト中は絶対パスか、PATH に追加して使う
export PATH="/d/work/dotfile/work/dist:$PATH"
# これで "dotfile" コマンドとして使える
```

### Windows 開発者モード

設定 → システム → 開発者向け → 開発者モード ON（symlink 作成に必須）

---

## Phase 1 — ローカル完結（リモート不要）

### A. init

```bash
dotfile init ~/dotfiles-test
```

**確認：**

- [ ] `~/dotfiles-test/` が作成される
- [ ] `.infra-version` にエンジンバージョンが書かれている
- [ ] `sync.toml` が存在（default_branch, auto, ignore）
- [ ] `ai-agent/link.toml`, `editor/link.toml`, `shell/link.toml` が存在
- [ ] `.dotfile-hook/pre-push`, `.dotfile-hook/post-merge` が存在
- [ ] `.gitignore` にマーカー行と自動生成セクションがある
- [ ] `git log` → `feat: initial dotfiles setup` の初回コミット
- [ ] `.backup/` ディレクトリが存在する

### B. link

```bash
cd ~/dotfiles-test
dotfile link
```

> **環境変数は不要。** `cd` でデータリポジトリに入れば、git root + `.infra-version` の存在で自動検出される。

**確認：**

- [ ] `~/.claude/CLAUDE.md` → `~/dotfiles-test/ai-agent/AGENTS.md` の symlink
- [ ] `~/.codex/AGENTS.md` → 同上
- [ ] `~/.claude/commands/` → `~/dotfiles-test/ai-agent/commands/` の symlink
- [ ] `~/.agents/skills/` → `~/dotfiles-test/ai-agent/skills/` の symlink
- [ ] 元のファイルが `.backup/ai-agent_<timestamp>/` に退避されている
- [ ] 2回目の `dotfile link` → `ok (already linked)` と表示（冪等）

### C. status

```bash
cd ~/dotfiles-test
dotfile status
```

**確認：**

- [ ] コンフリクトがない状態 → `[status] No conflicts.` と表示

### D. version

```bash
dotfile version
```

**確認：**

- [ ] `dotfile engine v<version>` と表示される

### E. gitignore

```bash
cd ~/dotfiles-test
# sync.toml を編集して ignore に追加
# ignore = ["backup", "raw", "secret-stuff"]
dotfile gitignore
```

**確認：**

- [ ] `.gitignore` の自動生成部分に `secret-stuff/` が追加される
- [ ] マーカー上の手書き部分は保持されたまま

---

## Phase 2 — リモート追加

### F. リモート接続 & push

```bash
# テスト用プライベートリポジトリ作成
gh repo create dotfiles-test --private

cd ~/dotfiles-test
git remote add origin git@github.com:kokukaityo/dotfiles-test.git
git push -u origin main

# ファイルを変更して push
echo "test change" >> ai-agent/AGENTS.md
dotfile push
```

**確認：**

- [ ] auto カテゴリ（ai-agent）の変更が commit される
- [ ] コミットメッセージに `update: AGENTS.md` のような分類
- [ ] GitHub 上に反映される

### F-2. manual カテゴリは push されない

```bash
mkdir -p ~/dotfiles-test/experiment
echo "test" > ~/dotfiles-test/experiment/note.txt
dotfile push
```

**確認：**

- [ ] `experiment/` は commit されない（auto に含まれてないため）

### G. pull（別の場所から変更を push → 元で pull）

```bash
# 別ディレクトリに clone
cd ~
git clone git@github.com:kokukaityo/dotfiles-test.git dotfiles-test-clone

# clone 側で変更して push
cd ~/dotfiles-test-clone
echo "remote change" >> ai-agent/AGENTS.md
git add -A && git commit -m "remote update" && git push

# 元のリポジトリで pull
cd ~/dotfiles-test
dotfile pull
```

**確認：**

- [ ] `[sync] Fast-forwarded to origin/main.` と表示
- [ ] ローカルに変更が反映されている

### H. 2台目セットアップ（clone → setup）

```bash
cd ~/dotfiles-test-clone
dotfile setup
```

> link のターゲットが1台目と同じパスを指すので、symlink が上書きされる。
> 実運用でも「新しいマシンで setup」はこの動作。

**確認：**

- [ ] `.dotfile-hook/` に hook が展開される
- [ ] `git config core.hooksPath` → `.dotfile-hook`
- [ ] symlink が配置される（1台目の link 結果を上書き）

---

## Phase 3 — 異常系

### I. コンフリクト発生

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
dotfile pull
```

**確認：**

- [ ] `conflict/<hostname>/<timestamp>` ブランチが作成される
- [ ] ローカルの変更がそのブランチに退避される
- [ ] main がリモートの HEAD にリセットされる（リモート側の内容になる）
- [ ] `.conflict-pending` ファイルが作成される
- [ ] `dotfile status` → `CONFLICT PENDING` 警告バナー

**設計思想：** 設定ファイルの auto-merge は危険なので行わない。ローカル変更を別ブランチに保護した上で、main をリモートに合わせる。ユーザーが明示的に解消する。

### I-2. コンフリクト解消

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
dotfile pull
```

**確認：**

- [ ] conflict ブランチ削除後の pull → `.conflict-pending` が消える
- [ ] `コンフリクト解消を確認し、マーカーを削除しました` と表示
- [ ] `status` が `[status] No conflicts.` に戻る

### J. pre-push hook — コンフリクトマーカー検出

```bash
cd ~/dotfiles-test
echo "<<<<<<< HEAD" >> ai-agent/AGENTS.md
git add -A && git commit -m "bad marker"
git push origin main
```

**確認：**

- [ ] push がブロックされる
- [ ] `未解決のコンフリクトマーカーが残っています` と表示

```bash
# 後始末
git reset --soft HEAD~1
git checkout -- ai-agent/AGENTS.md
```

### K. pre-push hook — シークレット検出

```bash
cd ~/dotfiles-test
echo 'api_key = "sk-abc123def456"' >> ai-agent/AGENTS.md
git add -A && git commit -m "secret leak"
git push origin main
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

### L. delete-category

```bash
cd ~/dotfiles-test

# テスト用カテゴリを作成
mkdir throwaway
echo "test" > throwaway/file.txt

# sync.toml の auto に追加: auto = ["ai-agent", "editor", "shell", "throwaway"]
dotfile push

# 削除
dotfile delete-category throwaway
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
mv ~/.claude.bak ~/.claude

rm -rf ~/.codex
mv ~/.codex.bak ~/.codex

rm -rf ~/.agents
mv ~/.agents.bak ~/.agents
```
