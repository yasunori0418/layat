---
id: "QA-bf975ac7-1e84-4f75-9fe7-8d6154d69ba1"
type: quality
name: "テスト系 item は、テスト資産・検証テーマ・脅威の 3 段の粒度で起こす"
description: "テスト系 item は CASE をテスト資産・TC を検証テーマ・RISK を脅威の単位で起こし、概要文書へ索引しない"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
specification: |
  A `test_case` SHALL be raised one per test asset (a test file, an e2e scenario or a flake
  check). A `test_condition` SHALL be raised per verification theme, and a `risk` SHALL be
  raised per threat that groups `test_condition`s. A `risk`, `test_condition` or
  `test_case` SHALL NOT be indexed in an overview document.
specification_ja: |
  `test_case` はテスト資産（テストファイル・e2e シナリオ・flake check）1 つにつき 1 件で
  起こさなければならない。`test_condition` は検証テーマごとに、`risk` は `test_condition` を
  束ねる脅威ごとに起こさなければならない。`risk`・`test_condition`・`test_case` を概要文書へ
  索引してはならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-11T00:57:16+09:00"
---
# テスト系 item は、テスト資産・検証テーマ・脅威の 3 段の粒度で起こす

## 仕様

既存のテスト実装から item を起こすときは、次の 3 段の単位で起こす。「対象」は
`docs/test/<対象>/` の区分を指す。区分の一覧と並び順は `dev/tests/test-categories.tsv` が持つ。

- CASE はテストファイル・e2e シナリオ・flake check の単位で起こす。frontmatter の
  `resource`（必須・単一値）に対象を正準表記で書き、人間向けの補足は本文の `## 対象` 節に書く。
  主な検証内容（テスト関数のテーマの列挙）も本文に書く。テスト関数の単位には割らない。
  CASE とテスト資産が 1 対 1 であることは `dev/tests/test-doc-map.sh` が強制する
- TC は検証テーマの単位で起こす（対象あたり数件）。同じ不変条件・観点を検証する CASE 群を
  束ねる
- RISK は TC を束ねる脅威の単位で起こす（対象あたり 2〜5 件）。「このテーマが壊れると何が
  起きるか」を、`requirement` / `design` への `threatens` で表す

関係の向きは次のとおり。

- CASE -`covers`→ TC -`mitigates`→ RISK -`threatens`→ REQ / DSG
- CASE -`follows`→ TP

### risk・TC・CASE は概要文書へ索引しない

risk・TC・CASE は `sara query` / `sara report` で辿る。`docs/spec.md` のリンク集に載せるのは
`requirement`・`quality`・`test_plan` までとする。

## 出典

Issue #273 と Issue #304 で決めたテスト系 item の粒度を規約として運用してきたものを、
Issue #649 で quality item として立てた。
この規約を定めた決定は [テスト系 item を CASE・TC・RISK の 3 段で起こし、概要文書へ索引しない](../sara-adr/test-items-three-tier-granularity.md) に記録してある。
