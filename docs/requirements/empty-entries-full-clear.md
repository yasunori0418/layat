---
id: "REQ-fc64de4c-c82b-419c-8706-07d8d97daa37"
type: requirement
name: "空の entries は正当な全クリアとして扱い、エラーにも警告にもしない"
description: "空の entries は正当な全クリアとして扱い、エラーにも警告にもせず、前世代の layat symlink を除去して空の新世代を積まなければならない"
section: errors
status: stable
derives_from:
  - "UC-1c280dce-7c72-44c0-95ea-d06344f62a47"
specification: |
  `entries = {}` SHALL be treated as a legitimate full clear, and SHALL NOT be an error or
  a warning. Every layat symlink of the previous generation SHALL be removed by conservative
  stale removal, and the new generation SHALL be empty. What conservative stale removal
  removes is stated by [古くなった配置の除去は前世代の記録通りを指す symlink のみに限り、copy は消さず孤児を警告する](stale-removal-invariants.md) and SHALL NOT be restated here.
specification_ja: |
  `entries = {}` は正当な全クリアとして扱わなければならず、エラーにも警告にもして
  はならない。前世代の全 layat symlink を古くなった配置の保守的な除去で除去しなければならず、
  新世代は空でなければならない。
  古くなった配置の保守的な除去が何を除去するかは [古くなった配置の除去は前世代の記録通りを指す symlink のみに限り、copy は消さず孤児を警告する](stale-removal-invariants.md) の規範であり、ここでは再掲しない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T22:44:47+09:00"
---
# 空の entries は正当な全クリアとして扱い、エラーにも警告にもしない

## 仕様

| 条件 | 動作 |
|---|---|
| `entries = {}`（空 manifest）| 正当な全クリア。前世代の全 layat symlink を保守的 stale 除去し新世代は空（警告なし）|

> **上は原文の写しで、規範は frontmatter が正**。古くなった配置の保守的な除去の不変条件そのもの
> （何を消し何を消さないか）は [古くなった配置の除去は前世代の記録通りを指す symlink のみに限り、copy は消さず孤児を警告する](stale-removal-invariants.md)、copy が古くなった配置の除去で消えず孤児の警告に
> なることも同 item の担当。

## 出典

`docs/spec.md`「エラー仕様」節の表の `entries = {}` 行。

決定の実体は [空の entries を正当な全クリアとして許可する](../adr/empty-entries-allowed.md)で、空 entries をエラーにせず全クリアとして受理することを定めている。
