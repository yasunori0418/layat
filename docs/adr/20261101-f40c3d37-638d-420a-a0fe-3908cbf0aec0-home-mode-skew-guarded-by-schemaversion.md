---
id: "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
type: adr
name: "home mode の CLI と lib の版ずれを manifest の検査に任せて許容する"
status: 採用
origin: "実装前のレビュー第 2 巡で詰めた残りの細目。ADR-0016 §4 を書き直した"
justifies:
  - "REQ-14f0aec9-abae-4621-82f3-40536a1ad904"
references:
  - "ADR-0007"
  - "ADR-0059"
---
# ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0: home mode の CLI と lib の版ずれを manifest の検査に任せて許容する

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-0007, ADR-0059, `docs/spec.md`
- 起点: 実装前のレビュー第 2 巡で詰めた残りの細目。ADR-0016 §4 を書き直した

## 背景

project mode では、devShell に pin した `layat` を同梱するので、CLI と `layat.lib` が同じ flake の
入力から来る。home mode では CLI をグローバルに導入するので、CLI と、利用者の flake が pin する
`layat.lib` が別の入力になり、版がずれ得る。

## 決定

home mode の版ずれは強制的には防がない。CLI と flake が pin する `layat` を同じ入力から揃えることを
推奨として文書に書くに留める。ずれたときは、engine が manifest を読むときの検査に任せる。engine は対応版より
新しい `schemaVersion` を拒否し、同じ版の中で知らない列挙値も拒否する（ADR-0059）。

推奨の規範は REQ-14f0aec9-abae-4621-82f3-40536a1ad904 に置く。

## 根拠

版がずれても、版の新しい manifest と知らない列挙値は上の検査で明示的なエラーになる。
同じ入力から揃える運用で回避もできる。

home mode にも devShell や `nix run` 経由の起動を強制すれば版ずれは無くなるが、PATH に常駐する CLI の
使い勝手を捨てることになる。

## 棄却した代替案

### home mode でも devShell か `nix run` 経由の起動を強制する

版ずれは根絶できるが、PATH に常駐する CLI の使い勝手を損なう。
