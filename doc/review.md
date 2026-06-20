## レビュー結果

### 重大

1. `sync.sh` が実質動作しません  
   [env.sh](C:/usr/dotfile/master/.infra/env.sh:5) の `_rel sync` が、階層を無視して全 `sync:` を取得します。

実測値:

```text
SYNC_YAML=".../conf.yaml
manual
auto
auto
auto
ignore
ignore"
```

そのため `pull`、`push`、`gitignore` は設定ファイルを読めず失敗します。`path.sync` セクションだけを取得する実装が必要です。

### 高

2. 自動pushが手動変更までコミットします  
   [sync.sh](C:/usr/dotfile/master/.infra/sync.sh:188) ではautoカテゴリを追加した後、[201行目](C:/usr/dotfile/master/.infra/sync.sh:201)でステージ全体をコミットしています。

事前にステージしていた `.infra/` やREADMEまで自動pushされます。autoカテゴリにpathspecを限定すべきです。

3. autoカテゴリ自体を削除すると削除が同期されません  
   [sync.sh](C:/usr/dotfile/master/.infra/sync.sh:189) はディレクトリが存在する場合だけ `git add` します。カテゴリディレクトリ全体の削除はステージされません。

4. ローカルが単純にaheadでも「競合」扱いになります  
   [sync.sh](C:/usr/dotfile/master/.infra/sync.sh:136) は以下を区別していません。

- 本当の分岐
- ローカルだけに未pushコミットがある状態

後者でもconflictブランチへ退避し、[161行目](C:/usr/dotfile/master/.infra/sync.sh:161)でmainをhard resetします。push失敗後の次回pullなどで発生します。

### 中

5. READMEと実装が大きくずれています  
   [README.md](C:/usr/dotfile/master/README.md:24)にはファイル監視・debounce・自動pushが記載されていますが、監視処理は存在しません。また多数の箇所が旧名 `.infra/sync.yaml` のままで、実体は `.infra/conf.yaml` です。

6. pre-pushは過去コミット内の秘密を検出できません  
   [pre-push](C:/usr/dotfile/master/.infra/hook/pre-push:19)は現在のツリーだけを検査します。一度コミットして次のコミットで削除した秘密は、push対象履歴に残っていても通過します。

7. setup後に未追跡ファイルが残ります  
   [setup.sh](C:/usr/dotfile/master/.infra/setup.sh:19)が `.gitattributes` を生成しますが、現在リポジトリには存在せず、ignoreもされません。初回setup直後に作業ツリーがdirtyになります。

8. post-mergeが失敗時にも成功表示します  
   [post-merge](C:/usr/dotfile/master/.infra/hook/post-merge:7)で `link.sh` が失敗しても、続けて「symlinks refreshed」と表示されます。

## 検証状況

- 作業ツリー: クリーン
- Bash構文検査: 全ファイル正常
- `env.sh`不具合: Git Bashで再現済み
- pre-push単体実行: 現在の内容では通過
- テストコード・CI・ShellCheck設定: なし
- ファイル変更: なし

最優先は `env.sh` の修正です。現状では同期機構の入口が壊れています。

.commandsは非推奨でagentsに統一されてるはず
