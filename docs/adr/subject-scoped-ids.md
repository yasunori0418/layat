---
id: "ADR-79af04fe-7ba0-4953-bb0d-75bda1707cd7"
type: adr
name: "--json の item id は subject スコープで導く"
status: stable
sources:
  - id: "issue-552"
    resource: "https://github.com/yasunori0418/layat/issues/552"
    title: "#552"
  - id: "issue-outturn-1"
    resource: "https://github.com/yasunori0418/outturn/issues/1"
    title: "yasunori0418/outturn#1"
origin: "layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §3 を書き直した"
justifies:
  - "REQ-57137302-de29-4f71-a565-034cd5de080b"
references:
  - "ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-01T23:01:38+09:00"
---
# ADR-79af04fe-7ba0-4953-bb0d-75bda1707cd7: --json の item id は subject スコープで導く

- ステータス: stable
- 日付: 2026-11-01
- 関連: [--json は outturn 規約に常時準拠する](outturn-conformance-principle.md), GitHub Issue #552, yasunori0418/outturn#1, outturn ADR-0013
- 起点: layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §3 を書き直した

## 背景

outturn は item の `id` を、identity（`{kind, key}`）を JCS で正規化して SHA-256 した値で導く。
layat の entry は config ごとに独立していて、別の config に同じ target があることもある。key に何を
含めるかで、id がどの範囲で一意になるかが決まる。

## 決定

entry の identity は `kind="entry"`、`key={target}`（root 相対の target）とし、config 名を key に
含めない。id は subject をまたいで衝突してよく、消費側は `(tool.name, subject, id)` の 3 つ組で参照を
解決する。この 3 つ組が常に揃うよう、`subject`（config 名）はすべての `SubjectResult` で必須にする。

導出式は [item id は identity の JCS を SHA-256 した小文字 hex とする](../requirements/json-item-id-derivation.md) に置く。

## 根拠

outturn#1 で確定した参照キー規約の 3 層化に従う。subject を id の導出に混ぜないので、id は subject と
同じ情報を二重に持たない。

## 棄却した代替案

### config 名を key に含め、id を全体で一意にする

subject と id が同じ情報を二重に持ち、参照キー規約の 3 層と食い違う。
