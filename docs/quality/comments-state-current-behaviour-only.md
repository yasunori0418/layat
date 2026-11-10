---
id: "QA-1b4601e5-33cb-435f-812e-052cc159ce8b"
type: quality
name: "コメントと利用者向け文字列は現在の動作だけを書き、決定の記録を参照しない"
description: "コメントは 3 行以内で現在の動作だけを書き、経緯と決定の記録への参照を書かない。利用者向け文字列に ADR 番号を出さない"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
specification: |
  The prose of consecutive comment lines SHALL fit within three lines, counting bullet
  items as prose. A comment SHALL state only the current behaviour and SHALL NOT state
  history or reference an ADR, Issue or PR. A user-facing string (help text, an error
  message and the like) SHALL NOT expose an ADR number. A comment SHALL NOT keep a code
  block other than a decision table, a user-facing sample or a usage example.
specification_ja: |
  連続するコメント行の散文は、箇条書きも散文に数えて 3 行以内に収めなければならない。
  コメントには現在の動作だけを書かなければならず、経緯と ADR・Issue・PR への参照を書いては
  ならない。利用者向け文字列（ヘルプ・エラーメッセージなど）に ADR 番号を出してはならない。
  コメントには、判定表・利用者向けのサンプル・実行例のほかにコードブロックを残してはならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-10T23:44:58+09:00"
---
# コメントと利用者向け文字列は現在の動作だけを書き、決定の記録を参照しない

## 仕様

コードのコメントと利用者向け文字列は、次の 4 項目に従う。

- 連続するコメント行の散文は 3 行以内にする。箇条書きも散文に数える
- 現在の動作だけを書く。経緯と ADR・Issue・PR への参照を書かない
- 利用者向け文字列（ヘルプ・エラーメッセージなど）に ADR 番号を出さない
- コードブロックを残してよいのは判定表・利用者向けのサンプル・実行例だけとする

## 出典

CLAUDE.md の規約として運用してきたコメント規約を、Issue #649 で quality item として立てた。
