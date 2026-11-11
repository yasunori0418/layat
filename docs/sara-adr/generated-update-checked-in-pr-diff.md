---
id: "SDR-5ef8ab39-cefe-4c6f-b743-730316d96568"
type: sara_adr
name: "generated の更新漏れは PR の差分で確かめる"
description: "base との差分で本文が変わったのに generated.at が変わっていない item を挙げ、全履歴と時刻は比べない"
status: stable
sources:
  - id: "issue-676"
    resource: "https://github.com/yasunori0418/layat/issues/676"
    title: "epic: 文書の鮮度と信頼を機械的に確かめる — 伝播する鮮度・generated.at と履歴の一致・verified と RISK の重さ"
origin: "epic #676 の grilling（2026-11-11）で確定"
justifies:
  - "QA-611977bd-3d11-4695-9fb2-a9f7211fcfbd"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-11T23:49:55+09:00"
---
# generated の更新漏れは PR の差分で確かめる

- ステータス: stable
- 日付: 2026-11-11
- 関連: GitHub Issue #676, `docs/model.md`
- 起点: epic #676 の grilling（2026-11-11）で確定

## 背景

[item に生成者・確認の記録・期限を持たせる](generated-verified-stale-after.md) は、本文を変えたら
`generated` を更新すると定め、それが守られているかは確かめず規約で守るとした。`generated.at` が
古いまま残ると、上流の relation を辿る鮮度の比較が、本文の変更を見落とす。

## 決定

base との差分で本文が変わったのに `generated.at` が変わっていない item を、更新漏れとして挙げる。
時刻は比べない。

- 比べる 2 点は `git merge-base <base> HEAD` のコミットと作業ツリーとする
- 本文は frontmatter の外の行とする。`# ` で始まる H1 の行、`- ステータス:` で始まる行、
  空白だけの差を除いて差があれば、本文が変わったとみなす
- 新しく足した item と削除した item は対象にしない。改名した item は旧パスの内容と比べる
- 挙げた結果は報告に留める

## 根拠

更新漏れは本文を変えるその変更の中で起きる。PR の差分を見れば、変更した item ごとに本文と
`generated.at` が一緒に変わったかを判定でき、時刻の精度に左右されない。

H1 と `- ステータス:` の行は frontmatter の `name` と `status` を写した行で、写しただけの変更では
`generated` を更新しない。空白だけの差は意味を変えない。新しく足した item は起票時に `generated` が
付き、削除した item は残らないので、比べる相手が無い。改名は `generated` を更新しない変更なので、
旧パスの内容と比べて本文の差だけを見る。

## 棄却した代替案

### 全履歴で、本文を変えたコミットの時刻と `generated.at` を比べる

規約どおりに書いた item でも、`generated.at` はコミットの author date より数分古くなる。本文を
書いた時刻に `generated` を付け、その後にコミットするためである。機械的な一括書き換えのコミットも
履歴に並ぶ。着手時点（main `2b0e66a6`）の 790 件に当てると 659 件が違反になり、更新漏れを
見分けられない。

### 許容幅を付けて時刻を比べる

書いてからコミットするまでの時間には上限が無く、幅をいくつにしても規約どおりの item が
外れるか、更新漏れを見逃す。一括書き換えのコミットは幅では除けない。
