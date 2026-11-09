---
id: "QA-4a623664-650d-4a08-800f-691f4ea6ff91"
type: quality
name: "用語の正名は用語 item が固定し、執筆はそれに従う"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
specification: |
  The canonical spelling of each domain term SHALL be fixed by a single `ubiquitous_term`
  item, and the index (`docs/glossary.md`) and every other document that states the same
  term SHALL derive from that item rather than fix a spelling of its own. Writing across
  the repository SHALL use the canonical term and SHALL avoid the spellings the item lists
  as ones to avoid. Where a term needed while writing has no item, the gap SHALL be
  resolved by adding an item or by choosing an existing one, rather than by coining a
  spelling in place.
specification_ja: |
  各ドメイン用語の正名は、単一の `ubiquitous_term` の item が固定しなければならず、索引
  （`docs/glossary.md`）と同じ用語を述べる他の文書は、独自に綴りを固定するのではなくその
  item から導出されなければならない。リポジトリ全体の執筆はその正名を用いなければならず、
  item が避けるものとして挙げる綴りを避けなければならない。執筆中に必要になった用語に item が
  無い場合、その欠落は、その場で綴りを作り出すのではなく、item を追加するか既存の用語を
  選ぶことで解消されなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T01:11:46+09:00"
---
# 用語の正名は用語 item が固定し、執筆はそれに従う

## 仕様

同じ概念が箇所ごとに違う語で呼ばれると、読み手は同じものを指しているのか判断できず、検索も
当たらなくなる。正名を 1 箇所で固定し、執筆はそこに従う。

**避けるべき同義語まで列挙する**のが要点になる。正名だけを挙げても、書き手は「これも同じ意味
だろう」と別の語を選べる。実際に揺れた語を避けるべき側として明示することで、規範が判断を
残さない形になる。

用語 item の無い語が要るという事態は、それ自体が信号になる。プロジェクトが使わない語彙を
持ち込もうとしているか、本当に語彙が欠けているかのどちらかで、どちらもその場で綴りを決めて
先へ進む扱いにはしない。

**本 item のスコープは執筆規約に限る。** コマンドの出力にどの文言を用いるかは、利用者が
直接触れる振る舞いであり個別の `requirement` の領分になる。本 item が縛るのは、そこで語を
選ぶときに参照する正名の所在と、リポジトリ全体の執筆でそれに従うことだけになる。

正名を述べる文書が複数あること自体は妨げない。索引や他の文書が同じ語を挙げてもよく、規範が
求めるのは**綴りを固定する一次が 1 つであること**になる。各語の英語と日本語の正名・避ける綴り・
定義は本 item の規範に含めず、`docs/ubiquitous-language/` の用語 item が持つ。語の一覧と item への
索引は `docs/glossary.md` が持つ。

## 出典

`docs/glossary.md` 冒頭（README・コードコメント・コマンド出力の表記基準としての正名宣言と
避けるべき同義語）、`CONTEXT.md` 冒頭（正名と避けるべき同義語の固定）、
`docs/agents/domain.md`（CONTEXT は意味を、glossary は綴りを正とするという分担）が実運用してきた
規範を、Issue #272 で quality item として立てたもの。
