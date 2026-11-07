---
id: "ADR-9e328362-2e90-441a-8eca-40081ce71b1b"
type: adr
name: "エラーはエンベロープに構造化し stderr のテキストと併存させる"
status: stable
issues:
  - "#552"
  - "yasunori0418/outturn#1"
origin: "layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §6 を書き直した"
justifies:
  - "REQ-a5053191-1c6a-449b-9c5e-5ff49dc5aead"
  - "REQ-9341fa5d-836e-4023-af53-cc7d273438d1"
  - "REQ-0a123b89-0399-4f76-b988-56a5f7e0becf"
references:
  - "ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:00:56+09:00"
---
# ADR-9e328362-2e90-441a-8eca-40081ce71b1b: エラーはエンベロープに構造化し stderr のテキストと併存させる

- ステータス: stable
- 日付: 2026-11-01
- 関連: ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71, GitHub Issue #552, yasunori0418/outturn#1
- 起点: layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §6 を書き直した

## 背景

layat はエラーを stderr の人間向けテキストで出す。outturn はエラーを `errors[]` と `item.error` として
規格の一部に持つ。stdout を JSON で専有しつつ、人が読む診断も残す必要がある。

## 決定

`--json` のエラーはエンベロープに構造化して載せ、同時に stderr の人間向けテキストも常に出す。置き場は
主体の有無で分ける。主体が決まる前の失敗はトップの `errors[]`、主体に紐づく失敗はその主体の
`results[i].errors[]`、item 起因の失敗は `item.error` に置く。outturn の `status` は、終了コード 0 で `success`、非 0 で
`error` に連動させる。`--json` は stdout に、`-v` の配置レポート（ADR-3dc21349-5504-48a4-ab88-812b92c3bdf9）は stderr に出るので、
両者は併用できる。

stderr との併存と `status` の連動は REQ-a5053191-1c6a-449b-9c5e-5ff49dc5aead に、置き場の細目は
REQ-9341fa5d-836e-4023-af53-cc7d273438d1 に、`-v` との直交は
REQ-0a123b89-0399-4f76-b988-56a5f7e0becf に置く。

## 根拠

outturn に適合するには、エラーをエンベロープに載せる必要がある。outturn §1 は stderr を診断チャネル
として認めているので、テキストを併存させても outturn に適合したままで、人間向けの読みやすさも保てる。

## 棄却した代替案

### エラーを stderr にだけ出し、エンベロープに載せない

outturn に適合しない。
