---
id: "SDR-35636015-710a-4cf8-b6ba-48ca047efed0"
type: sara_adr
name: "design の satisfies 対象に test_plan を含める"
status: stable
sources:
  - id: "issue-583"
    resource: "https://github.com/yasunori0418/layat/issues/583"
    title: "#583"
  - id: "issue-237"
    resource: "https://github.com/yasunori0418/layat/issues/237"
    title: "#237"
  - id: "issue-203"
    resource: "https://github.com/yasunori0418/layat/issues/203"
    title: "#203"
origin: "epic #203 の grilling セッション（2026-08-08）で確定。ADR-0049 §4 を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
references:
  - "SDR-34916a83-0cff-4524-a8ff-5a3195af52c3"
  - "SDR-2457aa00-fab7-4c99-af53-2e6643c21c79"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T22:44:41+09:00"
---
# SDR-35636015-710a-4cf8-b6ba-48ca047efed0: design の satisfies 対象に test_plan を含める

- ステータス: stable
- 日付: 2026-11-03
- 関連: [テスト計画の型を品質方針の型から分けて設ける](quality-and-test-plan-types.md), [model.yaml で組み込みモデルを全面置換し、汎用の型名を使う](model-yaml-full-replacement.md), GitHub Issue #583, #237, #203, `docs/model.yaml`
- 起点: epic #203 の grilling セッション（2026-08-08）で確定。ADR-0049 §4 を書き直した

## 背景

design の `satisfies` は `requirement` だけを対象にしていた。テスト計画にあたる `requirement` を test_plan へ
移す（[テスト計画の型を品質方針の型から分けて設ける](quality-and-test-plan-types.md)）と、それらを `satisfies` で指していた design item の辺は
張り先を失う。中には `satisfies` がその 1 本だけの design item もあり、移した途端に design 側が接続漏れになる。

## 決定

design の親に test_plan を加え、`satisfies` の対象を `requirement` と `test_plan` にする。

## 根拠

テストハーネスの実装形を表す design は、`requirement` ではなくテスト計画を満たす実体である。これは
`satisfies` の語義に収まる。移す側は辺の張り先を新しい TP の ID へ替えるだけで済む。

## 棄却した代替案

### design 側の接続漏れを許容する

張り忘れではない接続漏れが残り、warning が張り忘れの信号としてはたらかなくなる。

### design から test_plan への専用 relation を新設する

`satisfies` と意味が重なる relation を 2 本持つことになる。design item を書くたびに、どちらを張るかを
判断しなければならない。
