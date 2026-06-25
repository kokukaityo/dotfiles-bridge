# 開発ガイド

## 必要な環境

| ツール        | バージョン | 用途               |
| ------------- | ---------: | ------------------ |
| Go            |      1.26+ | build、test        |
| gofumpt       |     latest | formatter          |
| golangci-lint |     latest | lint               |
| Node.js / npm |     latest | bats統合テスト     |
| Git           |        2.x | 実処理と統合テスト |

```bash
go install mvdan.cc/gofumpt@latest
go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest
npm install
```

## 開発コマンド

| コマンド     | 内容                    |
| ------------ | ----------------------- |
| `make build` | `dist/dotfiles` を生成   |
| `make lint`  | golangci-lintを実行     |
| `make fmt`   | Goコードをgofumptで整形 |
| `make test`  | 全Goテストを実行        |
| `make bats`  | bats統合テストを実行    |
| `make clean` | build成果物を削除       |

依存更新後は `go mod tidy` を実行し、`go.mod` と `go.sum` を両方commitする。

## コーディング規約

- ユーザー向けメッセージは日本語にする。
- エラーは文脈を付けて返し、`internal/engine` から `os.Exit` を呼ばない。
- Git操作は `GitRunner` を通し、対象リポジトリとpathspecを明示する。
- ファイル権限は `0o755` のような8進リテラルで記述する。
- 外部入力となるパスやブランチ名は利用前に検証する。

### 命名規則

- 識別子は単数形を基本とする（ディレクトリ名・型名など、種類や型にラベルを貼るもの）。
- 以下は複数形を用いる。
    - iterable（配列など、複数要素を保持する変数）
    - 実務上の慣習として複数形が定着しているもの（REST URIなど）
    - 外部ツール・フレームワークが規約として読む名前（`skills/`、`commands/`、`.git/hooks/` など）
- 「集合を指すか」「型・種類にラベルを貼るか」で判断し、レイヤーや好みでは判断しない。

## テスト方針

- パース、バージョン比較、OS変換、メッセージ生成は単体テストする。
- Git操作は `t.TempDir()` 内の通常リポジトリとbareリポジトリで統合テストする。
- symlink、ファイル削除、hook発火は実データではなく隔離環境で検証する。
- Windows固有のファイル置換は、成功・復元・一時ファイル掃除を確認する。
- CLIエラーはプロセスを終了せず、Cobraコマンドの戻り値として検証する。

変更時の最低限の確認:

```bash
make fmt
make lint
make test
make bats
make build
```

## Git運用

- Conventional Commitsを使用する（`feat:`、`fix:`、`refactor:`、`docs:`、`test:`、`chore:` など）。
- `main` は安定版、開発は `feature/*` を基本とする。
- Git、symlink、削除を伴う手動検証は必ず一時ディレクトリで行う。

## 既知の制約

- Git hooksはbashを必要とする。
- Windowsのsymlink作成には開発者モードまたは適切な権限が必要。
- 旧 `sync.conf` と `link.yaml` は読み込まない。
- `template/README.md` はデータリポジトリ仕様の確定後に更新する。

## .agents/

| ディレクトリ       | 用途                         |
| ------------------ | ---------------------------- |
| `.agents/plan/`    | 複数セッション向けの実装計画 |
| `.agents/note/`    | 再利用する調査結果や判断記録 |
| `.agents/skills/`  | リポジトリ固有Skill          |
| `.agents/plugins/` | リポジトリ固有プラグイン情報 |

一時進捗やツール出力は保存しない。
