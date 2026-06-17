# dotfiles — AI設定の正本リポジトリ

複数端末（私用・業務）で、AIエージェント設定（AGENTS.md / Skills / commands 等）を
正本一元管理し、symlinkで各ツールへ配置し、Gitで端末間同期する。

## 設計原則

- **正本は1箇所**：`~/dotfiles` 配下の実体ファイル。各ツールの設定パスはsymlinkで参照する。
- **同期はGit**：commit/push/pull で端末間に伝播。履歴・差分が残るので「育てる」用途に向く。
- **コンフリクトは自動mergeしない**：hookは `pull --ff-only` で、分岐を検知したら**止めて警告**する。
  勝手にmergeして正本を汚さない。
- **シークレットは同期しない**：APIキー・トークン・`auth.json` 類はdotfilesに入れない（§セキュリティ）。
- **育成対象を分離**：AGENTS.mdは手動キュレーション（少数精鋭）。自動で溜めたいログ・手順はSkills/notes側へ。

## ディレクトリ構成

```
~/dotfiles/
├── ai/
│   ├── AGENTS.md                 # グローバル個人ルール（正本・手動編集）
│   ├── ai-setup-local.md         # リポジトリ用 生成器の元本
│   ├── skills/                   # 育てたSkill群（正本）
│   │   └── <skill-name>/
│   │       └── SKILL.md
│   └── commands/                 # カスタムスラッシュコマンド等
├── install/
│   ├── link.sh                   # symlink配置（OS分岐）
│   └── setup-hooks.sh            # Git hook 設置
├── hooks/
│   ├── post-merge               # pull後にsymlink再張り直し
│   └── pre-push                 # 未解決コンフリクト等を検知してpushを止める
├── .gitattributes               # symlinkを実体展開させない設定
└── README.md
```

## 配置されるsymlink（link.sh が張る）

| 正本（実体） | リンク先（各ツールが読む場所） |
|---|---|
| `~/dotfiles/ai/AGENTS.md` | `~/.codex/AGENTS.md` |
| `~/dotfiles/ai/AGENTS.md` | `~/.claude/CLAUDE.md` |
| `~/dotfiles/ai/skills/` | `~/.agents/skills/`（中身を個別リンク） |
| `~/dotfiles/ai/commands/` | `~/.claude/commands/` |

> import方式をやめてsymlinkに統一。Codexがimport記法を解釈しない可能性を回避でき、
> ツールから見れば常に「実体ファイルがそこにある」状態になる。

## 初期セットアップ手順（新端末）

```bash
git clone git@github.com:<you>/dotfiles.git ~/dotfiles
cd ~/dotfiles
bash install/link.sh        # symlinkを配置（既存ファイルはバックアップ）
bash install/setup-hooks.sh # Git hookを設置
```

これだけで、その端末の全AIツールが正本を参照する状態になる。

## 日常運用

- **編集**：どの端末でも `~/dotfiles/ai/` 配下を直接編集してよい（symlink先を編集しても実体に届く）。
- **同期**：
  - セッション開始時に自動 `pull --ff-only`（post-mergeフックがsymlinkを張り直す）。
  - **pushは手動**を基本にする：`cd ~/dotfiles && git add -A && git commit -m "..." && git push`。
  - pre-pushフックが、未解決コンフリクトマーカーやdetached状態を検知したらpushを中止する。

### コンフリクトが起きたら（`pull --ff-only` が失敗したら）
端末間で分岐している合図。慌てず：
```bash
cd ~/dotfiles
git fetch
git log --oneline --graph --all   # どちらが進んでいるか確認
git rebase origin/main            # 自分の変更を相手の上に乗せ直す（個人ファイルなので大抵すぐ済む）
# コンフリクト箇所を手で直して
git rebase --continue
git push
```
個人のdotfilesなので、同一行の同時編集が無い限りコンフリクトはまず起きない。
起きても解決は数分。**自動mergeさせず、必ず人が見て解決する**のが正本を守るコツ。

## 多端末編集でコンフリクトを最小化する習慣

- 編集を始める前に一度 `pull`（フックに任せず手でもよい）。
- 編集が終わったら**こまめにpush**（セッション終了まで溜めない）。push遅延が分岐の最大要因。
- 大きめの再構成は1端末に寄せる。

## セキュリティ（業務端末を含むため重要）

- **APIキー・トークン・`~/.codex/auth.json`・SSH鍵はdotfilesに置かない**。各端末ローカルに置き、環境変数で参照。
- MCPの**サーバー定義は同期してよいが、認証部分は環境変数化**して実値はコミットしない。
- リポジトリは必ず **private**。誤って鍵を入れた場合に備え `.gitignore` で `*auth*`, `*.key`, `*.pem`, `.env*` を除外。
- 業務端末：私的Gitリポジトリの設置が許される前提だが、**業務情報・社内固有設定は絶対にこのdotfilesへ入れない**
  （私的設定の同期に限定する）。業務固有のものは業務端末ローカルに別管理。

## 業務端末の運用ルール（編集可だが慎重に）

- 業務端末からも編集・pushしてよいが、**機微情報の混入に最大限注意**。push前に `git diff --cached` を必ず目視。
- 業務ネットワークのプロキシ等でpushが不安定なら、業務端末は**pull優先・pushは最小限**に寄せると事故が減る。
