# conf.yaml の paths セクションを読み込むローダー — source して使う
# ローダーファイルのパスを基準に考える
INFRA="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# 設定ファイルから相対パスを取得
_rel() { grep "^[[:space:]]*$1:" "$INFRA/conf.yaml" | sed 's/.*:[[:space:]]*//'; }

# パスを変数に設定
DOTFILE="$INFRA/$(_rel dotfile)"
SYNC_YAML="$INFRA/$(_rel sync)"

# 一時変数を削除
unset -f _rel