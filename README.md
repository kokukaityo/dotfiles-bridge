# dotfiles — 設定の正本リポジトリ

複数端末（私用・業務、Mac・Windows）で、各種ツール設定を正本一元管理する。
`~/dotfiles` を正本にし、symlink で各ツールへ配置し、Git で端末間同期する。

## このリポジトリでやること

- **正本管理**：設定ファイルの実体を `~/dotfiles` 配下に集約する。
- **symlink 配置**：各ツールが読む設定パスへ、正本から symlink を張る。
- **Git 同期**：commit / push / pull で複数端末へ変更を伝播する。
- **同期モードを宣言管理**：フォルダごとの同期モード（auto / manual / ignore）を `.infra/conf.sh` で定義する。
- **シークレットを除外**：APIキー・トークン・`auth.json` 類は同期しない。

## 日常運用

- **設定ファイルの編集**：symlink 先を直接編集してよい。保存すれば自動で同期される。
- **pull**：シェル起動時に自動で走る。普段は意識しなくてよい。
- **conflict**：ターミナルに警告が出たら、後述の「conflict の解消手順」に従う。
- **symlink 確認**：シェル起動時や sync 時に、各設定パスが正本への symlink のままか自動確認する。
- **インフラの編集**：`.infra/` 配下や README を編集した場合は、手動で commit / push する。

## 同期フロー

### push（自動 — ファイル監視）

`.infra/conf.sh` で `SYNC_AUTO` に指定されたカテゴリのファイル保存を監視し、自動で commit / push する。

```
ファイル保存を検知
→ conf.sh の SYNC_AUTO カテゴリか確認
→ ブランチ確認（main以外なら何もしない）
→ debounce（数秒待って連続保存をまとめる）
→ git add <対象カテゴリ>/
→ commit message を diff から自動生成
→ git commit
→ git push origin main
```

- **main 以外では auto-sync しない**。feature branch に切り替えるだけで sync が止まる。
- 対象カテゴリは `.infra/conf.sh` で宣言する（§同期モード設定）。
- `SYNC_MANUAL` のカテゴリ（`.infra/` 等）は手動で commit / push する。
- `SYNC_IGNORE` のカテゴリ（`backup/` 等）は Git 追跡しない。

### pull（セッション開始時）

シェル起動時（`.bashrc` / `.zshrc`）をベースに、各ツールの hook からも同じスクリプトを呼べる（冪等）。

```
git fetch origin
├─ ff-only 可能 → merge して完了
├─ 分岐検知 →
│   1. conflict/{hostname}/{YYYYMMDD-HHMMSS} ブランチを作成
│   2. ローカルの変更をそこに commit
│   3. main を origin/main に reset
│   4. 通知（下記）
└─ 最新 → 何もしない
```

### conflict 通知

分岐を検知したとき、以下の2つで通知する。

1. **ターミナル警告バナー**：シェル起動時に色付きの目立つ警告を表示する。
2. **マーカーファイル**：`.conflict-pending` を作成する。他ツールやスクリプトから検知できる。

`conflict branch` を削除した状態を解消済みとみなし、次回 pull 時に `.conflict-pending` が自動削除される。

### conflict の解消手順

main 上で直接 merge しない。解消用ブランチを切り、そこで resolve してから main に戻す。
main が壊れるリスクを避け、失敗しても捨ててやり直せるようにする。

```bash
cd ~/dotfiles
git log --oneline --graph --all              # 状況を確認

# 1. main から解消用ブランチを切る
git checkout -b resolve/conflict-xxx main

# 2. conflict branch を merge（コンフリクトがあれば手で直す）
git merge conflict/xxx/xxx
# ... 手動で resolve ...
git add . && git commit

# 3. 解消できたら main に戻して取り込む（ff-only で安全に）
git checkout main
git merge --ff-only resolve/conflict-xxx

# 4. 後片付け
git branch -d resolve/conflict-xxx
git branch -d conflict/xxx/xxx
git push origin main
# 次回のシェル起動時に .conflict-pending が自動削除される
```

## 設計原則

- **正本は1箇所**：`~/dotfiles` 配下の実体ファイル。各ツールの設定パスは symlink で参照する。
- **カテゴリ分類**：設定はカテゴリ別ディレクトリに分ける（`ai-agent/`, `editor/`, `shell/` 等）。命名は単数形で統一する。
- **同期モードは宣言的に管理**：フォルダごとの同期モード（auto / manual / ignore）を `.infra/conf.sh` で宣言する。スクリプトはこの定義を読んで動く。
- **同期はGit**：commit / push / pull で端末間に伝播する。履歴・差分が残るので「育てる」用途に向く。
- **conflict は退避**：分岐を検知したら `conflict branch` に退避し、main をリモートに合わせる。解消は手動で行う。
- **commit message は自動生成**：diff から変更内容を要約し、何をしたかが履歴に残る。
- **シークレットは同期しない**：APIキー・トークン・`auth.json` 類は dotfiles に入れない（§セキュリティ）。

## ディレクトリ構成

```
~/dotfiles/
├── .infra/                      # インフラ（隠しディレクトリ）
│   ├── conf.sh                  # 共通パス・フォルダ同期モード定義
│   ├── sync.sh                  # pull / push サブコマンド
│   ├── link.sh                  # symlink配置エンジン（link.yamlを読む）
│   ├── setup.sh                 # 初期セットアップ
│   └── hook/
│       ├── post-merge           # pull後にsymlink再配置
│       └── pre-push             # シークレット検知
├── ai-agent/                    # AI設定（auto-sync対象）
│   ├── AGENTS.md                # グローバル個人ルール
│   ├── ai-setup-local.md        # リポジトリ用 生成器の元本
│   ├── skills/                  # 育てたSkill群
│   ├── commands/                # カスタムスラッシュコマンド等
│   └── link.yaml                # symlink定義
├── editor/                      # エディタ設定（将来拡張）
│   └── link.yaml
├── shell/                       # シェル設定（将来拡張）
│   └── link.yaml
└── README.md
```

- **設定カテゴリ**（`ai-agent/`, `editor/`, `shell/`）：各カテゴリに `link.yaml` を置き、symlink の配置先を宣言する。
- **インフラ**（`.infra/`）：sync・link・hook を集約する。設定カテゴリとは視覚的に分離する。

## symlink 管理：link.yaml

各カテゴリの `link.yaml` で、OS 別に symlink の配置先を宣言する。
OS 別完全リスト方式として、各 OS セクションにその OS の全エントリを書く（共通セクションは使わない）。

```yaml
# ai-agent/link.yaml
darwin:
    AGENTS.md:
        - ~/.codex/AGENTS.md
        - ~/.claude/CLAUDE.md
    commands/:
        - ~/.claude/commands/
    skills/:
        - ~/.agents/skills/

win32:
    AGENTS.md:
        - ~/.codex/AGENTS.md
        - ~/.claude/CLAUDE.md
    commands/:
        - ~/.claude/commands/
    skills/:
        - ~/.agents/skills/
```

`.infra/link.sh` が OS を判定し、該当セクションのエントリを処理する。
新しいカテゴリを追加するときは、ディレクトリを作って `link.yaml` を置くだけ。

## 同期モード設定：conf.sh

`.infra/conf.sh` で、共通パスとフォルダごとの同期モードを宣言する。

| モード     | 動作                                    | 例                               |
| ---------- | --------------------------------------- | -------------------------------- |
| **auto**   | ファイル監視→自動 commit / push         | `ai-agent/`, `editor/`, `shell/` |
| **manual** | Git 追跡するが commit / push は手動     | `.infra/`, README                |
| **ignore** | Git 追跡しない（`.gitignore` 自動生成） | `backup/`, `raw/`                |

```bash
# .infra/conf.sh
INFRA="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILE="$INFRA/.."

SYNC_AUTO=(ai-agent editor shell)
SYNC_MANUAL=(.infra)
SYNC_IGNORE=(backup raw)
```

- いずれの配列にも未記載のフォルダは、同期処理の対象外となる（実質 `manual`）。
- `SYNC_IGNORE` のフォルダは `.gitignore` へ自動追記される。
- カテゴリの追加・変更は `conf.sh` の配列を編集するだけ。スクリプト側の変更は不要。

## ブランチ戦略

| ブランチ                                  | 用途                                             | 管理                                      |
| ----------------------------------------- | ------------------------------------------------ | ----------------------------------------- |
| **main**                                  | 正本。`auto-sync` 対象。常にリモートと一致させる | 自動                                      |
| **feature/xxx**                           | インフラ改修・構造変更の作業用                   | 手動。main から切って完了後 main に merge |
| **conflict/{hostname}/{YYYYMMDD-HHMMSS}** | 分岐検知時の自動退避先                           | 自動作成。手動 resolve 後に削除           |

- 普段使うのは **main のみ**。設定ファイルの `auto-sync` はすべて main に対して行われる。
- **feature branch** はインフラ改修や構造変更時に main から切る。main 以外では `auto-sync` が走らないので安全に作業できる。
- **conflict branch** は分岐検知時に自動で作られる。ローカルの変更を退避し、main はリモートに合わせる。

## commit message 自動生成

auto-commit 時に diff から変更内容をシェルスクリプトで要約する。

- フォーマット：`{action}: {files}`
- 例：
    - `update: AGENTS.md`
    - `add: skills/code-review`
    - `update: AGENTS.md, commands/review.md`
- 外部依存なし（シェルスクリプトのみで完結）

## 初期セットアップ（新端末）

```bash
git clone git@github.com:kokukaityo/dotfile.git ~/dotfiles
cd ~/dotfiles
bash .infra/setup.sh        # symlink配置 + Git hook設置
```

## セキュリティ（業務端末を含むため重要）

- **APIキー・トークン・`~/.codex/auth.json`・SSH鍵は dotfiles に置かない**。各端末ローカルに置き、環境変数で参照する。
- MCP の**サーバー定義は同期してよいが、認証部分は環境変数化**して実値はコミットしない。
- リポジトリは必ず **private**。誤って鍵を入れた場合に備え、`.gitignore` で `*auth*`, `*.key`, `*.pem`, `.env*` を広めに除外する。必要なファイルが過剰に無視されていないかは、shell から確認しやすい check コマンドを用意する。
- 業務端末：私的 Git リポジトリの設置が許される前提だが、**業務情報・社内固有設定は絶対にこの dotfiles へ入れない**。
- push 前に pre-push hook がシークレット混入を検知して止める。
