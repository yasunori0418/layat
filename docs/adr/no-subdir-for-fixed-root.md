---
id: "ADR-433b9539-21d9-46c5-91a1-fc2b04d5eb47"
type: adr
name: "fixed root には subdir を提供しない"
status: stable
origin: "次期マイルストーン追加計画の grilling（2026-07-04）。ADR-0040 §1b を書き直した"
justifies:
  - "REQ-08c8f639-790c-4970-a954-c9567538cf9b"
references:
  - "ADR-5b347bfa-39d6-47f0-be59-1cf31dc25509"
  - "ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed"
  - "ADR-eeeb9dc8-9089-43ec-946a-c3ed76e98137"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-433b9539-21d9-46c5-91a1-fc2b04d5eb47: fixed root には subdir を提供しない

- ステータス: stable
- 日付: 2026-11-03
- 関連: [root マーカーに subdir の適用形を持たせる](root-marker-subdir-applied-form.md), [root は 3 つの marker で表す](three-root-markers.md), [layat.lib.mkEnv で buildEnv 合成環境を提供する](mkenv-buildenv.md), `docs/spec.md`
- 起点: 次期マイルストーン追加計画の grilling（2026-07-04）。ADR-0040 §1b を書き直した

## 背景

root マーカーには `subdir` の適用形を持たせると決めたが、root は 3 つのマーカーのほかに、評価時に
固定される絶対パス文字列（fixed root）も取る。fixed root にも同じ仕組みを用意するかが残る。

## 決定

fixed root には `subdir` の仕組みを提供しない。fixed root の下を基点にしたいときは、
`"/abs/path/sub"` のように連結したパスを root に書く。適用形の範囲は
[root マーカーの subdir 適用形は全 target への前置と等価な評価時の糖衣で、root を動かさない](../requirements/root-marker-subdir-sugar.md) に置く。

## 根拠

マーカーは実行時に解決するので、解決の結果と下へのずらしを分けて書く手段が要る。fixed root は
評価時に値が決まっているので、その問題が無い。連結で足りる。

## 棄却した代替案

### fixed root にも `subdir` を受けさせる

文字列の root を attrset との union に広げることになり、型と確認が増えるうえ、連結したパスと区別の
つかない同じ結果を得る書き方が 2 通りになるだけで、利用者が得るものは何も無い。
