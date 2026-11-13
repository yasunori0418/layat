---
id: "QA-67715bb3-1162-4ccf-8441-2f44257a57da"
type: quality
name: "規範は 1 ファイル 1 item のグラフが持ち、item の一覧は生成した index.md だけが持つ"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
specification: |
  Normative content SHALL be held by the item graph alone, one item per file. The list of
  items of each type SHALL be held only by the index.md generated from the graph, and a
  hand-written document SHALL NOT keep such a list or restate the norms of the items.
  Each item SHALL be placed under the type its norm binds, following the project's
  discrimination rule, and its identifier SHALL be drawn from the project's numbering
  command rather than assigned by hand, so that parallel lanes cannot collide on a number.
  An item whose type carries a specification SHALL state that norm in English and
  SHALL carry the corresponding Japanese norm alongside it; neither field SHALL be treated
  as a gloss of the other.
specification_ja: |
  規範的な内容は 1 ファイル 1 item の item グラフだけが持たなければならない。型ごとの item の
  一覧は、グラフから生成した index.md だけが持たなければならず、手書きの文書はその一覧を持っては
  ならず、item の規範を書き写してもならない。
  各 item は、その規範が何を縛るかに応じてプロジェクトの型判別規約に従った型の下へ配置され
  なければならず、その識別子は手で振るのではなくプロジェクトの採番コマンドから採らなければ
  ならない（並列レーンで採番が衝突しないようにするため）。specification を持つ型の item は、
  その規範を英語で述べなければならず、対応する日本語の規範文を併記しなければならない。
  いずれの欄も他方の訳注として扱ってはならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-13T14:34:03+09:00"
---
# 規範は 1 ファイル 1 item のグラフが持ち、item の一覧は生成した index.md だけが持つ

## 仕様

規範を手書きの文書と item の両方に書くと二重管理になり、改訂で片方だけが古くなる。規範の所在を
グラフ側に一本化し、型ごとの item の一覧はグラフから生成した index.md が持つ。各ディレクトリの
手書きの README.md は、item から導けない全体像の散文だけを持つ。この分担は文書の分量ではなく
**どちらを正とするか**の宣言であり、README.md へ仕様や item の一覧を書き足すことは規範の複製にあたる。

**配置と採番を規約で固定する**のが要点になる。型を取り違えた item は形が正しい限り機械検証を
検出されずに通るため（[ドキュメントグラフのトレーサビリティは機械的に検証する](document-graph-mechanically-verified.md) の検証は参照の解決・一意性・非循環しか見ない）、正しい型へ置く
ことはレビューで守られる判断になる。採番を手で行わないのは、並列で作業するレーンが同じ番号を
採ることを構造的に避けるため。

二言語で書く規範は、英語と日本語のどちらも正として扱うことを含む。片方を訳注に降格させると、
実際に読まれる側の言語で規範の強度が落ちる。強度の対応（SHALL と「〜しなければならない」など）
まで含めた執筆規約は [specification は SHALL 系で書き、日本語の規範文は同じ強度の規範助動詞で終える](specification-shall-family-and-matching-auxiliaries.md) が持ち、これも機械検証されないためレビューで
守る。

型の判別基準の具体・採番コマンドの実装・強度の対応表は本 item の規範に含めない。判別は
[規範の型は、それが誰を縛るかで決める](norm-type-by-who-it-binds.md) が、執筆の規約は [specification は SHALL 系で書き、日本語の規範文は同じ強度の規範助動詞で終える](specification-shall-family-and-matching-auxiliaries.md) が、採番の仕組みは [トレーサビリティ検証基盤（sara によるドキュメントグラフの機械検証）](../infrastructure/traceability-verification.md) が持つ。

## 出典

`docs/spec.md` 冒頭の書き方の規約、規約文書の型判別規約と
`specification` / `specification_ja` の執筆規約、`CLAUDE.md` の ID 規約節、
`docs/dev/definition-of-done.md` の DOD-04 が実運用してきた規範を、Issue #272 で quality item
として立てたもの。
