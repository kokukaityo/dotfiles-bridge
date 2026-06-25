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
4. 親 plan（masterplan・phase plan）を更新する。
    - アーカイブした plan を参照している `_` 接頭辞付き plan（`_masterplan.md`, `_phase*.md` など）を探す。
    - 該当行の「状態」列を実態に合わせて更新する（例: 「完了」「archive済み」「中止」）。
    - リンク先をアーカイブ後のパスに書き換える（例: `[plan.md](plan.md)` → `[plan.md](../archive/plan.md)`）。
    - phase plan 内の全タスクがアーカイブ済みになった場合は、masterplan 側の該当 phase の状態も更新する。
5. 更新内容をユーザーに報告する。

## 注意

- 移動後にリンク切れが生じる場合は報告する。
