---
id: "ADR-cbac86c1-d774-4cf3-abe6-ca8a7c2f1995"
type: adr
name: "一部の entry だけを適用する --only を提供する"
status: deprecated
origin: "ADR-0002 §決定で棄却された提案（実装前の spec にあった --only）を書き直した"
justifies:
  - "REQ-4fc98fa6-fd5f-4f25-8b10-60755bf49bd2"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# 一部の entry だけを適用する --only を提供する

- ステータス: deprecated
- 日付: 2026-11-01
- 起点: ADR-0002 §決定で棄却された提案（実装前の spec にあった --only）を書き直した

## 背景

初期の spec では、config の一部の entry だけを適用する `--only` を CLI に持たせていた。

世代管理を nix profile に乗せると決めた際（[世代は link farm と nix profile に乗せる](generations-on-link-farm-and-nix-profile.md)）、この提案は実装に入る前に棄却された。
profile の世代は manifest 全体を単位に atomic に切り替わるので、一部の entry だけを適用すると
世代の記録と実際の配置がずれるためである。CLI に `--only` が実装されたことは無い。

## 決定

`--only` は提供しない。選択的に更新したいときは config（`layat.<name>`）を分ける。
規範は [一部 entry だけを適用する --only は提供しない](../requirements/no-only-flag.md) に置く。

## 根拠

一部の entry だけを適用すると、世代が記録する内容と FS の状態が食い違い、次の古くなった
配置の除去や rollback の基準が崩れる。config を分ければ、独立して更新したい単位がそのまま独立した世代になるので、
`--only` を提供しなくても選択的な更新はできる。

## 棄却した代替案

### 提案どおり `--only` を提供し、部分適用の結果も新しい世代として積む

世代の中身が config と一致しなくなり、世代は「その config を完全に適用した状態」という前提が崩れる。
