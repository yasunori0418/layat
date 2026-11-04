---
id: "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
type: adr
name: "home mode の CLI と lib の版ずれを manifest の確認に任せて許容する"
status: 採用
origin: "実装前のレビュー第 2 巡で詰めた残りの細目。ADR-0016 §4 を書き直した"
justifies:
  - "REQ-14f0aec9-abae-4621-82f3-40536a1ad904"
references:
  - "ADR-560852e1-3d07-46b9-ae1a-d15f82d28a5e"
  - "ADR-f85f1afd-9b87-4f7a-922e-97c62554f198"
  - "ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed"
  - "ADR-88231ffc-4836-4252-be34-894944fd28c3"
  - "ADR-86a77232-dc99-4c04-bd6d-2127410e8828"
  - "ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246"
  - "ADR-e683ac05-c0e0-41c1-8b6a-38b605972fe1"
  - "ADR-28605860-770e-41df-95dd-7dc05b20497a"
  - "ADR-955c2601-88fb-4be8-85de-ec7d0a8c727d"
  - "ADR-65b3439a-b90a-4056-b55a-522399bc71e6"
  - "ADR-d15a216f-b5d6-4df9-9825-b23a341bec1d"
  - "ADR-553eea60-6ab4-4da7-93a9-3acf44b2fae3"
  - "ADR-5e46ebd6-e44b-48dc-a93f-a8a3a3903967"
---
# ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0: home mode の CLI と lib の版ずれを manifest の確認に任せて許容する

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-560852e1-3d07-46b9-ae1a-d15f82d28a5e, ADR-f85f1afd-9b87-4f7a-922e-97c62554f198, ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed, ADR-88231ffc-4836-4252-be34-894944fd28c3, ADR-86a77232-dc99-4c04-bd6d-2127410e8828, ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246, ADR-e683ac05-c0e0-41c1-8b6a-38b605972fe1, ADR-28605860-770e-41df-95dd-7dc05b20497a, ADR-955c2601-88fb-4be8-85de-ec7d0a8c727d, ADR-65b3439a-b90a-4056-b55a-522399bc71e6, ADR-d15a216f-b5d6-4df9-9825-b23a341bec1d, ADR-553eea60-6ab4-4da7-93a9-3acf44b2fae3, ADR-5e46ebd6-e44b-48dc-a93f-a8a3a3903967, `docs/spec.md`
- 起点: 実装前のレビュー第 2 巡で詰めた残りの細目。ADR-0016 §4 を書き直した

## 背景

project mode では、devShell に pin した `layat` を同梱するので、CLI と `layat.lib` が同じ flake の
入力から来る。home mode では CLI をグローバルに導入するので、CLI と、利用者の flake が pin する
`layat.lib` が別の入力になり、版がずれ得る。

## 決定

home mode の版ずれは強制的には防がない。CLI と flake が pin する `layat` を同じ入力から揃えることを
推奨として文書に書くに留める。ずれたときは、engine が manifest を読むときの確認に任せる。engine は対応版より
新しい `schemaVersion` を拒否し、同じ版の中で知らない列挙値も拒否する（ADR-b5431ecf-fe60-4b29-86af-9857478ce663）。

推奨の規範は REQ-14f0aec9-abae-4621-82f3-40536a1ad904 に置く。

## 根拠

版がずれても、版の新しい manifest と知らない列挙値は上の確認で明示的なエラーになる。
同じ入力から揃える運用で回避もできる。

home mode にも devShell や `nix run` 経由の起動を強制すれば版ずれは無くなるが、PATH に常駐する CLI の
使い勝手を捨てることになる。

## 棄却した代替案

### home mode でも devShell か `nix run` 経由の起動を強制する

版ずれは根絶できるが、PATH に常駐する CLI の使い勝手を損なう。
