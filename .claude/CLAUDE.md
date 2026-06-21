# dotfile — Bash CLI engine for cross-machine dotfile synchronization

複数マシン間で dotfile（設定ファイル）を Git ベースで同期する CLI エンジン。
エンジン（このリポジトリ / public OSS / MIT）とデータ（各ユーザーの private repo）は分離済み。

## Stack

- 言語: Bash（POSIX sh 互換を意識しつつ bashism 許容）
- ビルドシステム / パッケージマネージャ: なし
- テスト: 未導入（E2E 優先方針、bats 予定）
- CI: 未導入（GitHub Actions 予定）

## コーディング規約

- `set -euo pipefail` を全スクリプト先頭に入れる
- 関数名:
  - `bin/dotfile` 内: `cmd_<subcommand>`（例: `cmd_init`, `cmd_setup`）
  - `lib/` 内: `動詞_名詞`（例: `detect_os`, `generate_commit_msg`, `resolve_dotfile_dir`）
- エラーメッセージ: `[dotfile]` `[sync]` `[pre-push]` `[post-merge]` 等のプレフィックス付き、stderr（`>&2`）へ出力
- ユーザー向け出力メッセージは日本語
- macOS 互換: `readlink -f` は使わない。`bin/dotfile` の `resolve_path()` を参照
- 変数展開にはデフォルト値を設定する: `${VAR:-}` or `${VAR:-default}`

## コミット規約

- Conventional Commits（`feat:` / `fix:` / `refactor:` / `docs:` / `chore:`）
- 日本語メッセージ可

## 参照ドキュメント

- `doc/Architecture.md` — 全体構成・設計パターン・レイヤー間の責務と依存関係
- `doc/Development.md` — 開発環境・テスト方針・既知バグ・将来計画
