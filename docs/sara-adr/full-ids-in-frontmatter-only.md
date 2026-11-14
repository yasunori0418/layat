---
id: "SDR-12bab183-d9d7-4210-a45b-c4e6568f297b"
type: sara_adr
name: "フル ID は frontmatter だけに書き、item にはリンクで触れる"
description: "フル ID の書き場所を frontmatter の id と relation リストに限り、H1 見出しは name だけにし、本文や specification などの文字列では item へ Markdown リンクで触れる。"
section: ids-and-filenames
status: stable
sources:
  - id: "issue-674"
    resource: "https://github.com/yasunori0418/layat/issues/674"
    title: "epic: item のファイル名から日付と UUID を外して <slug>.md にする"
  - id: "issue-692"
    resource: "https://github.com/yasunori0418/layat/issues/692"
    title: "epic: 散文リンク移行の後始末 — 表示崩れ・重複表記・文の破綻・frontmatter のフル ID"
origin: "H1 見出しの ID の要否の見直し（2026-11-09）で確定。[フル ID は frontmatter と H1 見出しだけに書き、item にはリンクで触れる](full-ids-in-frontmatter-and-h1.md) を書き直した"
justifies:
  - "QA-1dd63789-24ed-4a86-bf12-b62d8c526ba5"
supersedes:
  - "SDR-bd136f38-c991-46f9-ada9-06499d259a8b"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T22:23:22+09:00"
---
# フル ID は frontmatter だけに書き、item にはリンクで触れる

- ステータス: stable
- 日付: 2026-11-09
- 関連: GitHub Issue #674, #692
- 起点: H1 見出しの ID の要否の見直し（2026-11-09）で確定。[フル ID は frontmatter と H1 見出しだけに書き、item にはリンクで触れる](full-ids-in-frontmatter-and-h1.md) を書き直した

## 背景

item のファイル名に日付とフル UUID が入っていると、GitHub の URL・ファイル一覧・`index.md` のリンクの
先頭が日付と乱数で埋まり、初めて読む人は item を名前で探せない。H1 見出しの先頭に UUID があるときも同じで、
本文を開いても最初に目に入るのは乱数になる。

ID は frontmatter の `id:` が持ち、sara はファイル名も H1 見出しも見ない。どちらに ID を書かなくても、
グラフの検証と `sara query` の結果は変わらない。

`specification` / `specification_ja` / `origin` は YAML の文字列だが、人はここも文章として読む。

item の `name` には、`<name>` のように山括弧を含む語を持つものがある。リンク文字列や H1 にそのまま
書くと、Markdown がタグに解釈し、描画で語が消える。

リンクの直後に同じ name を鉤括弧や丸括弧で続けると、読み手は同じ語を二度読むことになる。

## 決定

フル ID（`<PREFIX>-<フル UUIDv4>`）を書くのは、frontmatter の `id:` と relation リストだけにする。
item 本文の H1 見出しは `# <name>` とし、ID を書かない。8 文字の省略形は使わない。

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
ファイル名の形式は CLAUDE.md の「ID 規約」表に置く。規則に反する ID をどう防ぐかは
[frontmatter の外のフル ID と省略形は起票の機械化と契約テストで防ぐ](frontmatter-only-ids-checked-by-contract-test.md) が決める。

## 根拠

ファイル名が slug だけなら、URL・一覧・リンクの文字列から内容の見当がつく。H1 見出しが name だけなら、
本文の先頭で item の主張をそのまま読める。

散文にフル ID を書くと、文の途中に長い UUID が入って読みづらい。item どうしの参照関係は frontmatter の
relation が担保するので、散文のフル ID は過剰である。Markdown リンクなら、読み手は item を名前で読める。

`specification` / `specification_ja` / `origin` も同じ形にすれば、書き方が 1 つで済む。英語の規範文に
日本語の name が入るが、sara の RFC2119 の検証には影響しない。

`<…>` は `name` と H1 の側でコードスパンに入れる。`index.md` などの生成物へ `name` がそのまま写るので、
元を直せば生成物も一緒に直る。

コードスパンの規則と繰り返しの禁止は、文字列の形で機械的に判定できる。レビューの目に頼らず契約テストに任せる。

リンクは item の改名で切れうる。epic #674 の全件移行で、すべてのリンク先を確かめた。

ファイル名・H1 見出し・散文に ID がなくても、item は frontmatter の `id:` で特定できる。ID を grep すれば `id:` と
relation に当たり、`sara query` へ渡しても同じ item に当たる。省略形に戻さないのは、sara が見るのは
正式 ID の一意性だけで、省略形を保つには sara の外の層を自前で保守し続けることになるため。

slug の重複は UUID と違って起こりうる。ただし、同じ型のディレクトリで同じ slug を取ることは少なく、
起きても起票の拒否かマージの衝突で表に出る。既存の item が気付かないうちに上書きされることはない。

## 棄却した代替案

### H1 見出しに ID を残す

本文の先頭で ID が分かるが、同じ ID は数行上の frontmatter にある。`sara init` が生成する見出しも、
ID を読むスクリプトも無いので、書き続ける手間に見合わない。

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
