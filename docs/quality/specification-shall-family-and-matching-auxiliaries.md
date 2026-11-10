---
id: "QA-1f2fac5e-cfc3-46d7-89eb-34a9be7a07bf"
type: quality
name: "specification は SHALL 系で書き、日本語の規範文は同じ強度の規範助動詞で終える"
description: "specification は SHALL 系で書き、specification_ja は同じ強度の規範助動詞で終える"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
specification: |
  A `specification` SHALL write its mandatory strengths with the SHALL family (SHALL /
  SHALL NOT) only and SHALL NOT use the RFC 2119 synonyms of that family; SHOULD, SHOULD
  NOT and MAY SHALL keep their own strengths. In a `specification_ja`, each sentence that
  states the norm SHALL end in the normative auxiliary of the same strength, while a
  sentence that supplies background, rationale or a worked example SHALL stay in the
  declarative. The strength of a norm SHALL NOT drift between the two fields. A mention of
  another item in either field SHALL be written as a Markdown link rather than a bare ID.
specification_ja: |
  `specification` は必須の強度を SHALL 系（SHALL / SHALL NOT）だけで書かなければならず、
  MUST / MUST NOT を書いてはならない。SHOULD / SHOULD NOT / MAY はそれぞれの強度のまま
  使わなければならない。`specification_ja` で規範を述べる文は、同じ強度の規範助動詞で終え
  なければならず、背景・根拠・例を述べる文は平叙文のままにしなければならない。2 つの欄の間で
  規範の強度をずらしてはならない。どちらの欄でも、他の item への言及は ID を裸で書かず
  Markdown リンクで書かなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-10T23:08:41+09:00"
---
# specification は SHALL 系で書き、日本語の規範文は同じ強度の規範助動詞で終える

## 仕様

`requirement`・`quality`・`test_plan` の 3 型は、どれも規範を 2 度持つ。英語の `specification`
（sara は型を問わず RFC2119 キーワードの有無を検証する）と、日本語の `specification_ja`（sara は
まったく検査しない）である。どちらも規範であり、どちらも他方の訳注ではない。以下の規則は 3 型の
すべてに当てはまり、2 つの欄を 1 つの規範として読めるようにする。

これらを機械的に強制する仕組みは無い。sara が問うのは*何らかの* RFC2119 キーワードがあるか
だけで、どの綴りかは問わず、`specification_ja` はまったく見ない。したがってこれらの規則は
レビューで守る。それで緩すぎると分かったら、検査はそのために作られた散文の
linter（textlint など）に置き、自前のスクリプトは書かない。

`specification` や `specification_ja` が他の item に触れるときは、散文と同じ形の Markdown
リンク `[name](この item のディレクトリからの相対パス)` で書き、ID を裸で書かない。英語の規範の
中に日本語の `name` があっても、sara の RFC2119 の検査には影響しない。

### `specification` は SHALL 系だけを使う

`SHALL` / `SHALL NOT` と書く。RFC2119 では同義だが、`MUST` / `MUST NOT` は書かない。同じ強度に
2 つの綴りが混ざると、item 群は 2 つが違うかのように読める。`SHOULD` / `SHOULD NOT` / `MAY` は
それぞれの強度を持つので、そのまま使う。`docs/model.yaml` の `placeholder` が SHALL を使うのも
同じ理由である。

### `specification_ja` は平叙文でなく規範助動詞を使う

規範を述べる文は規範助動詞で終える。平叙文（「〜する」「〜となる」）は、実装への要求ではなく
今の実装の説明として読めてしまう。その区別を記録するために item がある。

この規則は item 単位でなく、文ごとに当てる。1 つの `specification_ja` の中で、規範を述べる文は
助動詞を取り、背景・根拠・例を述べる文は平叙文のままにする。規範を述べない文に助動詞を押し付け
ない。

### 強度の対応は固定する

| `specification` | `specification_ja` |
|---|---|
| SHALL (= MUST) | 〜しなければならない |
| SHALL NOT (= MUST NOT) | 〜してはならない |
| SHOULD / SHOULD NOT | 〜すべきである / 〜すべきでない |
| MAY | 〜してもよい |

強度が変わらない限り、文の都合で活用させてよい（例: 後ろの節へつなぐ「〜してはならず、」）。
起きてはならないのは、2 つの欄の間で強度がずれることである。`SHALL` を「〜すべきである」と訳すと、
このリポジトリの読み手の多くが実際に読む欄で規範が弱まる。

## 出典

`specification` の様式として運用してきた規約を、Issue #649 で quality item として立てた。
