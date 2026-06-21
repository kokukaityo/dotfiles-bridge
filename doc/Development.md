# 開発ガイド

## 開発環境

- **Windows**: Git Bash（MSYS2 / MinGW）
- **macOS / Linux**: bash
- **Codespace**: `.devcontainer/` あり。セットアップ手順は `doc/codespace_setup.md` を参照
- **ShellCheck**: 未導入（将来導入予定）

## ブランチ運用

| ブランチ | 用途 |
|---|---|
| `main` | 安定版。auto-sync 対象 |
| `feature/*` | 開発作業用。main から切って完了後 main に merge |

## テスト方針

現状はテスト未導入。以下の優先順位で段階的に導入する。

1. **Docker E2E**（最優先）: `ubuntu:24.04` でクリーン環境を作り、`dotfile init` → `setup` → `push` → `pull` を実際に実行。「他人のクリーン環境で動くか」を検証する
2. **bats**（次点）: 壊れたら痛い経路（symlink 作成、sync、conflict 退避）だけ 3〜5 本の振る舞いテスト
3. **GitHub Actions CI**（最後）: push ごとに Docker(Linux) + macOS runner で bats を回す

E2E を優先する理由: このツールの本質は symlink・Git・ファイル配置という副作用の検証であり、純粋関数のロジック検証ではない。

## 変更時のチェックリスト

- [ ] `set -euo pipefail` がスクリプト先頭にあるか
- [ ] macOS 互換: `readlink -f` を使っていないか
- [ ] `sync.sh` の `cmd_push()` 内の pathspec が `SYNC_AUTO` カテゴリに限定されているか
- [ ] `template/` への変更が必要か（データリポジトリ側に影響する変更の場合）
- [ ] エラーメッセージに適切なプレフィックス（`[dotfile]`, `[sync]` 等）があるか
- [ ] Windows（Git Bash）で動作するか（パス区切り、symlink 権限に注意）

## 既知バグ・改善点

現行コードと照合して有効なもののみ記載。

| # | 深刻度 | 内容 | 該当箇所 |
|---|---|---|---|
| 1 | 中 | `pre-push` はワークツリーのみ検査。一度コミットして次のコミットで削除したシークレットは、push 対象の履歴に残っていても検出されない | `lib/hook/pre-push` |
| 2 | 低 | `post-merge` で `link.sh` が失敗しても `"symlink refreshed"` と表示される | `lib/hook/post-merge` |

### 修正済みの項目（旧 doc/review.md より）

以下は分離リファクタリング（`01c8761`）で対応済み:

- `sync.sh` の設定読み込み不具合 → `conf.sh` で `sync.conf` を直接 source する方式に変更
- 自動 push が手動変更まで commit → `git commit --only` + pathspec 限定
- auto カテゴリのディレクトリ削除が同期されない → `cmd_delete_category()` で対応
- ローカル ahead が競合扱い → `merge_base` 比較で ahead を正しく判定

## .agent/ の使い方

エージェント間の作業受け渡し用ディレクトリ。中身はリモートに push されない（`.gitignore` 設定済み）。

| ディレクトリ | 用途 |
|---|---|
| `.agent/plan/` | 設計書・実装計画。次のエージェントが読んで作業を続行できる状態にする |
| `.agent/notes/` | 作業ログ・備忘録。調査結果や判断の記録 |

## 将来計画

### 短期（shell のまま）

- 既知バグの修正（上記テーブル参照）
- Docker E2E テスト環境の構築
- bats テストの導入（壊れたら痛い経路 3〜5 本）

### 中期

- GitHub Actions CI（Linux + macOS）
- installer スクリプト（`curl | bash` でインストール）
- `dotfile self-update` サブコマンド

### 長期

- Go によるエンジン書き直し（シングルバイナリ配布）
- Homebrew tap
