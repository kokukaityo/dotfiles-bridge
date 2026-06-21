# dotfile — AI エージェント向け開発ガイド

複数マシン間で設定ファイルを Git ベースで同期する Bash CLI エンジン。
このリポジトリは public なエンジンであり、ユーザー固有の設定データは private なデータリポジトリに分離する。

## 応答

- 応答は日本語で、明瞭・簡潔・実用的に記述する。
- 時刻は JST（UTC+9）の24時間制で表記する。
- 単位はメートル法、通貨は日本円（¥）を基本とし、海外製品・サービスは元通貨も併記する。
- 技術的な話題では事実と根拠を優先し、懸念や別案があれば率直に示す。
- 不明点を推測で埋めると結果が大きく変わる場合は、実装前に確認する。

## Stack

- 言語: Bash（POSIX sh 互換を意識しつつ bashism を許容）
- 対象環境: Windows Git Bash、macOS、Linux
- ビルドシステム / パッケージマネージャ: なし
- テスト: 未導入。Docker E2E、bats の順で導入予定
- CI: 未導入。GitHub Actions を導入予定

## 主要な構成

- `bin/dotfile`: CLI エントリポイントとサブコマンドのルーティング
- `lib/conf.sh`: データリポジトリ解決、設定読み込み、バージョン互換チェック
- `lib/sync.sh`: pull、push、カテゴリ削除、`.gitignore` 生成、状態表示
- `lib/link.sh`: OS 判定、`link.yaml` 解析、symlink 配置
- `lib/hook/`: pre-push / post-merge Git hook
- `template/`: `dotfile init` で生成するデータリポジトリの雛形
- `doc/Architecture.md`: 設計、責務、依存関係
- `doc/Development.md`: 開発環境、テスト方針、既知の問題

変更前に、対象機能に応じて上記ドキュメントと周辺コードを読むこと。

## 設計上の制約

- エンジンにユーザー固有データやシークレットを含めない。
- `bin/dotfile` はルーティングを担い、主要ロジックは `lib/` に置く。
- `cmd_init()` と `cmd_setup()` は初期化前にも動作できる構造を維持する。
- 同期カテゴリをコードへハードコードせず、`sync.conf` の `SYNC_AUTO`、`SYNC_MANUAL`、`SYNC_IGNORE` を使う。
- `link.yaml` は OS ごとの完全なエントリ一覧とし、暗黙の共通設定や上書き規則を追加しない。
- データリポジトリの形式に影響する変更では、`template/` と互換性への影響も確認する。

## Bash コーディング規約

- 実行スクリプトの先頭で `set -euo pipefail` を有効にする。
- `bin/dotfile` 内のコマンド関数は `cmd_<subcommand>` とする。
- `lib/` 内の関数は `detect_os`、`generate_commit_msg` のような動詞ベースの snake_case とする。
- 変数を引用し、未設定を許容する展開には `${VAR:-}` または `${VAR:-default}` を使う。
- ユーザー向けメッセージは日本語にする。
- エラーは `[dotfile]`、`[sync]` など適切なプレフィックスを付け、stderr へ出力する。
- macOS で利用できない `readlink -f` は使わない。必要なら `bin/dotfile` の `resolve_path()` を参照する。
- GNU 固有機能への依存を増やす場合は macOS と Git Bash での代替手段を確認する。

## 変更時の注意

- `lib/sync.sh` の push 対象を `SYNC_AUTO` カテゴリ外へ広げない。
- symlink、Git履歴、ユーザーファイルを扱う処理では、既存データを失わない設計を優先する。
- `rm`、`git reset --hard`、force push などの破壊的操作を新たに実行する場合は、事前にユーザーへ確認する。
- シークレット検知を弱める変更や、個人情報をテンプレートへ追加する変更は避ける。
- 依頼と無関係なファイルや既存の未コミット変更を編集しない。

## 検証

現時点では自動テストがないため、変更内容に応じて最低限次を実施する。

1. 変更した Bash ファイルを `bash -n <file>` で構文確認する。
2. CLI に影響する場合は `bash bin/dotfile help` など、副作用のない経路で基本動作を確認する。
3. Git、symlink、ファイル削除を伴う確認は、実データではなく一時ディレクトリまたは隔離したテストリポジトリで行う。
4. Windows Git Bash、macOS、Linux の互換性リスクを確認する。

検証できなかった項目は、完了報告で理由とともに明記する。

## ドキュメント

- コマンド、設定形式、利用手順を変更した場合は `README.md` を更新する。
- 設計やレイヤー責務を変更した場合は `doc/Architecture.md` を更新する。
- 開発手順、テスト方針、既知の問題を変更した場合は `doc/Development.md` を更新する。
- 長い作業の設計や引き継ぎが必要な場合は `.agents/plan/`、調査メモは `.agents/notes/` を使う。
- Codex 標準の拡張領域と共存するため、Skills は `.agents/skills/`、プラグイン関連は `.agents/plugins/` に分離する。

## コミット

- Conventional Commits を使用する（`feat:`、`fix:`、`refactor:`、`docs:`、`test:`、`chore:` など）。
- コミットメッセージは日本語でもよい。
- ユーザーから明示的に依頼されない限り、コミットや push は行わない。
