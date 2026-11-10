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
  at: "2026-11-11T00:28:41+09:00"
---
# 設計判断は ADR に記録し、改訂の書き戻しと item への接続を同じ変更の中で完了させる

## 仕様

決定を記録しない、あるいは記録したが後続の改訂で一部だけが古くなった ADR を有効なまま残す状態は、
その ADR だけを読んだ人に「上書きされた内容」を最新仕様と誤読させる。本 item はその防止を規範として
固定する。

**「同じ変更の中で」**を条件に含めるのが要点になる。`deprecated` と `supersedes` も item への接続も、
後回しにしてよい形にすると守られなくなる。接続先の item が無いことは接続を空にする理由にならず、item を先に
起こすか、決定が定まっていないものとして ADR の粒度を解き直すかのどちらかになる。

接続を空のまま残さない規範は、**空を埋め忘れとして機械判定できる状態**を保つためにある。
接続の非空性は [ドキュメントグラフのトレーサビリティは機械的に検証する](document-graph-mechanically-verified.md) の機械検証が見つけ、`deprecated` と `supersedes` の対応は
`dev/tests/adr-supersede.sh` が見つける。どちらも frontmatter しか見ないので、置き換えで deprecated にする ADR の全決定に
後継が揃っているかは視野の外にあり、そこは人手のレビューで守る（判定の条件は下の
「deprecated にできる条件」節が持つ）。

ADR の粒度・本文の構成・統合と分割の手順・`justifies` の選定基準と、接続先の型の一覧は本 item の
規範に含めない。前者は [ADR は 1 本に 1 つの決定を 4 節で書き、justifies は規範の形成に寄与した item へ張る](adr-one-decision-four-sections-justifies-by-contribution.md) が、
layat への当てはめは `docs/adr/README.md` が、接続可能な型は `docs/model.yaml` が持つ。
ステータス・出典・置き換え・採番の手引きを次に示す。

### ステータス

| ステータス | 意味 |
|---|---|
| `draft` | 検討中で、まだ決定していない |
| `stable` | 決定として有効 |
| `deprecated` | 後継の ADR に置き換えられた ADR と、提案を棄却した ADR |

`deprecated` の ADR には 2 種類ある。`supersedes` で指されていれば後継に置き換わった ADR で、
指されていなければ提案を棄却した記録になる。

### 出典

起点になった Issue など、item にできない外部の出典は frontmatter の `sources` に書く。
要素は `id` / `resource` / `title` の組にする。

- `resource` には出典の URL を書く
- `id` は 1 本の ADR の中で一意にする
- `title` には、URL を開かなくても出典の見当がつく文字列を書く

### 改訂は丸ごとの置き換えで表す

既存の ADR の決定を 1 つでも変えるときは、その ADR を部分的に書き換えず、**丸ごと新しい ADR として
書き直し、古い ADR を `deprecated` にする**。節単位の部分改訂を表す関係や、古い ADR への改訂注記は使わない。
部分改訂を積み重ねると、古い ADR のどこが生きているかを読者が注記から再構成しなければならなくなるため。

- `supersedes` は文書丸ごとの失効だけを表す。後継の ADR から置き換えた ADR へ張る
- `supersedes` で指された ADR は `deprecated` にする。この対応は frontmatter だけで機械的に確かめる
- 改訂を伴わない参照は `references` で表す。`supersedes` とは併用しない

#### deprecated にできる条件

古い ADR を置き換えて `deprecated` にできるのは、**古い ADR の全決定に後継が揃ったとき**だけ。後継は、
新しく起こした ADR か、別の有効な ADR に既にある決定のどちらでもよい。

1 本の古い ADR を複数の新しい ADR へ分けて書き直す途中では、新しい ADR の `origin` に元の ADR と節を
書くにとどめ、`supersedes` はまだ張らない。最後の決定に後継が揃い、古い ADR を `deprecated` にする変更の
中で、全ての後継から `supersedes` をまとめて張る。

#### deprecated にした ADR に残すもの

置き換えで `deprecated` にした ADR は、本文も既存の注記も原文のまま残す。変えるのは次の 3 つだけ。

- 本文のステータス行
- frontmatter の `status`
- 冒頭に置く後継の一覧（どの決定がどの ADR に移ったかを示すバナー）

`justifies` も外さない。要求や設計の item の出典が古い ADR を指していても、その参照は経緯として
そのまま有効になる。

### 採番

ADR も他の item と同じく UUID で採番する。連番は並列の作業で同じ番号を取り合い、マージ後に
採り直しが要る。UUID ならこの衝突が構造的に起きない。

## 出典

ADR の規則として運用してきた置き換えの規則・`justifies` 必須の規約と `docs/adr/README.md`
（セルフチェック手順）、`docs/dev/definition-of-done.md` の DOD-03、
`.github/PULL_REQUEST_TEMPLATE.md` の「docs / ADR 影響」欄が実運用してきた規範を、Issue #272 で
quality item として立てたもの。
