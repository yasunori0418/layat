---
id: "SDR-fb902413-ff78-45b2-9ddd-3fc25f493208"
type: sara_adr
name: "棄却した提案の要件と設計は提案の形の題で deprecated にして残す"
description: "棄却した提案を記録する要件と設計は、提案の形の題と本文に棄却の記述を添え、status を deprecated にして残す。"
status: stable
sources:
  - id: "pr-716-review"
    resource: "https://github.com/yasunori0418/layat/pull/716#discussion_r4514298788"
    title: "PR #716 のレビューコメント"
origin: "PR #716 のレビューコメント（2026-11-10）で確定"
justifies:
  - "QA-67715bb3-1162-4ccf-8441-2f44257a57da"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-10T00:43:33+09:00"
---
# 棄却した提案の要件と設計は提案の形の題で deprecated にして残す

- ステータス: stable
- 日付: 2026-11-10
- 関連: `docs/model.md`、[一部の entry だけを適用する --only を提供する](../adr/partial-apply-only-flag.md)
- 起点: PR #716 のレビューコメント（2026-11-10）で確定

## 背景

`status` の `deprecated` を提案の棄却に使うのは ADR だけだった。要件と設計には、機能を提供しないという否定を決定として書いた item があり、棄却 ADR と書き方が一致していなかった。

## 決定

要件と設計で棄却した提案は、提案の形の題で `status` を `deprecated` にして残す。本文の先頭に棄却した旨を書き、`specification` を持つ型では `specification` も提案の形で書く。

肯定の規範が中心で否定が付随するだけの item は対象にしない。題に否定が入っているなら、決定の形に改めるだけにする。

## 根拠

棄却 ADR と同じ読み方になる。`name` が提案を表し、`status` が棄却を表す。

## 棄却した代替案

### 否定の題のまま `stable` で残す

「提供しない」という否定が、有効な要件と同じ扱いで並ぶ。棄却 ADR とも書き方が一致しない。

### `specification` は `SHALL NOT` のまま残す

提案の形の題と本文が逆向きになる。
