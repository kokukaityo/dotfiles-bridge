# dotfiles — 設定の正本リポジトリ

複数端末（私用・業務、Mac・Windows）で、各種ツール設定を正本一元管理する。
symlinkで各ツールへ配置し、Gitで端末間同期する。

## 設計原則

- **正本は1箇所**：`~/dotfiles` 配下の実体ファイル。各ツールの設定パスはsymlinkで参照する。
- **カテゴリ分類**：設定はカテゴリ別ディレクトリに分ける（ai-agent/, editor/, shell/ 等）。命名は単数形で統一。
- **2層管理**：設定ファイルはファイル監視で自動sync。インフラ（スクリプト・hook）は手動管理。
- **同期はGit**：commit/push/pull で端末間に伝播。履歴・差分が残るので「育てる」用途に向く。
- **conflict は退避**：分岐を検知したら conflict branch に退避し、main をリモートに合わせる。手動で resolve する。
- **commit message は自動生成**：diff から変更内容を要約。何をしたかが履歴に残る。
- **シークレットは同期しない**：APIキー・トークン・`auth.json` 類はdotfilesに入れない（§セキュリティ）。

## ディレクトリ構成

```
~/dotfiles/
├── .infra/                      # インフラ（隠しディレクトリ）
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

- **設定カテゴリ**（ai-agent/, editor/, shell/）：各カテゴリに `link.yaml` を置き、symlinkの配置先を宣言する。
- **インフラ**（.infra/）：sync・link・hookを集約。設定カテゴリとは視覚的に分離。

## symlink管理：link.yaml

各カテゴリの `link.yaml` で、OS別にsymlinkの配置先を宣言する。
OS別完全リスト方式：各OSセクションにそのOSの全エントリを書く（共通セクションは使わない）。

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

`.infra/link.sh` がOSを判定し、該当セクションのエントリを処理する。
新しいカテゴリを追加するときは、ディレクトリを作って `link.yaml` を置くだけ。

## ブランチ戦略

| ブランチ | 用途 | 管理 |
| --- | --- | --- |
| **main** | 正本。auto-sync対象。常にリモートと一致させる | 自動 |
| **develop** | インフラ改修・ファイル整理の作業用 | 手動。完了後mainにmerge |
| **conflict/{hostname}/{YYYYMMDD-HHMMSS}** | 分岐検知時の自動退避先 | 自動作成。手動resolve後に削除 |

- 普段使うのは **main のみ**。設定ファイルの自動syncはすべてmainに対して行われる。
- **develop** は大きめの構造変更や `.infra/` の改修時に使う。auto-syncは走らないので安全に作業できる。
- **conflict branch** は分岐検知時に自動で作られる。ローカルの変更を退避し、mainはリモートに合わせる。

## 同期フロー

### push（自動 — ファイル監視）

auto-sync対象カテゴリ（現在は ai-agent/ のみ）のファイル保存を監視し、自動でcommit/pushする。

```
ファイル保存を検知
→ debounce（数秒待って連続保存をまとめる）
→ git add <対象カテゴリ>/
→ commit message を diff から自動生成
→ git commit
→ git push origin main
```

- 他のカテゴリ（editor/, shell/）は将来opt-in可能な設計。
- インフラ（.infra/, README）は自動syncの対象外。手動でcommit/pushする。

### pull（セッション開始時）

シェル起動時（.bashrc/.zshrc）をベースに、各ツールのhookからも同じスクリプトを呼べる（冪等）。

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

分岐を検知したとき、以下の2つで通知する：

1. **ターミナル警告バナー**：シェル起動時に色付きの目立つ警告を表示。
2. **マーカーファイル**：`.conflict-pending` を作成。他ツール/スクリプトから検知可能。
   conflict branch が解消されたら、次回pull時に自動削除される。

### conflict の解消手順

```bash
cd ~/dotfiles
git log --oneline --graph --all       # 状況を確認
git checkout conflict/xxx/xxx         # conflict branch の内容を確認
git checkout main
git merge conflict/xxx/xxx            # main に取り込む（コンフリクトがあれば手で直す）
git branch -d conflict/xxx/xxx        # branch 削除
git push origin main
# 次回のシェル起動時に .conflict-pending が自動削除される
```

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

## 日常運用

- **設定ファイルの編集**：symlink先を直接編集してよい。保存すれば自動で同期される。
- **pull**：シェル起動時に自動。意識不要。
- **conflict**：ターミナルに警告が出たら、上記の解消手順に従う。
- **インフラの編集**：`.infra/` 配下を編集した場合は手動でcommit/push。

## セキュリティ（業務端末を含むため重要）

- **APIキー・トークン・`~/.codex/auth.json`・SSH鍵はdotfilesに置かない**。各端末ローカルに置き、環境変数で参照。
- MCPの**サーバー定義は同期してよいが、認証部分は環境変数化**して実値はコミットしない。
- リポジトリは必ず **private**。誤って鍵を入れた場合に備え `.gitignore` で `*auth*`, `*.key`, `*.pem`, `.env*` を除外。
- 業務端末：私的Gitリポジトリの設置が許される前提だが、**業務情報・社内固有設定は絶対にこのdotfilesへ入れない**。
- push前に pre-push hook がシークレット混入を検知して止める。
