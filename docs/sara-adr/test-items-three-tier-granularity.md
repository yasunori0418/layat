---
id: "SDR-e444ba7f-eac7-4cac-bff8-c3d7910d836a"
type: sara_adr
name: "テスト系 item を CASE・TC・RISK の 3 段で起こし、概要文書へ索引しない"
description: "CASE はテスト資産、TC は検証テーマ、RISK は TC を束ねる脅威の単位で起こし、概要文書へ索引しない。"
section: risk-test-defect
status: deprecated
sources:
  - id: "issue-273"
    resource: "https://github.com/yasunori0418/layat/issues/273"
    title: "#273"
  - id: "issue-304"
    resource: "https://github.com/yasunori0418/layat/issues/304"
    title: "#304"
  - id: "issue-649"
    resource: "https://github.com/yasunori0418/layat/issues/649"
    title: "#649"
origin: "Issue #273 の grilling（2026-08-09）で確定"
justifies:
  - "QA-bf975ac7-1e84-4f75-9fe7-8d6154d69ba1"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-11T00:07:44+09:00"
---
# テスト系 item を CASE・TC・RISK の 3 段で起こし、概要文書へ索引しない

> **廃止（2026-11-13）**: 本 SDR の決定は次の後継へ移った。以下の本文は当時の記録として読む。
>
> - 全体 → [テスト系 item を CASE・TC・RISK の 3 段で起こす](test-items-in-three-tiers.md)

- ステータス: deprecated
- 日付: 2026-11-11
- 関連: GitHub Issue #273, #304, #649
- 起点: Issue #273 の grilling（2026-08-09）で確定

## 背景

#273 では既存のテスト実装から risk・test_condition・test_case を逆算して起こすために、item の
単位を決める必要があった。#304 では CASE とテスト資産の対応を契約テストで機械的に照合する
仕組みを作った。決定は Issue の本文と CLAUDE.md にしか無く、記録が無かった。

## 決定

- CASE はテストファイル・e2e シナリオ・flake check の単位で起こし、テスト関数の単位には割らない。
  CASE とテスト資産を 1 対 1 にする
- TC は検証テーマの単位で起こし、同じ不変条件・観点を検証する CASE 群を束ねる
- RISK は TC を束ねる脅威の単位で起こし（対象あたり 2〜5 件）、`requirement` / `design` への
  `threatens` で表す
- risk・TC・CASE は概要文書へ索引せず、`sara query` / `sara report` で辿る

## 根拠

テスト資産の単位にすれば、CASE と資産が 1 対 1 になり、資産の追加やリネームに CASE が追従して
いるかを契約テストで確かめられる。risk 以下は件数が多く、関係は sara で辿れるので、概要文書の
索引は `requirement`・`quality`・`test_plan` までにとどめる。

## 棄却した代替案

### テスト関数の単位で CASE を起こす

CASE が 200 件を超え、ファイル単位のテスト資産とも 1 対 1 にならない。
