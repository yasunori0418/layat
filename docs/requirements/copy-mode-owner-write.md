---
id: "REQ-84e3c717-adf5-4ff3-b0db-d039b82ef19c"
type: requirement
name: "copy は元の mode を保存しつつ owner-write を付与する"
status: stable
derives_from:
  - "UC-403fbe32-b146-401b-8b53-fe67c1e169c5"
specification: |
  A copy SHALL preserve the mode of the source while adding owner-write to it (for
  instance `0444` becoming `0644` and `0555` becoming `0755`). Since store paths are
  read-only, preserving the mode as it stands would leave the result uneditable, whereas
  the purpose of a copy is for the user to own and edit it detached from the store; the
  relative structure of the permissions — the execute bit and the group / other bits —
  SHALL therefore be kept while the owner is left able to edit.
specification_ja: |
  copy は元の mode を保存しつつ owner-write を付与しなければならない（例: `0444` → `0644` /
  `0555` → `0755`）。store パスは read-only のためそのまま保存すると編集できないが、copy は
  「store から切り離してユーザーが所有・編集する」用途である。したがって perm の相対構造
  （実行ビット・group / other）は保ちつつ、所有者が編集できる状態にしなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# REQ-84e3c717-adf5-4ff3-b0db-d039b82ef19c: copy は元の mode を保存しつつ owner-write を付与する

## 仕様

**mode は保存しつつ owner-write を付与する**（例: `0444 → 0644` / `0555 → 0755`）。
store パスは read-only（0444 / 0555）のため、そのまま保存すると編集できない。copy は
「store から切り離してユーザーが所有・編集する」用途なので、perm の相対構造（実行ビット・
group/other）は保ちつつ所有者が編集できる状態にする。

> **上は原文の写しで、規範は frontmatter が正**。この mode 規則は通常の place-once コピーと
> `apply --recopy` の再コピーの双方に適用される。place-once そのものは [copy は target 不在のときだけマテリアライズする place-once で世代管理の対象外とする](copy-place-once.md)、
> recopy そのものは [apply --recopy は config 内の全 copy target を src から無条件に上書き再コピーする](apply-recopy.md) / [recopy の上書きは削除ではなく同一親内への rename 退避で行う](recopy-rename-aside.md) の担当。

## 出典

`docs/spec.md`「配置動作仕様」→「copy モード（place-once・ユーザー管理）」節の箇条書き第 4 項
（および同節コードブロック・「recopy」節の「mode 保存 + owner-write 付与」）。

決定の実体は [copy は再帰コピーで mode を保存し owner-write にする](../adr/copy-recursive-mode-preserving-owner-write.md)「copy は再帰コピーで mode を保存し owner-write にする」。
