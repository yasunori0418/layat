---
id: "QA-1dd63789-24ed-4a86-bf12-b62d8c526ba5"
type: quality
name: "フル ID は frontmatter だけに書き、散文では item にリンクで触れる"
description: "ID はフル UUIDv4 で frontmatter だけに書き、ファイル名は slug とし、散文では name を文字列にしたリンクで item に触れる"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
specification: |
  An item ID SHALL take the form `<PREFIX>-<full UUIDv4>` and SHALL be written only in the
  frontmatter `id:` and in relation lists. An item's file name SHALL be its slug alone
  (`<slug>.md`). Prose SHALL refer to an item with a Markdown link whose text is the item's
  name.
specification_ja: |
  item の ID は `<PREFIX>-<フル UUIDv4>` の形としなければならず、frontmatter の `id:` と
  relation リストだけに書かなければならない。item のファイル名は slug だけ（`<slug>.md`）と
  しなければならない。散文での item の参照は、item の name を文字列にした Markdown リンクと
  しなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-10T23:55:19+09:00"
---
# フル ID は frontmatter だけに書き、散文では item にリンクで触れる

## 仕様

| 用途 | 形式 | 例 |
|---|---|---|
| 正式 ID（frontmatter の `id:`・relation リスト）| `<PREFIX>-<フル UUIDv4>` | — |
| ファイル名 | `<slug>.md` | `mkmanifest-pure-function.md` |
| 散文中の参照 | `[item の name](相対パス)` | `[mkManifest は配置データを生成する純粋関数である](../requirements/mkmanifest-pure-function.md)` |

フル ID を書くのは frontmatter の `id:` と relation リストだけにする。item 本文の H1 見出しは
`# <name>` とし、ID を書かない。8 文字の省略形は使わない。散文で item に触れるときは Markdown
リンクにし、リンク文字列にも ID を書かない。ID をそのまま grep すれば frontmatter の `id:` と
relation に当たり、`sara query` へ渡しても同じものに当たる。省略形から正式 ID を探す手順は
要らない。

- `specification` / `specification_ja` / `origin` の文字列の中で item に触れるときも、
  `[item の name](item のディレクトリからの相対パス)` の Markdown リンクにする
- `name` と H1 に `<name>` のような `<…>` を含む語があれば、コードスパンに入れる。Markdown で
  タグと解釈され、描画で消えるため
- 散文でリンクの直後に、同じ name を鉤括弧や丸括弧で繰り返さない
- 起票は ADR を含め、devShell 同梱の `sara-new` コマンドで行う。採番（`sara init` が引く
  UUIDv4）とファイル名規約の適用を一度に済ませ、`id:` と `file:` の 2 行を返す。連番を手で
  振らない（並列レーンでの採番の衝突を構造的に避けるため）

  ```bash
  sara-new <型> <slug> <配置ディレクトリ> -- --description <要約> [<sara init のオプション>...]
  sara-new requirement lock-ordering docs/requirements -- --description '<1 文の要約>'
  ```

- ADR も UUID で採番する（`ADR-<フル UUIDv4>`。sara 品質管理機構の決定は
  `SDR-<フル UUIDv4>`）。連番の `ADR-NNNN` は削除済みで、新 ADR の `origin` と「起点:」行に
  来歴として残る。ADR の粒度・本文の構成・`justifies` の選定基準は
  [ADR は 1 本に 1 つの決定を 4 節で書き、justifies は規範の形成に寄与した item へ張る](adr-one-decision-four-sections-justifies-by-contribution.md)、
  layat への当てはめは `docs/adr/README.md` が持つ

## 出典

[フル ID は frontmatter だけに書き、item にはリンクで触れる](../sara-adr/full-ids-in-frontmatter-only.md)、
[ADR は UUID だけで採番する](../sara-adr/adr-uuid-only-numbering.md)、
[frontmatter の外のフル ID と省略形は起票の機械化と契約テストで防ぐ](../sara-adr/frontmatter-only-ids-checked-by-contract-test.md)、
[sara-id を廃止し sara init を包む起票ラッパーで slug のファイル名を適用する](../sara-adr/sara-new-wrapper-applies-slug-filename.md)
が定めた ID の書き場所と採番を、CLAUDE.md の ID 規約として運用してきた。Issue #649 で
quality item として立てた。
