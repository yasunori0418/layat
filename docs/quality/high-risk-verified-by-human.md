---
id: "QA-aa70fff6-3e0a-4163-bf4c-97f1730f7496"
type: quality
name: "level が high の risk は人が確かめた記録を持つ"
description: "level が high の risk は、generated.at 以降に人が内容を確かめた verified を持つ"
specification: |
  A risk whose `level` is high SHALL carry at least one `verified` entry whose `by` begins
  with `human:` and whose `at` is at or after the risk's `generated.at`. An entry by any
  other actor, or one dated before `generated.at`, SHALL NOT count. The requirement SHALL
  apply to the high risk itself only, and a risk whose `status` is `deprecated` SHALL be
  left out. A high risk lacking such an entry SHALL be enumerated mechanically, and the
  result SHALL be reported without gating merging. A `verified` entry SHALL NOT be added
  unless a person has read the risk's content.
specification_ja: |
  `level` が high の risk は、`by` が `human:` で始まり、`at` がその risk の `generated.at`
  以降である `verified` の要素を 1 つ以上持たなければならない。それ以外の actor の要素と、
  `generated.at` より前の要素は数えてはならない。この要求は high の risk 自身にだけ当て
  なければならず、`status` が `deprecated` の risk は除かなければならない。この要素を
  持たない high の risk は機械的に列挙されなければならず、その結果は報告に留め、マージの
  条件にしてはならない。人が risk の内容を読まないまま `verified` の要素を足してはならない。
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-11T23:49:49+09:00"
---
# level が high の risk は人が確かめた記録を持つ

## 仕様

risk の `level` は likelihood と impact から導かれ（→ [risk の level は likelihood と impact から導き、手で決めない](risk-level-derived-from-likelihood-and-impact.md)）、
high は見立てを誤ったときの損失が最も大きい。本 item は high の risk について、エージェントが
書いたまま人が確かめていない状態を機械的に見分けられるようにする。

数えるのは、`by` が `human:` で始まり、`at` が `generated.at` 以降の `verified` だけになる。
`generated.at` より前の確認はその後に変わった本文を確かめていない。対象を high の risk 自身に
限る理由と、medium の risk や test_condition・test_case に広げない理由は
[level が high の risk に人の verified を求める](../sara-adr/human-verified-required-for-high-risk.md) が持つ。

`verified` は自動では付けない。人が内容を確かめたときにだけ手で足す。

未達の risk は列挙して報告に留め、マージの条件にしない
（→ [傾向の計測は報告に留め、マージの条件にしない](measurement-reports-without-gating.md)）。
未達の件数は、人が確かめるのを待っている risk の数で、何かが破損したことを意味しない。

列挙の実装と CI での実行は infrastructure の item が持つ。

## 出典

epic #676 の grilling（2026-11-11）で確定した、鮮度と信頼の確認の仕様。
