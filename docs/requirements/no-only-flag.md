---
id: "REQ-4fc98fa6-fd5f-4f25-8b10-60755bf49bd2"
type: requirement
name: "一部 entry だけを適用する --only は提供しない"
status: stable
derives_from:
  - "UC-1c280dce-7c72-44c0-95ea-d06344f62a47"
specification: |
  A `--only` flag that applies only some of the entries SHALL NOT be provided, because it
  conflicts with the atomicity of a profile generation. Selective updating SHALL instead
  be achieved by splitting the config (`layat.<name>`).
specification_ja: |
  一部 entry だけを適用する `--only` は提供してはならない。profile 世代の atomic 性と
  衝突するため。選択的更新は config（`layat.<name>`）の分割で担保しなければならない。
risk_exempt: "`--only` を提供しないという不在の要件であり、テストで確かめる振る舞いが無い。"
generated:
  by: "claude-code/claude-opus-5"
  at: "2026-09-20T08:21:50+09:00"
---
# 一部 entry だけを適用する --only は提供しない

## 仕様

`--only`（一部 entry だけ適用）は profile 世代の atomic 性と衝突するため提供しない。
選択的更新は config（`layat.<name>`）の分割で担保する。

## 出典

`docs/spec.md`「CLI 仕様」→「サブコマンド体系」の最終箇条書き。
