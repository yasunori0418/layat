---
id: "SDR-f27192ea-3ca8-4794-8a6a-805dd8ab323d"
type: sara_adr
name: "level が high の risk に人の verified を求める"
description: "level が high の risk 自身に、generated.at 以降の human: の verified を求め、medium・low の risk と TC・CASE には求めない"
section: verification
status: stable
sources:
  - id: "issue-676"
    resource: "https://github.com/yasunori0418/layat/issues/676"
    title: "epic: 文書の鮮度と信頼を機械的に確かめる — 伝播する鮮度・generated.at と履歴の一致・verified と RISK の重さ"
origin: "epic #676 の grilling（2026-11-11）で確定"
justifies:
  - "QA-aa70fff6-3e0a-4163-bf4c-97f1730f7496"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-11T23:49:55+09:00"
---
# level が high の risk に人の verified を求める

- ステータス: stable
- 日付: 2026-11-11
- 関連: GitHub Issue #676, `docs/model.md`
- 起点: epic #676 の grilling（2026-11-11）で確定

## 背景

item の多くはエージェントが書き、`generated.by` はエージェントの actor になる。`verified` は人が
内容を確かめたときにだけ手で足す記録で、自動では付かない。どの item を人が確かめるべきかは
決まっていない。

risk の `level` は likelihood と impact から導かれ、high は起きやすさと影響の組み合わせが最も重い
ものを表す。着手時点（main `2b0e66a6`）の risk は high が 24 件・medium が 22 件・low が 5 件で、
`verified` を持つ item は 1 件も無い。

## 決定

`level` が high の risk は `generated.at` 以降の `human:` の `verified` を持たなければならないとし、
持たないものを挙げる。

- `human:` の `verified` は、`by` が `human:` で始まり、`at` がその risk の `generated.at` 以降の要素とする
- 対象は `level` が high の risk 自身に限る。`status` が `deprecated` の risk は対象から外す
- 挙げた結果は報告に留める。着手時点で high の risk 24 件前後が初日から挙がるが、件数を減らすために
  item を書き換えない

## 根拠

high の risk は、見積もりを誤ったときの損失が最も大きい。人が脅威の見立てを確かめる対象を、
level で機械的に決められる。

`generated.at` より前の `verified` は、その後に本文が変わっているので、いまの見立てを確かめた
記録にならない。エージェントの actor の `verified` は書き手と同じ種類の読み手による確認で、
人の確認の代わりにならない。

## 棄却した代替案

### medium の risk まで広げる

着手時点で 46 件前後が挙がり、high の risk の確認が埋もれる。medium の見立ての誤りは損失が
小さく、人の確認の手間に見合わない。

### risk の下流の test_condition と test_case にも求める

test_condition と test_case は risk に対処するテーマと実体で、数が多い。risk の見立てが確かめて
あれば、その下流が古い前提のまま残っていないかは、上流の relation を辿る鮮度の比較が挙げる。
