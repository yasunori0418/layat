---
id: "SDR-7631120c-0006-4394-a9cc-32d2f386a67d"
type: sara_adr
name: "テスト系 item を CASE・TC・RISK の 3 段で起こす"
description: "CASE はテスト資産、TC は検証テーマ、RISK は TC を束ねる脅威の単位で起こす"
section: risk-test-defect
status: stable
sources:
  - id: "issue-719"
    resource: "https://github.com/yasunori0418/layat/issues/719"
    title: "#719"
origin: "epic #719 の実行中の裁定（2026-11-13）で確定"
justifies:
  - "QA-bf975ac7-1e84-4f75-9fe7-8d6154d69ba1"
supersedes:
  - "SDR-e444ba7f-eac7-4cac-bff8-c3d7910d836a"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-13T16:29:14+09:00"
---
# テスト系 item を CASE・TC・RISK の 3 段で起こす

- ステータス: stable
- 日付: 2026-11-13
- 関連: GitHub Issue #719, `dev/tests/test-doc-map.sh`
- 起点: epic #719 の実行中の裁定（2026-11-13）で確定

## 背景

[テスト系 item を CASE・TC・RISK の 3 段で起こし、概要文書へ索引しない](test-items-three-tier-granularity.md)
は、既存のテスト実装から risk・test_condition・test_case を逆算して起こすときの item の単位を決め、
あわせて risk・TC・CASE を概要文書へ索引しないと決めていた。

その後 [item の一覧は生成した index.md だけが持つ](item-lists-held-by-generated-index.md) で、
手書きの文書はどの型の item の一覧も持たなくなった。risk・TC・CASE だけを概要文書の索引から外す
規則は、一覧の扱いを決める SDR と重なる（#719）。

## 決定

- CASE はテストファイル・e2e シナリオ・flake check の単位で起こし、テスト関数の単位には割らない。
  CASE とテスト資産を 1 対 1 にする
- TC は検証テーマの単位で起こし、同じ不変条件・観点を検証する CASE 群を束ねる
- RISK は TC を束ねる脅威の単位で起こし（対象あたり 2〜5 件）、`requirement` / `design` への
  `threatens` で表す

## 根拠

テスト資産の単位にする根拠は、CASE と資産が 1 対 1 になり、資産の追加やリネームに CASE が追従して
いるかを契約テストで確かめられることにある。

一覧の扱いは型を問わず「item の一覧は生成した index.md だけが持つ」が決めるので、テスト系 item に
限った索引の規則は持たない。

## 棄却した代替案

### テスト関数の単位で CASE を起こす

CASE が 200 件を超え、ファイル単位のテスト資産とも 1 対 1 にならない。
