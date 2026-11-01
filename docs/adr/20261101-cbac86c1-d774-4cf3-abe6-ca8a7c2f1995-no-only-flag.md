---
id: "ADR-cbac86c1-d774-4cf3-abe6-ca8a7c2f1995"
type: adr
name: "--only は持たない"
status: 採用
origin: "ADR-0002 §決定（--only を廃止する）を書き直した"
justifies:
  - "REQ-4fc98fa6-fd5f-4f25-8b10-60755bf49bd2"
---
# ADR-cbac86c1-d774-4cf3-abe6-ca8a7c2f1995: --only は持たない

- ステータス: 採用
- 日付: 2026-11-01
- 起点: ADR-0002 §決定（--only を廃止する）を書き直した

## 背景

config の一部の entry だけを適用したいという需要がある。CLI に対象 entry を絞る `--only` を
持たせれば応えられる。

## 決定

一部の entry だけを適用する `--only` は持たない。選択的に更新したいときは config（`layat.<name>`）を
分ける。規範は REQ-4fc98fa6-fd5f-4f25-8b10-60755bf49bd2 に置く。

## 根拠

profile の世代は manifest 全体を 1 単位として atomic に切り替わる。一部の entry だけを適用すると、
世代が記録する内容と FS の状態が食い違い、次の stale 除去や rollback の基準が崩れる。config を
分ければ、独立して更新したい単位がそのまま独立した世代になる。

## 棄却した代替案

### `--only` を持ち、部分適用の結果も新しい世代として積む

世代の中身が config と一致しなくなる。世代は「その config を完全に適用した状態」という前提が崩れる。
