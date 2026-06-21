# アーキテクチャ

## 全体構成

### エンジン/データ分離モデル

```
┌──────────────────────────────┐     ┌──────────────────────────────┐
│  dotfiles エンジン            │     │  ユーザーデータリポジトリ      │
│  (kokukaityo/dotfile)        │     │  (各ユーザーの private repo)  │
│  public / MIT                │     │                              │
│                              │     │  sync.conf                   │
│  bin/dotfiles ─── エントリ   │────▶│  .infra-version              │
│  lib/         ─── ロジック   │     │  ai-agent/  editor/  shell/  │
│  template/    ─── 雛形       │     │  (各カテゴリ + link.yaml)     │
└──────────────────────────────┘     └──────────────────────────────┘
     インストール先:                      デフォルト: ~/dotfiles
     ~/.local/share/dotfiles/             dotfiles init で生成
```

- エンジンに個人データは含まない。ユーザーは `dotfiles init` でデータリポジトリを生成する
- fork ではなく template 方式を採用。GitHub の fork ネットワークは CFOR（Cross Fork Object Reference）脆弱性があり、private repo の commit が fork 元の public repo 経由で参照可能になるため

### データリポジトリの解決順序

`lib/conf.sh` の `resolve_dotfiles_dir()` が以下の優先順位でデータリポジトリを探す:

1. `DOTFILES_DIR` 環境変数が設定されている → そのパス
2. カレント git root に `.infra-version` が存在する → そのリポジトリ
3. `$HOME/dotfiles` に `.infra-version` が存在する → `$HOME/dotfiles`
4. いずれも見つからない → エラー

---

## レイヤー構造と責務

```
bin/dotfiles         サブコマンドを受け取り、lib/ に委譲する
    │
    ├── lib/conf.sh      データリポジトリ解決、sync.conf 読み込み、バージョン互換チェック
    │       ↑ source される（lib/ 内の全スクリプトが依存）
    │
    ├── lib/sync.sh      Git 操作: pull / push / delete-category / gitignore / status
    │
    ├── lib/link.sh      symlink 配置: link.yaml 解析、OS 判定、symlink 作成
    │
    └── lib/hooks/
        ├── pre-push     push 前ガード: コンフリクトマーカー検知、シークレット検知
        └── post-merge   merge 後: symlink 再配置（dotfiles link を自動実行）
```

### bin/dotfiles — ルーティング層

- サブコマンド（`init`, `setup`, `link`, `pull`, `push`, `delete-category`, `gitignore`, `status`, `version`, `help`）を受け取り、対応する処理に振り分ける
- `cmd_init()` と `cmd_setup()` はこのファイル内に実装（`conf.sh` 不要で動く必要があるため）
- `link`, `pull`, `push` 等は `bash "$DOTFILES_ENGINE_LIB/sync.sh"` や `bash "$DOTFILES_ENGINE_LIB/link.sh"` で委譲
- `resolve_path()`: macOS の `readlink -f` 非対応を回避する自前のシンボリックリンク解決

### lib/conf.sh — 設定解決層

- `source` して使う（単体実行しない）
- データリポジトリのパスを `$DOTFILES` にセット
- `sync.conf` を source して `SYNC_AUTO`, `SYNC_MANUAL`, `SYNC_IGNORE` 配列を取得
- `check_version_compat()`: エンジンの `VERSION` とデータの `.infra-version` のメジャーバージョンを比較。不一致時は WARNING

### lib/sync.sh — 同期層

- `conf.sh` を source して自動的にデータリポジトリのコンテキストを取得
- **pull**: fetch → ff-only 可能なら merge / ローカルが ahead ならスキップ / 分岐検知なら conflict branch に退避
- **push**: main ブランチ限定。`SYNC_AUTO` カテゴリのファイルだけ `git add` → 自動 commit message 生成 → push
- **delete-category**: `SYNC_AUTO` カテゴリの完全削除（ディレクトリ削除 + sync.conf 書き換えを1 commit で atomic に実行）
- **gitignore**: `sync.conf` の `SYNC_IGNORE` からマーカー付きで `.gitignore` を自動生成。マーカーより上の手動記述は保持
- **status**: `.conflict-pending` マーカーの有無で警告バナーを表示

### lib/link.sh — symlink 配置層

- `conf.sh` を source してデータリポジトリのパスを取得
- `detect_os()`: `$OSTYPE` から OS キー（`darwin` / `win32` / `linux` / `unknown`）を判定
- `parse_link_yaml()`: 外部依存なしの手製 YAML パーサー。インデントベースで OS セクション → ソースファイル → ターゲットパスを解析
- `link()`: 既存ファイルがあれば `.bak.<timestamp>` にバックアップしてから symlink 作成
- `process_category()`: カテゴリディレクトリごとに link.yaml を処理
- 全カテゴリの `*/link.yaml` をイテレート

### lib/hooks/ — Git Hooks

- **pre-push**: ワークツリー内のコンフリクトマーカーとシークレットパターンを `git grep` で検査。`SKIP_SECRET_SCAN=1` でシークレット検知をバイパス可能
- **post-merge**: `dotfiles link`（PATH にあれば）または `lib/link.sh` を直接実行して symlink を再配置

### template/ — データリポジトリの雛形

- `dotfiles init` 実行時に `cp -r` でコピーされる
- `sync.conf`: デフォルトのカテゴリ定義
- `.infra-version`: エンジンバージョンとの互換チェック用
- 各カテゴリのサンプル（`ai-agent/`, `editor/`, `shell/` + `link.yaml`）

---

## 設計パターン・設計判断

### sync.conf による宣言的カテゴリ管理

カテゴリごとの同期モードを `sync.conf` の Bash 配列で宣言する:

| モード | 動作 | 例 |
|---|---|---|
| `SYNC_AUTO` | `dotfiles push` で自動 commit & push | `ai-agent`, `editor`, `shell` |
| `SYNC_MANUAL` | Git 追跡するが commit/push は手動 | — |
| `SYNC_IGNORE` | `.gitignore` に自動追加 | `backup`, `raw` |

スクリプト側にカテゴリ名のハードコーディングはない。カテゴリの追加・変更は `sync.conf` だけで完結する。

### link.yaml の OS 別完全リスト方式

```yaml
darwin:
    AGENTS.md:
        - ~/.codex/AGENTS.md
win32:
    AGENTS.md:
        - ~/.codex/AGENTS.md
```

各 OS セクションにその OS の全エントリを書く。「共通セクション」は持たない。
理由: OS 間で配置先が異なるケースが多く、共通セクション + 上書きの組み合わせは可読性が下がる。

### コンフリクト時の退避戦略

リモートと分岐した場合、merge ではなく `conflict/{hostname}/{timestamp}` ブランチに退避し、main を `origin/main` に reset する。

理由: dotfiles の同期で merge conflict が発生した場合、自動解決は危険。ユーザーが明示的に解消用ブランチで resolve してから main に ff-only で取り込む。main が壊れるリスクを排除する。

### バージョン互換チェック

エンジンの `VERSION` ファイルとデータの `.infra-version` のメジャーバージョンを比較する。不一致時は WARNING を出すが処理は続行する。

理由: エンジン更新でデータリポジトリのスキーマが変わる可能性がある。メジャーバージョン一致を要求することで不整合を早期に検知する。

### シークレット除外方針

- `.gitignore` で `*auth*`, `*.key`, `*.pem`, `.env*` を広めに除外
- `pre-push` hook でパターンマッチング（`api_key`, `secret`, `password`, `token`, `BEGIN RSA PRIVATE KEY` 等）
- プレースホルダー（`example`, `placeholder`, `<your`, `${`）は除外
