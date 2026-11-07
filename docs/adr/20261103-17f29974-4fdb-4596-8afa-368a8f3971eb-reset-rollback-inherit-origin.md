---
id: "ADR-17f29974-4fdb-4596-8afa-368a8f3971eb"
type: adr
name: "記録を書き直す操作は記録済みの出所を引き継ぐ"
status: stable
issues:
  - "#580"
  - "#498"
origin: "Issue #498（epic）の grilling（2026-10-31）。ADR-0070 §4 を書き直した"
justifies:
  - "REQ-ee5085a5-077b-4f5a-aa6a-841bc5515282"
references:
  - "ADR-8e253f2e-c255-482a-a0f8-0837e122e80f"
  - "ADR-0db061d1-fa30-436d-91d7-c0cec8804c52"
  - "ADR-767c33e5-73eb-4ef0-aba7-c87673fa5c2e"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-17f29974-4fdb-4596-8afa-368a8f3971eb: 記録を書き直す操作は記録済みの出所を引き継ぐ

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-8e253f2e-c255-482a-a0f8-0837e122e80f, ADR-0db061d1-fa30-436d-91d7-c0cec8804c52, ADR-767c33e5-73eb-4ef0-aba7-c87673fa5c2e, GitHub Issue #580, #498
- 起点: Issue #498（epic）の grilling（2026-10-31）。ADR-0070 §4 を書き直した

## 背景

reset と rollback も記録を書き直すが、どちらの操作も新しい manifest を生成しないので、apply
と違って今回の出所を持たない。

## 決定

reset と rollback は、記録を書き直すときに記録済みの出所を引き継ぐ。記録に出所が無ければ、
出所を省いて版 2 で書く。

## 根拠

どちらの操作も、出所を新しく決める材料を持たない。記録済みの値を保てば、情報を失わない。
出所を省いて書いた記録は、次の apply で出所を比べずに今回の出所が書き込まれる。

## 棄却した代替案

### 書き直すときに出所を消す

次の apply が出所を比べられず、別の出所からの上書きに気付けなくなる。

### rollback では戻り先の世代から出所を読む

世代は出所を持たない。出所を持つのは記録だけである。
