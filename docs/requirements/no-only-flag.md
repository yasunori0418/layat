---
id: "REQ-4fc98fa6-fd5f-4f25-8b10-60755bf49bd2"
type: requirement
name: "一部の entry だけを適用する --only を提供する"
status: deprecated
derives_from:
  - "UC-1c280dce-7c72-44c0-95ea-d06344f62a47"
specification: |
  A `--only` flag that applies only some of the entries SHALL be provided.
specification_ja: |
  一部の entry だけを適用する `--only` を提供しなければならない。
risk_exempt: "棄却した提案であり、テストで確かめる振る舞いが無い。"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-10T00:44:25+09:00"
---
# 一部の entry だけを適用する --only を提供する

## 仕様

この提案は棄却した。`--only` は提供しない。部分適用は profile 世代の atomic 性と衝突する。
決定と理由は [一部の entry だけを適用する --only を提供する](../adr/partial-apply-only-flag.md) に置く。

## 出典

`docs/spec.md`「CLI 仕様」→「サブコマンド体系」の最終箇条書き。
