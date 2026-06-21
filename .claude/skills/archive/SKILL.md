---
name: archive
description: 完了済みの plan や不要になった note を .agents/archive/ に移動する
user_instructions: 対応済みのplanやレガシーなnoteをarchiveに移動して
---

# archive スキル

完了済みの plan や、役割を終えた note を `.agents/archive/` に移動する。

## 手順

1. `.agents/plan/` と `.agents/note/` のファイルを読み、アーカイブ対象を判定する。
   - plan: 全タスクが完了、または中止・不要になったもの
   - note: 参照先の plan がアーカイブ済み、または内容が現状と乖離しているもの
2. 対象をユーザーに提示し、確認を取る。判断に迷うものは保留として報告する。
3. 承認されたファイルを `.agents/archive/` にそのまま移動する。

## 注意

- 移動後にリンク切れが生じる場合は報告する。
