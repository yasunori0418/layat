---
id: "QA-611977bd-3d11-4695-9fb2-a9f7211fcfbd"
type: quality
name: "上流より古い item と generated の更新漏れは機械的に列挙する"
description: "上流より基準時刻が古い下流の item と、本文を変えたのに generated を更新していない item を機械的に列挙し、結果は報告に留める"
specification: |
  The reference time of an item SHALL be the latest of its `generated.at` and every
  `verified[].at` that is at or after its `generated.at`, whatever the actor. An item that
  declares a `refines`, `derives_from`, `satisfies`, `threatens`, `mitigates`, `covers` or
  `follows` relation SHALL be enumerated as awaiting re-verification when the
  `generated.at` of the item it points to is later than its own reference time;
  `justifies` and the peer relations SHALL NOT take part, and an item whose `status` is
  `deprecated` SHALL be left out on either end. Comparing the merge base of a base
  revision and `HEAD` with the working tree, an item whose body changed while its
  `generated.at` did not SHALL be enumerated as missing a `generated` update. The body
  SHALL be the lines outside the frontmatter, with the H1 line starting with `# `, lines
  starting with `- ステータス:` and whitespace-only differences disregarded. An item added
  or deleted between the two sides SHALL NOT be compared, and a renamed item SHALL be
  compared with the content at its old path. Both enumerations SHALL be mechanical, and
  their results SHALL be reported without gating merging.
specification_ja: |
  item の基準時刻は、その item の `generated.at` と、`generated.at` 以降の `verified[].at`
  （actor を問わない）のうち最も新しい時刻でなければならない。`refines`・`derives_from`・
  `satisfies`・`threatens`・`mitigates`・`covers`・`follows` の relation を宣言している item は、
  指している item の `generated.at` が自身の基準時刻より新しいとき、再検証待ちとして列挙され
  なければならない。`justifies` と peer の relation は対象にしてはならず、`status` が
  `deprecated` の item は両端のどちらでも除かなければならない。`git merge-base <base> HEAD` の
  コミットと作業ツリーを比べ、本文が変わったのに `generated.at` が変わっていない item は、
  `generated` の更新漏れとして列挙されなければならない。本文は frontmatter の外の行とし、`# ` で
  始まる H1 の行、`- ステータス:` で始まる行、空白だけの差を除かなければならない。2 点の間で
  足した item と削除した item は比べてはならず、改名した item は旧パスの内容と比べなければ
  ならない。2 つの列挙は機械的に行わなければならず、その結果は報告に留め、マージの条件に
  してはならない。
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-12T00:02:04+09:00"
---
# 上流より古い item と generated の更新漏れは機械的に列挙する

## 仕様

sara の構造の検査は、型・relation・接続性が正しいかを見るが、上流の本文が変わったあとに下流が
古い前提のまま残っているかは見ない。本 item は `generated` と `verified` を使って、読み直しの
要る item を 2 つの観点で列挙する。

再検証待ちは上流の relation を辿って鮮度を比べる。下流の item の基準時刻より上流の
`generated.at` が新しければ挙げる。辿る relation・基準時刻・除く item の理由は
[鮮度は上流の relation を辿って下流へ伝播させる](../sara-adr/freshness-propagates-along-upstream-relations.md) が持つ。

`generated` の更新漏れは、base との差分で本文が変わったのに `generated.at` が変わっていない
item を挙げる。比べる 2 点は `git merge-base <base> HEAD` のコミットと作業ツリーで、本文は
frontmatter の外の行から H1 の行と `- ステータス:` の行を除いたものとし、空白だけの差は数えない。新しく足した
item と削除した item は比べず、改名した item は旧パスの内容と比べる。この範囲にした理由と、時刻を
比べない理由は [generated の更新漏れは PR の差分で確かめる](../sara-adr/generated-update-checked-in-pr-diff.md) が持つ。

どちらの列挙も、挙がったことは何かが破損したことを意味しない。上流を書き換えれば、下流の読み直しが
済むまで件数が増える。そのため結果は報告に留め、マージの条件にしない
（→ [傾向の計測は報告に留め、マージの条件にしない](measurement-reports-without-gating.md)）。
挙がった item をどう扱うかは人が判断する。件数を減らすために `verified` を足したり `generated.at` を
動かしたりしない。

列挙の実装と CI での実行は infrastructure の item が持つ。

## 出典

epic #676 の grilling（2026-11-11）で確定した、鮮度と信頼の確認の仕様。
