---
id: "ADR-5d85f49f-c136-425f-bea4-ad8ab8dd8632"
type: adr
name: "root 相対 entry で engine が root を使うのは配置前の実在確認だけにする"
status: 採用
issues:
  - "#362"
origin: "Issue #362（src 側のプロジェクトルート実行時解決）の grilling（2026-10-12）。ADR-0059 §7 を書き直した"
justifies:
  - "DSG-fd39b9ee-a5d2-46c2-bdcd-fc399a1e0296"
references:
  - "ADR-b7f9b972-69b7-4bbe-affc-97978e1a55f7"
  - "ADR-5e46ebd6-e44b-48dc-a93f-a8a3a3903967"
---
# ADR-5d85f49f-c136-425f-bea4-ad8ab8dd8632: root 相対 entry で engine が root を使うのは配置前の実在確認だけにする

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-b7f9b972-69b7-4bbe-affc-97978e1a55f7, ADR-5e46ebd6-e44b-48dc-a93f-a8a3a3903967, `docs/design.md`
- 起点: Issue #362（src 側のプロジェクトルート実行時解決）の grilling（2026-10-12）。ADR-0059 §7 を書き直した

## 背景

root 相対マーカーのリンクは target 起点の相対パスで張るので、リンク先の文字列の計算に root は
要らない。一方、リンク先の実体が無いまま張ると、dangling な symlink ができる。out-of-store の
entry には、配置前にリンク先の実在を確かめる手順がある。

## 決定

リンク先を計算する `LinkDest` は、root を受け取らず Entry だけから計算する純粋関数のままとする。
engine が root 相対の entry で root を使うのは、配置前の確認で `<root>/<src>/<subpath>` の実在を
確かめ、無ければ dangling を作らずに止める箇所だけにする。計算と確認の細目は
DSG-fd39b9ee-a5d2-46c2-bdcd-fc399a1e0296 に置く。

## 根拠

`LinkDest` が純粋関数のままなら、planner・古くなった配置の除去・drift 修復・`--json` は root を持ち回らずに
同じ文字列で比較できる。root への依存を配置前の確認の 1 か所に閉じれば、root の値が変わったときに
影響を受ける箇所がそこだけになる。copy との併用は評価時に拒否しているので、copy の処理は触らない。

## 棄却した代替案

### `LinkDest` に root を渡す

比較に関わる全ての処理が root を受け取るように変わり、相対リンクにした利点が消える。

### 実在を確かめずに張る

リンク先の書き損じが、dangling な symlink としてエラーにならずに残る。
