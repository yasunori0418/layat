---
id: "ADR-bd8a4c8a-5e7e-485a-bccf-20dab63125f2"
type: adr
name: "CLI は outturn/go に依存する"
status: 採用
issues:
  - "#552"
  - "yasunori0418/outturn#1"
origin: "layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §8 の Go 依存を書き直した"
justifies:
  - "REQ-2ea19863-eaa2-466b-b1ed-3f56f6417c62"
  - "REQ-57137302-de29-4f71-a565-034cd5de080b"
  - "REQ-637599dc-a1ec-4af5-9e97-e882c7df56d0"
references:
  - "ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71"
  - "ADR-79af04fe-7ba0-4953-bb0d-75bda1707cd7"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-01T23:14:40+09:00"
---
# ADR-bd8a4c8a-5e7e-485a-bccf-20dab63125f2: CLI は outturn/go に依存する

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71, ADR-79af04fe-7ba0-4953-bb0d-75bda1707cd7, GitHub Issue #552, yasunori0418/outturn#1
- 起点: layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §8 の Go 依存を書き直した

## 背景

エンベロープの形と item id の導出（JCS と SHA-256）は、outturn の規格が定める。outturn は Go の参照実装
`github.com/yasunori0418/outturn/go` を、id 導出の適合ベクタと合わせて提供している。engine は第三者依存を
持たず、CLI 層は許可した最小限の第三者依存だけを持つ。

## 決定

CLI 層は `github.com/yasunori0418/outturn/go` に依存し、エンベロープの汎用型と `DeriveID` を使う。
この依存は CLI 層に閉じる。`engine.Result` などを直接 marshal せず、DTO を通して outturn の型へ
詰め替える。

DTO を経た payload の生成は REQ-2ea19863-eaa2-466b-b1ed-3f56f6417c62 に、id の導出は
REQ-57137302-de29-4f71-a565-034cd5de080b に、CLI の依存の許可範囲は REQ-637599dc-a1ec-4af5-9e97-e882c7df56d0 に置く。

## 根拠

id の導出を規格の実装と共有すれば、適合ベクタ（`id-vectors.json`）とずれる危険が下がる。DTO を挟むので、
engine の型は出力の契約から切り離される。

## 棄却した代替案

### エンベロープ型と id の導出を layat の中で自前実装する

依存は増えないが、規格が改訂されるたびに追従が要り、適合ベクタとずれる危険を抱える。

### engine の結果型を直接 marshal する

engine の内部構造を変えると、出力の契約まで変わってしまう。
