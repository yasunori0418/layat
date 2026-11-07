---
id: "ADR-963df3dd-ac51-41fc-94e2-a89c0675dc23"
type: adr
name: "版 1 しか読めない engine との非互換を受け入れる"
status: 採用
issues:
  - "#580"
  - "#498"
origin: "Issue #498（epic）の grilling（2026-10-31）。ADR-0070 §5a を書き直した"
justifies:
  - "REQ-ee5085a5-077b-4f5a-aa6a-841bc5515282"
references:
  - "ADR-8e253f2e-c255-482a-a0f8-0837e122e80f"
  - "ADR-767c33e5-73eb-4ef0-aba7-c87673fa5c2e"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-963df3dd-ac51-41fc-94e2-a89c0675dc23: 版 1 しか読めない engine との非互換を受け入れる

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-8e253f2e-c255-482a-a0f8-0837e122e80f, ADR-767c33e5-73eb-4ef0-aba7-c87673fa5c2e, GitHub Issue #580, #498
- 起点: Issue #498（epic）の grilling（2026-10-31）。ADR-0070 §5a を書き直した

## 背景

記録の版を 2 に上げたので、版 1 しか読めない engine は版 2 の記録で error になる。home-manager が
pin した `layatPackage` と PATH 上の CLI でバージョンがずれると、この組み合わせが起きる。

## 決定

版 1 しか読めない engine が版 2 の記録で止まることを、pre-1.0 の既知の事項として受け入れる。

## 根拠

読みを緩めて未知の版を「記録なし」に倒すと、誤った比較の基準で古くなった配置を除去する。
不正な記録で止める方針は変えず、版のずれは pre-1.0 の範囲で受け入れる。止まった利用者は、
両方のバージョンを揃えれば解消できる。

## 棄却した代替案

### 未知の版の記録を「記録なし」として読む

誤った比較の基準で古くなった配置を除去し、利用者の配置を壊しうる。
