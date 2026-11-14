---
id: "SDR-6c6c0681-4d5e-429e-8ee9-acf8dd94ad54"
type: sara_adr
name: "item のファイル名は slug だけにする"
description: "item のファイル名を `<slug>.md` にし、フル ID は frontmatter と H1 見出しだけに書き、散文は Markdown リンクで item を参照する。"
section: ids-and-filenames
status: deprecated
sources:
  - id: "issue-674"
    resource: "https://github.com/yasunori0418/layat/issues/674"
    title: "epic: item のファイル名から日付と UUID を外して <slug>.md にする"
origin: "issue #674 の grilling（2026-11-08）で確定"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
supersedes:
  - "SDR-722ce265-2aaf-4a90-915d-32d26b7a38bb"
references:
  - "SDR-4656dd7b-0ea4-48fb-a6c4-a8c96ad50103"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-08T20:13:24+09:00"
---
# item のファイル名は slug だけにする

> **廃止（2026-11-08）**: 本 SDR の決定は次の後継へ移った。以下の本文は当時の記録として読む。
>
> - 全体 → [フル ID は frontmatter と H1 見出しだけに書き、item にはリンクで触れる](full-ids-in-frontmatter-and-h1.md)

- ステータス: deprecated
- 日付: 2026-11-08
- 関連: [省略形の禁止は起票の機械化と全面移行で担保する](short-form-ban-enforced-by-tooling.md), GitHub Issue #674
- 起点: issue #674 の grilling（2026-11-08）で確定

## 背景

item のファイル名は `<YYYYMMDD>-<フル UUIDv4>-<slug>.md` の形だった。GitHub の URL・ファイル一覧・
`index.md` のリンクでは先頭が日付と乱数で埋まり、初めて読む人は item を名前で探せない。

ID は frontmatter の `id:` が持ち、sara はファイル名を見ない。ファイル名に ID を書かなくても、
グラフの検証と `sara query` の結果は変わらない。

## 決定

フル ID（`<PREFIX>-<フル UUIDv4>`）を書くのは、frontmatter の `id:`・relation リストと、item 本文の
H1 見出し `# <ID>: <name>` だけにする。8 文字の省略形は使わない。

散文で item に触れるときは、`[item の name や自然な語句](相対パス)` の Markdown リンクにする。
リンク文字列にも ID を書かない。

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

リンクは item の改名で切れうる。epic #674 の全件移行で、すべてのリンク先を確かめる。

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
