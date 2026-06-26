# Security Policy

## セキュリティ機能

dotfiles に API キーやトークンが混入するリスクに対し、3 層の防御を備えています。

1. **`.gitignore` 自動生成** — `dotfiles gitignore` で `*auth*`, `*.key`, `*.pem`, `.env*` 等を自動的に除外します
2. **pre-push hook のシークレットスキャン** — push 前に差分をスキャンし、`api_key`, `secret`, `password`, `token`, 秘密鍵ヘッダーなどのパターンを検出した場合は push をブロックします
3. **ignore カテゴリ** — `sync.toml` の `ignore` に登録したカテゴリは Git 追跡から完全に除外されます

加えて、`sync.toml` の `auto` に登録していないカテゴリは `dotfiles push` や `watch` による自動 push の対象外です（`manual` 扱い）。新しいカテゴリを追加しても、明示的に `auto` へ登録しない限りリモートへ自動送信されないため、意図しない流出に対するフェイルセーフになっています。

## 推奨事項

- `mode = "local"`（デフォルト）ではデータは一切外部に送信されません
- リモート同期を使う場合は、リポジトリを **private** に設定してください
- シークレットスキャンの誤検知は `SKIP_SECRET_SCAN=1 git push` で回避できますが、内容を確認してから使用してください

## サポート対象バージョン

Alpha 版のため、最新リリースのみセキュリティ修正の対象です。

| バージョン | サポート |
| --- | --- |
| latest release | :white_check_mark: |
| それ以前 | :x: |

## 脆弱性の報告

セキュリティ上の問題を見つけた場合は、**public な Issue ではなく** [GitHub Security Advisories](https://github.com/kokukaityo/dotfiles-bridge/security/advisories/new) から非公開で報告してください。

報告には以下を含めてください:

- 再現手順
- 影響範囲の見立て
- 可能であれば修正案

確認後、パッチリリースで対応します。修正がリリースされた時点で報告者にクレジットを記載します（希望しない場合はその旨お伝えください）。

## 対象スコープ

以下が対象です:

- dotfiles-bridge 本体（Go ソースコード）
- テンプレート（`template/` 配下）
- インストーラスクリプト（`install.sh`）
- Git hooks（`template/hooks/` 配下）

ユーザーが自身のデータリポジトリ（`~/dotfiles`）に置くファイルの内容は対象外です。
