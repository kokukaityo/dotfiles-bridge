# dotfiles-bridge

> **Alpha 版です。** 実データへの適用前にバックアップを取ることを推奨します。
> フィードバックや不具合報告は [Issues](https://github.com/kokukaityo/dotfiles-bridge/issues) へお願いします。

設定ファイル（エージェントのルール、エディタの設定、シェルの dotfiles）はアプリごとに閉じて持っています。
そのため、こんな場面に出くわします。

- 新しい PC を買ったとき、設定を一から作り直す or 古い PC からコピーしてくる
- Claude Code の CLAUDE.md を育てたが、別のマシンでは使えない
- 同じ settings.json を VS Code と Cursor の両方に置きたいが、手動コピーが面倒

`dotfiles-bridge` はこれらを1箇所に集約し、symlink で各アプリに橋渡しします。単なる同期ツールではなく、外部に一切送信しない単一端末内のアプリ間共有も、複数端末間の設定同期も、同じ仕組みで対応できます。

## 特徴

- **一元管理** — 散らばった設定ファイルを任意のフォルダ（デフォルトは `~/dotfiles`）に集約します
- **OS 別配置** — symlink でアプリが期待するパスに自動配置します（Windows / macOS / Linux）。symlink なので編集が即座に反映されます
- **バージョン管理** — Git による自動 commit・push・pull で変更履歴を追跡できます。分岐時はローカル変更を安全に退避します
- **スコープの柔軟性** — 外部に一切送信しない単一端末内のアプリ間共有も、複数端末間の設定同期も、設定ファイルで柔軟に切り替えられます

## 既知の制約

- **Alpha 版です。** 実データへの適用前にバックアップを推奨します。
- 主に Windows で検証しています。macOS / Linux での動作は未検証です。
- watch サービスは `~/dotfiles` 前提の利用を推奨します。別パスでは手動起動か `DOTFILES_DIR` 環境変数の設定が必要です。
- link 先の親ディレクトリが存在しない場合、現状は自動作成します。この挙動は今後変更される可能性があります。
- CI・E2E テスト・installer の堅牢化は今後対応予定です。
- Windows で symlink を作成するには開発者モードの有効化が必要です（設定 → システム → 開発者向け）。

既知の不具合や今後の改善予定は [Issues](https://github.com/kokukaityo/dotfiles-bridge/issues) を参照してください。

## ユースケース

- **AI エージェント設定**: AGENTS.md・スキル・コマンドを Claude Code と Codex に同時配置。育てた設定がどの端末でもすぐ使えます
- **VS Code 設定**: settings.json を Git で管理し、変更履歴の追跡・復元が可能に。Cursor や Windsurf にも同時配置できます
- **シェル設定**: .bashrc や .zshrc を OS 別に管理し、マシン間で共有します
- **単一マシン内の共有**: リモート同期なし（`mode = "local"`）でも、1台の PC 内でアプリ間の設定を橋渡しできます

各ユースケースの詳細と `link.toml` の記述例は [doc/UseCase.md](doc/UseCase.md) を参照してください。

## 他のツールとの違い

dotfiles 管理ツールは複数あります。dotfiles-bridge は「symlink ベースのシンプルなアプローチ」を重視した設計です。

|                      | dotfiles-bridge                  | [chezmoi](https://www.chezmoi.io/) | [GNU Stow](https://www.gnu.org/software/stow/) | bare git repo |
| -------------------- | -------------------------------- | ---------------------------------- | ---------------------------------------------- | ------------- |
| **アプローチ**       | symlink                          | テンプレート展開（コピー）         | symlink                                        | Git 直接操作  |
| **OS 別配置**        | link.toml で定義                 | テンプレート内の分岐               | ディレクトリ構造で暗黙的                       | 手動          |
| **Git 統合**         | 自動 commit・push・pull          | 自動 commit・push・pull            | なし（別途管理）                               | 手動          |
| **テンプレート機能** | なし                             | あり（Go template）                | なし                                           | なし          |
| **シークレット管理** | pre-push hook で検出・ブロック   | 暗号化して同期可能                 | なし                                           | なし          |
| **ファイル監視**     | あり（自動 push）                | なし                               | なし                                           | なし          |
| **学習コスト**       | 低（設定ファイルをそのまま置く） | 中（テンプレート構文の習得）       | 低                                             | 低〜中        |
| **設定の即時反映**   | あり（symlink）                  | なし（`chezmoi apply` が必要）     | あり（symlink）                                | 手動          |

chezmoi はテンプレートエンジンやシークレット暗号化など、機能面では最も豊富です。dotfiles-bridge はテンプレート機能を持たない代わりに、設定ファイルをそのまま置けるシンプルさと、symlink による即時反映を重視しています。

## セキュリティ

dotfiles への API キーやトークンの混入を防ぐため、`.gitignore` 自動生成・pre-push シークレットスキャン・ignore カテゴリの 3 層で防御しています。`mode = "local"`（デフォルト）ではデータは一切外部に送信されません。

詳細は [SECURITY.md](SECURITY.md) を参照してください。脆弱性の報告もそちらに案内があります。

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

### 1. データリポジトリを作成する

```bash
dotfiles init ~/dotfiles
```

テンプレートからカテゴリ（ai-agent, shell, vscode）、設定ファイル、Git hooks が展開され、初回コミットが作成されます。

### 2. symlink を配置する

```bash
cd ~/dotfiles
dotfiles link
```

各カテゴリの `link.toml` に基づいて、OS に応じた symlink を配置します。配置先に既存ファイルがある場合は自動的にバックアップされます。

この時点でローカルでの一元管理が使えます。設定ファイルを編集すると、symlink 経由で各アプリに即座に反映されます。

### 3. リモートに接続する（任意）

複数マシン間で同期する場合は、リモートリポジトリを接続し、`sync.toml` の `mode` を `"remote"` に変更します。ローカルのみで使う場合はこの手順は不要です。

```bash
cd ~/dotfiles
# sync.toml: mode = "remote" に変更

git remote add origin git@github.com:<user>/<repo>.git
git push -u origin main
```

### 別のマシンで使う

```bash
git clone git@github.com:<user>/<repo>.git ~/dotfiles
export DOTFILES_DIR="$HOME/dotfiles"
dotfiles install
```

`dotfiles install` は hooks・gitignore の設定と symlink の配置をまとめて行います。

### 複数マシン間の同期フロー

`mode = "remote"` の場合、設定の変更は以下の流れで同期されます。

```
PC-A で設定を編集
  → dotfiles push（または watch で自動 push）
  → PC-B で dotfiles pull
  → symlink 経由で即反映
```

## コマンド

| コマンド                          | 説明                                         |
| --------------------------------- | -------------------------------------------- |
| `dotfiles init [path]`            | データリポジトリを作成（既定: `~/dotfiles`） |
| `dotfiles install`                | hooks・gitignore 設定と symlink 配置         |
| `dotfiles link`                   | OS に応じた symlink を配置                   |
| `dotfiles pull`                   | リモートから同期                             |
| `dotfiles push`                   | auto カテゴリの変更を commit して push       |
| `dotfiles watch`                  | ファイル変更を監視して自動 push              |
| `dotfiles delete-category <name>` | カテゴリを設定と Git 履歴から削除            |
| `dotfiles gitignore`              | `.gitignore` の自動生成部分を更新            |
| `dotfiles status`                 | コンフリクト退避状態を表示                   |
| `dotfiles version`                | バージョン情報を表示                         |

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

### local モードと remote モードの違い

`sync.toml` の `mode` で切り替えます。

- `"local"`（デフォルト）: `dotfiles push` は Git commit のみ。リモートへの push は行いません。1台の PC 内で設定を集約・管理するのに適しています
- `"remote"`: commit に加え、origin への push / pull も行います。複数マシン間で設定を同期する場合に使います

## 開発

開発に参加する場合は [doc/Development.md](doc/Development.md) を参照してください。
設計の詳細は [doc/Architecture.md](doc/Architecture.md) にあります。
プロジェクトの成り立ちと設計判断は [doc/Story.md](doc/Story.md) に記録しています。

## Support

このプロジェクトが役に立ったら、コーヒーを奢ってもらえると励みになります。

<a href="https://www.buymeacoffee.com/kokukaityo" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me A Coffee" height="50"></a>

## ライセンス

[MIT](LICENSE)
