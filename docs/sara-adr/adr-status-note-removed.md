---
id: "SDR-fa7eff8e-788f-48e7-951d-4cfbf95d5a1f"
type: sara_adr
name: "adr の status_note を持たない"
section: frontmatter-fields
status: stable
sources:
  - id: "issue-499"
    resource: "https://github.com/yasunori0418/layat/issues/499"
    title: "#499"
origin: "Issue #499 の grilling（2026-11-01）で確定。ADR-0048 §2 を書き直した"
justifies:
  - "QA-030f926b-5ae7-4543-b8cb-e66aae5e2b5e"
references:
  - "SDR-3ff9f355-a667-476e-b10d-087248cdbfcf"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:08:50+09:00"
---
# adr の status_note を持たない

- ステータス: stable
- 日付: 2026-11-01
- 関連: [ADR の失効は supersedes と 廃止 で表す](adr-supersedes-and-deprecated-status.md), GitHub Issue #499
- 起点: Issue #499 の grilling（2026-11-01）で確定。ADR-0048 §2 を書き直した

## 背景

ADR のステータス行には「採用（2026-06-14 改訂: … → ADR-0015）」のようなカッコ書きの改訂注記が
付いていた。enum の値に収まらないので、frontmatter に `status_note` フィールドを設けて受け皿にしていた。

[ADR の失効は supersedes と 廃止 で表す](adr-supersedes-and-deprecated-status.md) と決めたので、ADR は
部分改訂を受けなくなる。

## 決定

ADR の型は `status_note` フィールドを持たない。本文のステータス行にもカッコ書きを付けない。
ステータスは enum の値だけで表す。

## 根拠

カッコ書きの中身は部分改訂の記録である。部分改訂が無くなれば、書く内容が無い。

受け皿を残すと、置き換えずに注記で済ませる書き方を誘う。フィールドごと外して、その書き方をできなくする。

## 棄却した代替案

### `status_note` を任意フィールドとして残す

旧 ADR の注記は本文の改訂注記に同じ内容があり、失われる情報は無い。残す利点が無いので採らない。
