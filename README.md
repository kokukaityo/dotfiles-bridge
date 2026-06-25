# dotfiles-bridge

> **Alpha 版です。** 実データへの適用前にバックアップを取ることを推奨します。
> フィードバックや不具合報告は [Issues](https://github.com/kokukaityo/dotfiles-bridge/issues) へお願いします。

本来アプリごとに閉じている設定ファイルを、アプリ間・マシン間で安全に橋渡しする Go 製の dotfiles 管理ツールです。

CLI コマンド名は `dotfiles` です。

## dotfiles-bridge とは

開発環境の設定ファイル — AI エージェントのルール、エディタの設定、シェルの dotfiles — はそれぞれのアプリがローカルに閉じて持っています。そのため、こんな場面に出くわします。

- 新しい PC を買ったとき、設定を一から作り直す
- Claude Code の AGENTS.md を育てたが、別のマシンでは使えない
- 同じ settings.json を VS Code と Cursor の両方に置きたいが、手動コピーが面倒

dotfiles-bridge はこれらの設定ファイルを1箇所に集約し、symlink で各アプリに配置します。設定ファイルは本来それぞれのアプリに閉じているもので、それを「橋渡し」するのが dotfiles-bridge の役割です。

### 特徴

- **一元管理** — 散らばった設定ファイルを `~/dotfiles` に集約します
- **OS 別配置** — symlink でアプリが期待するパスに自動配置します（Windows / macOS / Linux）。symlink なので編集が即座に反映されます
- **バージョン管理** — Git による自動 commit・push・pull で変更履歴を追跡できます。分岐時はローカル変更を安全に退避します
- **スコープの柔軟性** — 1台の PC 内のアプリ間共有も、複数マシン間の同期も、同じ仕組みで対応します

## ユースケース

- **AI エージェント設定**: AGENTS.md・スキル・コマンドを Claude Code と Codex に同時配置。育てた設定がどの端末でもすぐ使えます
- **VS Code 設定**: settings.json を Git で管理し、変更履歴の追跡・復元が可能に。Cursor や Windsurf にも同時配置できます
- **シェル設定**: .bashrc や .zshrc を OS 別に管理し、マシン間で共有します
- **単一マシン内の共有**: リモート同期なし（`mode = "local"`）でも、1台の PC 内でアプリ間の設定を橋渡しできます

各ユースケースの詳細と `link.toml` の記述例は [doc/UseCase.md](doc/UseCase.md) を参照してください。

## 他のツールとの違い

dotfiles 管理ツールは複数あります。dotfiles-bridge は「symlink ベースのシンプルなアプローチ」を重視した設計です。

| | dotfiles-bridge | [chezmoi](https://www.chezmoi.io/) | [GNU Stow](https://www.gnu.org/software/stow/) | bare git repo |
|---|---|---|---|---|
| **アプローチ** | symlink | テンプレート展開（コピー） | symlink | Git 直接操作 |
| **OS 別配置** | link.toml で定義 | テンプレート内の分岐 | ディレクトリ構造で暗黙的 | 手動 |
| **Git 統合** | 自動 commit・push・pull | 自動 commit・push・pull | なし（別途管理） | 手動 |
| **テンプレート機能** | なし | あり（Go template） | なし | なし |
| **シークレット管理** | pre-push hook で検出・ブロック | 暗号化して同期可能 | なし | なし |
| **ファイル監視** | あり（自動 push） | なし | なし | なし |
| **学習コスト** | 低（設定ファイルをそのまま置く） | 中（テンプレート構文の習得） | 低 | 低〜中 |
| **設定の即時反映** | あり（symlink） | なし（`chezmoi apply` が必要） | あり（symlink） | 手動 |

chezmoi はテンプレートエンジンやシークレット暗号化など、機能面では最も豊富です。dotfiles-bridge はテンプレート機能を持たない代わりに、設定ファイルをそのまま置けるシンプルさと、symlink による即時反映を重視しています。

## セキュリティ

dotfiles に API キーやトークンが混入するリスクに対し、3層の防御を備えています。

1. **`.gitignore` 自動生成**: `dotfiles gitignore` で `*auth*`, `*.key`, `*.pem`, `.env*` 等を自動的に除外します
2. **pre-push hook のシークレットスキャン**: push 前に差分をスキャンし、`api_key`, `secret`, `password`, `token`, 秘密鍵ヘッダーなどのパターンを検出した場合は push をブロックします
3. **ignore カテゴリ**: `sync.toml` の `ignore` に登録したカテゴリは Git 追跡から完全に除外されます

誤検知で push がブロックされた場合は `SKIP_SECRET_SCAN=1 git push` で回避できます。

また、`mode = "local"`（デフォルト）であればデータは一切外部に送信されません。ローカルの Git コミットのみで動作するため、リモートリポジトリなしで完結できます。

リモート同期を使う場合は、リポジトリを **private** に設定することを推奨します。

技術的な詳細は [doc/Architecture.md](doc/Architecture.md) を参照してください。

## インストール

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/kokukaityo/dotfiles-bridge/main/install.sh | bash
```

インストール先を変更する場合:

```bash
INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/kokukaityo/dotfiles-bridge/main/install.sh | bash
```

### Windows

[GitHub Releases](https://github.com/kokukaityo/dotfiles-bridge/releases) から zip をダウンロードし、展開した `dotfiles.exe` を PATH の通ったディレクトリに配置してください。

### ソースからビルド

Go 1.26 以降と Git 2.x が必要です。

```bash
git clone https://github.com/kokukaityo/dotfiles-bridge.git
cd dotfiles-bridge
make build
# dist/dotfiles が生成される
```

## クイックスタート

### 新しく始める

```bash
# データリポジトリを作成
dotfiles init ~/dotfiles
cd ~/dotfiles

# リモートに接続（任意 — local モードなら不要）
git remote add origin git@github.com:<user>/<repo>.git
git push -u origin main
```

`dotfiles init` を実行すると、テンプレートからカテゴリ（ai-agent, shell, vscode）、設定ファイル、Git hooks が展開されます。`dotfiles link` も自動実行され、symlink が配置されます。

### 別のマシンで使う

```bash
git clone git@github.com:<user>/<repo>.git ~/dotfiles
export DOTFILES_DIR="$HOME/dotfiles"
dotfiles install
```

### 複数マシン間の同期フロー

`mode = "remote"` の場合、設定の変更は以下の流れで同期されます。

```
PC-A で設定を編集
  → dotfiles push（または watch で自動 push）
  → PC-B で dotfiles pull（またはシェル起動時に自動 pull）
  → symlink 経由で即反映
```

### シェル起動時に自動同期（任意）

`~/.bashrc` や `~/.zshrc` に追加:

```bash
export DOTFILES_DIR="$HOME/dotfiles"
command -v dotfiles >/dev/null && dotfiles pull
command -v dotfiles >/dev/null && dotfiles status
```

`dotfiles install` を実行すると、ファイル監視サービスが OS のログイン時自動起動に登録されます。以降は設定ファイルの変更が自動で commit・push されます。

## コマンド

| コマンド                          | 説明                                       |
| --------------------------------- | ------------------------------------------ |
| `dotfiles init [path]`            | データリポジトリを作成（既定: `~/dotfiles`）|
| `dotfiles install`                | hooks・gitignore 設定と symlink 配置       |
| `dotfiles link`                   | OS に応じた symlink を配置                 |
| `dotfiles pull`                   | リモートから同期                           |
| `dotfiles push`                   | auto カテゴリの変更を commit して push     |
| `dotfiles watch`                  | ファイル変更を監視して自動 push            |
| `dotfiles delete-category <name>` | カテゴリを設定と Git 履歴から削除          |
| `dotfiles gitignore`              | `.gitignore` の自動生成部分を更新          |
| `dotfiles status`                 | コンフリクト退避状態を表示                 |
| `dotfiles version`                | バージョン情報を表示                       |

データリポジトリは次の順で解決されます:

1. 環境変数 `DOTFILES_DIR`
2. 現在の Git ルート（`sync.toml` がある場合）
3. `~/dotfiles`

## データリポジトリ

`dotfiles init` で作成されるデータリポジトリの構成:

```text
~/dotfiles/
├── sync.toml             # 同期設定
└── <category>/           # カテゴリディレクトリ（自由に追加）
    ├── link.toml         # symlink 定義
    └── (設定ファイル群)
```

設定ファイル（`sync.toml`・`link.toml`）の詳細は、データリポジトリ内の README（[template/README.md](template/README.md)）を参照してください。

## コンフリクト

pull 時にローカルとリモートが分岐していた場合、自動 merge は行いません。dotfiles は個人の正本であり、自動 merge で意図しない内容が紛れ込むリスクを避けるためです。

ローカル側を `conflict/<hostname>/<timestamp>` ブランチへ退避し、既定ブランチをリモートへ戻します。`conflict/*` ブランチが残っている間、`dotfiles status` が警告を表示します。

解消手順:

```bash
cd ~/dotfiles
dotfiles status                       # 退避ブランチを確認
git cherry-pick conflict/xxx/xxx      # 必要な変更を取り込む
git branch -d conflict/xxx/xxx        # 退避ブランチを削除
dotfiles push                         # 解消結果を push
```

## Tips / FAQ

### コマンド名 `dotfiles` と `~/dotfiles` ディレクトリで補完が被る

bash と fish ではコマンドが優先されるため、実質問題ありません。zsh では候補が2つ出ることがありますが、Tab で選択できます。

気になる場合はエイリアスで回避できます:

```bash
alias dfb=dotfiles
```

### Windows で symlink を作成できない

Windows で symlink を作成するには、開発者モードの有効化が必要です。

設定 → システム → 開発者向け → 開発者モード を ON にしてください。

### push が pre-push hook でブロックされた

pre-push hook がシークレットらしきパターンを検出しています。差分を確認し、API キーやトークンが含まれていないか確認してください。

誤検知の場合:

```bash
SKIP_SECRET_SCAN=1 git push
```

### local モードと remote モードの違い

`sync.toml` の `mode` で切り替えます。

- `"local"`（デフォルト）: `dotfiles push` は Git commit のみ。リモートへの push は行いません。1台の PC 内で設定を集約・管理するのに適しています
- `"remote"`: commit に加え、origin への push / pull も行います。複数マシン間で設定を同期する場合に使います

## 既知の制約

- **Alpha 版です。** 実データへの適用前にバックアップを推奨します。
- 主に Windows で検証しています。macOS / Linux での動作は未検証です。
- watch サービスは `~/dotfiles` 前提の利用を推奨します。別パスでは手動起動か `DOTFILES_DIR` 環境変数の設定が必要です。
- link 先の親ディレクトリが存在しない場合、現状は自動作成します。この挙動は今後変更される可能性があります。
- CI・E2E テスト・installer の堅牢化は今後対応予定です。
- Windows で symlink を作成するには開発者モードの有効化が必要です（設定 → システム → 開発者向け）。

## 開発

開発に参加する場合は [doc/Development.md](doc/Development.md) を参照してください。
設計の詳細は [doc/Architecture.md](doc/Architecture.md) にあります。
プロジェクトの成り立ちと設計判断は [doc/Story.md](doc/Story.md) に記録しています。

## ライセンス

[MIT](LICENSE)
