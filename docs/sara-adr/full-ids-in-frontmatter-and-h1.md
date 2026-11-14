---
id: "SDR-bd136f38-c991-46f9-ada9-06499d259a8b"
type: sara_adr
name: "フル ID は frontmatter と H1 見出しだけに書き、item にはリンクで触れる"
description: "フル ID の書き場所を frontmatter と H1 見出しに限り、`specification` などの文字列でも item へは Markdown リンクで触れ、ファイル名は `<slug>.md` にする。"
section: ids-and-filenames
status: deprecated
sources:
  - id: "issue-674"
    resource: "https://github.com/yasunori0418/layat/issues/674"
    title: "epic: item のファイル名から日付と UUID を外して <slug>.md にする"
  - id: "issue-692"
    resource: "https://github.com/yasunori0418/layat/issues/692"
    title: "epic: 散文リンク移行の後始末 — 表示崩れ・重複表記・文の破綻・frontmatter のフル ID"
origin: "epic #692 の grilling（2026-11-08）で確定。[item のファイル名は slug だけにする](slug-only-filenames.md) を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
supersedes:
  - "SDR-6c6c0681-4d5e-429e-8ee9-acf8dd94ad54"
references:
  - "SDR-4656dd7b-0ea4-48fb-a6c4-a8c96ad50103"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:09:48+09:00"
---
# フル ID は frontmatter と H1 見出しだけに書き、item にはリンクで触れる

> **廃止（2026-11-09）**: 本 SDR の決定は次の後継へ移った。以下の本文は当時の記録として読む。
>
> - 全体 → [フル ID は frontmatter だけに書き、item にはリンクで触れる](full-ids-in-frontmatter-only.md)

- ステータス: deprecated
- 日付: 2026-11-08
- 関連: [省略形の禁止は起票の機械化と全面移行で担保する](short-form-ban-enforced-by-tooling.md), GitHub Issue #674, #692
- 起点: epic #692 の grilling（2026-11-08）で確定。[item のファイル名は slug だけにする](slug-only-filenames.md) を書き直した

## 背景

item のファイル名に日付とフル UUID が入っていると、GitHub の URL・ファイル一覧・`index.md` のリンクの
先頭が日付と乱数で埋まり、初めて読む人は item を名前で探せない。

ID は frontmatter の `id:` が持ち、sara はファイル名を見ない。ファイル名に ID を書かなくても、
グラフの検証と `sara query` の結果は変わらない。

`specification` / `specification_ja` / `origin` は YAML の文字列だが、人はここも文章として読む。

item の `name` には、`<name>` のように山括弧を含む語を持つものがある。リンク文字列や H1 にそのまま
書くと、Markdown がタグに解釈し、描画で語が消える。

リンクの直後に同じ name を鉤括弧や丸括弧で続けると、読み手は同じ語を二度読むことになる。

## 決定

フル ID（`<PREFIX>-<フル UUIDv4>`）を書くのは、frontmatter の `id:`・relation リストと、item 本文の
H1 見出し `# <ID>: <name>` だけにする。8 文字の省略形は使わない。

散文で item に触れるときは、`[item の name や自然な語句](相対パス)` の Markdown リンクにする。
リンク文字列にも ID を書かない。

`specification` / `specification_ja` / `origin` の文字列の中で item に触れるときも、
`[item の name](item のディレクトリからの相対パス)` の Markdown リンクにする。

`name` と H1 に `<name>` のような `<…>` を含む語があれば、コードスパンに入れる。

散文では、リンクの直後に同じ name を鉤括弧や丸括弧で繰り返さない。

コードスパンの規則と繰り返しの禁止は、契約テスト `dev/tests/link-text.sh` が検出する。

item のファイル名は `<slug>.md` にし、日付と UUID を含めない。

slug は同じディレクトリの中で一意にする。`sara-new` は出力先が既に存在すれば起票を拒否する。
`index` は生成物の `index.md` と同じ名前になるので、slug に使えない。

並列のレーンが同じディレクトリに同じ slug を起こすと、起票では検出できない。この場合はマージの
add/add 衝突で検出し、片方を改名する。ここは防がずに受け入れる。

ID の規約は [トレーサビリティ検証基盤（sara によるドキュメントグラフの機械検証）](../infrastructure/traceability-verification.md) の「ID 規約」節に、
ファイル名の形式は CLAUDE.md の「ID 規約」表に置く。省略形の禁止をどう担保するかは
[省略形の禁止は起票の機械化と全面移行で担保する](short-form-ban-enforced-by-tooling.md) が決める。

## 根拠

ファイル名が slug だけなら、URL・一覧・リンクの文字列から内容の見当がつく。

散文にフル ID を書くと、文の途中に長い UUID が入って読みづらい。item どうしの参照関係は frontmatter の
relation が担保するので、散文のフル ID は過剰である。Markdown リンクなら、読み手は item を名前で読める。

`specification` / `specification_ja` / `origin` も同じ形にすれば、書き方が 1 つで済む。英語の規範文に
日本語の name が入るが、sara の RFC2119 の検証には影響しない。

`<…>` は `name` と H1 の側でコードスパンに入れる。`index.md` などの生成物へ `name` がそのまま写るので、
元を直せば生成物も一緒に直る。

コードスパンの規則と繰り返しの禁止は、文字列の形で機械的に判定できる。レビューの目に頼らず契約テストに任せる。

リンクは item の改名で切れうる。epic #674 の全件移行で、すべてのリンク先を確かめた。

ファイル名や散文に ID がなくても、item は frontmatter の `id:` で特定できる。ID を grep すれば `id:` と
relation に当たり、`sara query` へ渡しても同じ item に当たる。省略形に戻さないのは、sara が見るのは
正式 ID の一意性だけで、省略形を保つには sara の外の層を自前で保守し続けることになるため。

slug の重複は UUID と違って起こりうる。ただし、同じ型のディレクトリで同じ slug を取ることは少なく、
起きても起票の拒否かマージの衝突で表に出る。既存の item が気付かないうちに上書きされることはない。

## 棄却した代替案

### ファイル名にフル ID を残す

slug の重複は構造的に起きないが、長く読みにくいファイル名がそのまま残り、背景の問題が解けない。
ファイル名から ID を読む用途は、frontmatter の `id:` を grep すれば代わりが利く。

### `generated.at` 順の index.md

日付を外したあとの時系列の並びとして検討した。`generated.at` は本文を最後に変えたコミットの時刻で、
作成日ではない。sara の JSON にも出ない。`index.md` の並びは slug 順のままにし、時系列は ADR 本文の
「- 日付:」行と `git log` で追う。

### `specification` だけ slug のコードスパンをリンク文字列にする

規範文を英語だけで書けるが、リンクの書き方が 2 通りになる。slug から name を引く手間も読み手に残る。

### リンク文字列だけを `\<name\>` でエスケープする

リンク文字列の描画は直るが、`name` と H1 は直らず、`index.md` などの生成物にタグが残る。リンク文字列と
`name` の文字列も一致しなくなる。
