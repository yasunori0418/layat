---
id: "SDR-a28f6eb7-f0ab-4a3c-9a94-c8a0d0a6a66f"
type: sara_adr
name: "ID の書き場所と省略形の禁止は起票の機械化と契約テストで担保する"
description: "フル ID の書き場所と 8 文字省略形の禁止を、起票ラッパーによる機械化と契約テスト link-text.sh の検出で担保する。"
section: ids-and-filenames
status: deprecated
sources:
  - id: "issue-692"
    resource: "https://github.com/yasunori0418/layat/issues/692"
    title: "epic: 散文リンク移行の後始末 — 表示崩れ・重複表記・文の破綻・frontmatter のフル ID"
origin: "epic #692 の完了後の残存調査（2026-11-09）で確定。[省略形の禁止は起票の機械化と全面移行で担保する](short-form-ban-enforced-by-tooling.md) を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
supersedes:
  - "SDR-4656dd7b-0ea4-48fb-a6c4-a8c96ad50103"
references:
  - "SDR-bd136f38-c991-46f9-ada9-06499d259a8b"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T22:05:39+09:00"
---
# ID の書き場所と省略形の禁止は起票の機械化と契約テストで担保する

> **廃止（2026-11-09）**: 本 SDR の決定は次の後継へ移った。以下の本文は当時の記録として読む。
>
> - 全体 → [frontmatter の外のフル ID と省略形は起票の機械化と契約テストで防ぐ](frontmatter-only-ids-checked-by-contract-test.md)

- ステータス: deprecated
- 日付: 2026-11-09
- 関連: [フル ID は frontmatter と H1 見出しだけに書き、item にはリンクで触れる](full-ids-in-frontmatter-and-h1.md), GitHub Issue #692
- 起点: epic #692 の完了後の残存調査（2026-11-09）で確定。[省略形の禁止は起票の機械化と全面移行で担保する](short-form-ban-enforced-by-tooling.md) を書き直した

## 背景

フル ID の書き場所と 8 文字省略形の禁止は、[フル ID は frontmatter と H1 見出しだけに書き、item にはリンクで触れる](full-ids-in-frontmatter-and-h1.md) が決めている。本 SDR は、この規則をどう担保するかだけを決める。

規則に反する ID が書かれる場面は、人が手でファイルを作るときと、文章やコードに手で書くときの 2 つある。
`sara check` が見るのは frontmatter の relation なので、本文のコードスパンやコードブロック、コードの文字列にある ID は検出しない。

## 決定

ファイルの作成は起票ラッパーで機械化する。加えて、契約テスト `dev/tests/link-text.sh` が文書とコードを走査し、次の 2 つを検出する。

- frontmatter の `id:`・relation リストと item の H1 見出し以外に書かれたフル ID
- 8 文字省略形

架空の ID をテストデータに持つ `dev/tests/` は走査しない。ID 規約の例に使う 1 つの ID だけは、`CLAUDE.md` と `dev/skills/` で許す。

## 根拠

起票の機械化と全面移行のあとも、散文だけを見る確認では見つからないコードスパンとコードブロックにフル ID が 24 件、`tests/` の文字列に省略形が 2 件残っていた。実例が残ると、それを見て同じ書き方が増える。

ID の形は文字列だけで判定できる。Markdown を解析せず、行の中の「接頭辞・16 進 8 桁・その後に `-` も 16 進も続かない」という形で省略形を拾えば、リンクや表の折り返しで誤検出しない。`dev/tests/` を除いた全ファイルを走査した結果、検出は規則に反する箇所だけだった。

## 棄却した代替案

### フル ID だけを検出し、省略形の再発は受け入れる

省略形の検出は、フル ID と同じ走査に判定を 1 つ足すだけで済む。検出を省いても維持の手間はほとんど減らず、再発を見逃す分だけ損になる。

### 走査の対象を Markdown に限る

省略形は `tests/` のシェルと Nix の文字列にも残っていた。Markdown だけでは、この形の違反を検出できない。
