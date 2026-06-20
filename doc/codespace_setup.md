## セットアップまとめ

### 構成図

```
GitHub リポジトリ
    ↕                        ↕
PC の VSCode          GitHub Codespaces
（ローカルで直接開発）   └ Claude Code（常駐）
                              ↕
                      スマホの Claude アプリ
                      （外出先でPlan/doc作業）
```

---

### やること一覧

**① リポジトリに `.devcontainer/devcontainer.json` を追加**（済）

```jsonc
{
    // postCreateCommand は動作未確認。当面は手動インストール。
    // "postCreateCommand": "npm install -g @anthropic-ai/claude-code",
    "postStartCommand": "nohup claude &"
}
```

**② Codespace を作成**（初回のみ）

- GitHub リポジトリ → Code → Codespaces → New codespace
- 作成後、ターミナルで `npm install -g @anthropic-ai/claude-code` を手動実行

**③ 認証**（初回のみ）

方法A: claude.ai アカウントでログイン

- Codespace のターミナルで `claude auth login` を実行
- claude.ai アカウントで認証（APIキーではなくOAuth）
- Codespace を再作成するとやり直しが必要

方法B: Codespaces Secrets を使う（推奨）

- GitHub → Settings → Codespaces → Secrets で `ANTHROPIC_API_KEY` を設定
- Secrets は自分のGitHubアカウントに紐づくので、他人がCodespaceを作っても渡らない
- Codespace を再作成しても自動で環境変数として注入される

**④ Remote Control を全セッションで有効化**（初回のみ）

- `claude` を起動して `/config` → `Enable Remote Control for all sessions` を `true` に設定
- `postStartCommand` で `claude remote-control` が毎回起動するので、この設定と合わせて使う

---

### 日常の使い方

| 状況                     | やること                                                            |
| ------------------------ | ------------------------------------------------------------------- |
| 外出前                   | 特に何もしなくてOK（30分アイドルで自動停止）                        |
| 外出先でつなぎたい       | Claude アプリ → Code タブ → `my-session` をタップ                   |
| セッションが切れていたら | GitHub アプリ → Codespaces → 起動 → 自動で Claude Code も立ち上がる |
| PCで開発したい           | VSCode でローカルリポジトリをそのまま編集・push                     |
| 話題を切り替えたい       | `/clear` でコンテキストをリセット                                   |

---

### 注意点

- Codespace の無料枠は月120時間（2コア）なので使わないときは停止推奨
- `postStartCommand` はバックグラウンド実行（`nohup ... &`）で対応済み
- 追加セッションが必要な場合はターミナルで `claude &` を手動実行
