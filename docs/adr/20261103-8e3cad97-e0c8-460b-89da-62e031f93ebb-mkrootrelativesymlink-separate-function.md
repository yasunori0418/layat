---
id: "ADR-8e3cad97-e0c8-460b-89da-62e031f93ebb"
type: adr
name: "root 相対マーカーを別関数 mkRootRelativeSymlink で作る"
status: 採用
issues:
  - "#362"
origin: "Issue #362（src 側のプロジェクトルート実行時解決）の grilling（2026-10-12）。ADR-0059 §4 を書き直した"
justifies:
  - "REQ-a77181dd-a59c-40bb-855e-3430ad970eb6"
  - "REQ-1dcc9a33-b0f2-43e0-8310-fc4b19e68fe7"
  - "REQ-99ca5381-6c53-426c-b145-7b4297c53868"
references:
  - "ADR-73c24620-d2c0-4249-97fa-1b6296f1e79c"
  - "ADR-451679d5-473b-4faf-bf2d-a6f6d3e1e131"
  - "ADR-5e46ebd6-e44b-48dc-a93f-a8a3a3903967"
  - "ADR-6d2327f7-f513-42c3-89f8-808467db1914"
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
  - "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-8e3cad97-e0c8-460b-89da-62e031f93ebb: root 相対マーカーを別関数 mkRootRelativeSymlink で作る

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-73c24620-d2c0-4249-97fa-1b6296f1e79c, ADR-451679d5-473b-4faf-bf2d-a6f6d3e1e131, ADR-5e46ebd6-e44b-48dc-a93f-a8a3a3903967, ADR-6d2327f7-f513-42c3-89f8-808467db1914, ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd, ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0, `docs/spec.md`
- 起点: Issue #362（src 側のプロジェクトルート実行時解決）の grilling（2026-10-12）。ADR-0059 §4 を書き直した

## 背景

out-of-store は明示関数が返すマーカーでだけ選ばせ、値の形による暗黙の分岐は持たない
（ADR-73c24620-d2c0-4249-97fa-1b6296f1e79c）。既存の `mkOutOfStoreSymlink` は絶対パスを取る。
root 相対の `src` を表す手段を、この関数の引数の形で足すか、別の関数で足すかを決める必要がある。

## 決定

root 相対の `src` は、別の関数 `lib.mkRootRelativeSymlink` が返すマーカーで表す。この関数は
root 相対の文字列を包んだマーカーを返す純粋関数で、`mkOutOfStoreSymlink` は絶対パス専用のまま
据え置く。マーカーの形と判別は REQ-a77181dd-a59c-40bb-855e-3430ad970eb6 と
REQ-1dcc9a33-b0f2-43e0-8310-fc4b19e68fe7、`src` が取る種別は
REQ-99ca5381-6c53-426c-b145-7b4297c53868 に置く。

## 根拠

関数の名前で意図が明示され、絶対パスと root 相対のどちらを書いたつもりかが記述から読める。
`mkOutOfStoreSymlink` の受け付ける値を変えないので、既存の記述の意味も変わらない。

## 棄却した代替案

### `mkOutOfStoreSymlink` に相対の文字列も受けさせ、相対なら root 相対とみなす

絶対パスの書き忘れ（`"dotfiles/nvim"`）が、警告なしに root 相対として扱われる。型による暗黙の
分岐を廃したのと同じ問題を、引数の形で持ち込む。
