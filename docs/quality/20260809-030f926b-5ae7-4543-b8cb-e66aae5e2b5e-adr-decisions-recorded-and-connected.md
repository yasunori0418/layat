---
id: "QA-030f926b-5ae7-4543-b8cb-e66aae5e2b5e"
type: quality
name: "設計判断は ADR に記録し、改訂の書き戻しと item への接続を同じ変更の中で完了させる"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
depends_on:
  - "QA-6bf957d9-17d9-4660-92b7-ebd6eeb71a8c"
specification: |
  A change that carries a design decision SHALL record that decision as an ADR. Where a
  change alters any decision of an existing ADR, that ADR SHALL be rewritten as a whole into
  new ADRs rather than amended in part, and once every decision of the existing ADR has a
  successor, the successors SHALL supersede it and its status SHALL become deprecated within the
  same change. Every ADR SHALL connect to at least one item it justifies, and where no such item exists
  yet, it SHALL be created within the same change rather than the connection being left
  empty, so that an empty connection remains readable as an omission. Whether every decision
  of a superseded ADR has a successor is not mechanically verified, so it SHALL be upheld by
  review.
specification_ja: |
  設計判断を伴う変更は、その判断を ADR として記録しなければならない。既存 ADR の決定を 1 つでも
  変える場合は、その ADR を部分的に書き換えず丸ごと新しい ADR へ書き直さなければならず、既存 ADR の
  全決定に後継が揃った時点で、同じ変更の中で後継から supersedes を張り、既存 ADR のステータスを
  deprecated にしなければならない。すべての ADR は、自身が裏づける item へ
  1 本以上接続しなければならず、接続先の item がまだ無い場合は、接続を空のまま残すのではなく
  同じ変更の中で item を起こさなければならない（空の接続が埋め忘れとして読み取れる状態を
  保つため）。置き換えで deprecated にする ADR の全決定に後継が揃っているかは機械的に検証されないため、レビューで
  守られなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-08T00:53:48+09:00"
---
# QA-030f926b-5ae7-4543-b8cb-e66aae5e2b5e: 設計判断は ADR に記録し、改訂の書き戻しと item への接続を同じ変更の中で完了させる

## 仕様

決定を記録しない、あるいは記録したが後続の改訂で一部だけが古くなった ADR を有効なまま残す状態は、
その ADR だけを読んだ人に「上書きされた内容」を最新仕様と誤読させる。本 item はその防止を規範として
固定する。

**「同じ変更の中で」**を条件に含めるのが要点になる。`deprecated` と `supersedes` も item への接続も、
後回しにしてよい形にすると守られなくなる。接続先の item が無いことは接続を空にする理由にならず、item を先に
起こすか、決定が定まっていないものとして ADR の粒度を解き直すかのどちらかになる。

接続を空のまま残さない規範は、**空を埋め忘れとして機械判定できる状態**を保つためにある。
接続の非空性は QA-6bf957d9-17d9-4660-92b7-ebd6eeb71a8c の機械検証が見つけ、`deprecated` と `supersedes` の対応は
`dev/tests/adr-supersede.sh` が見つける。どちらも frontmatter しか見ないので、置き換えで deprecated にする ADR の全決定に
後継が揃っているかは視野の外にあり、そこは人手のレビューで守る（判定の条件は
`docs/agents/adr-rules.md` の「deprecated にできる条件」節が持つ）。

ADR の粒度・置き換えの手順・接続先の型の一覧は本 item の規範に含めない。規則は
`docs/agents/adr-rules.md` が、layat への当てはめは `docs/adr/README.md` が、接続可能な型は
`docs/model.yaml` が持つ。

## 出典

`docs/agents/adr-rules.md`（置き換えの規則・`justifies` 必須の規約）と `docs/adr/README.md`
（セルフチェック手順）、`docs/dev/definition-of-done.md` の DOD-03、
`.github/PULL_REQUEST_TEMPLATE.md` の「docs / ADR 影響」欄が実運用してきた規範を、Issue #272 で
quality item として立てたもの。
