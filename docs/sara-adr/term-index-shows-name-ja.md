---
id: "SDR-376f1a7d-ecfd-4d44-b4ca-e2f23aa4e1ec"
type: sara_adr
name: "用語の index.md はリンク文字列に name_ja を添える"
description: "用語の index.md は、name_ja を持つ語のリンク文字列を name（name_ja）にし、英語と日本語のどちらの正名からも語を引けるようにする"
status: stable
sources:
  - id: "issue-719"
    resource: "https://github.com/yasunori0418/layat/issues/719"
    title: "#719"
origin: "epic #719 の grilling（2026-11-13）で確定"
justifies:
  - "QA-4a623664-650d-4a08-800f-691f4ea6ff91"
references:
  - "SDR-52dccd7e-666c-46f8-8a8d-1a50dd0f721b"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-13T14:22:22+09:00"
---
# 用語の index.md はリンク文字列に name_ja を添える

- ステータス: stable
- 日付: 2026-11-13
- 関連: GitHub Issue #719, `dev/scripts/okf-index.sh`, `docs/ubiquitous-language/index.md`
- 起点: epic #719 の grilling（2026-11-13）で確定

## 背景

用語の item は英語の正名を `name` に、日本語の正名を `name_ja` に持つ（[用語を ubiquitous_term の item で管理する](ubiquitous-term-type.md)）。
[OKF の index.md を sara check の出力から生成してコミットする](okf-index-generation.md) で生成する
`docs/ubiquitous-language/index.md` は、リンク文字列に `name` だけを使う。そのため日本語の正名からは
語を引けない。

英語と日本語の正名を並べた一覧は、手書きの `docs/glossary.md` が持っている。

## 決定

`dev/scripts/okf-index.sh` は、`name_ja` を持つ `ubiquitous_term` の行のリンク文字列を
`name（name_ja）` にする。行の形は `* [name（name_ja）](./<file>) - <説明>` とし、括弧は全角にする。

`name_ja` を持たない語はリンク文字列を `name` だけにする。ほかの型のリンク文字列は変えない。

## 根拠

一覧の 1 行に両方の正名が並べば、どちらの綴りからでも語を見つけられる。この決定で、手書きの一覧が
担っていた役割を、生成した index.md が引き受けられる。`name_ja` は `sara check --format json` の出力に
含まれるので、frontmatter を読み直さずに取れる。

## 棄却した代替案

### `name_ja` を説明の先頭に置く

説明は `description` か本文の最初の段落から取る。そこに正名を混ぜると、語の定義と名前の区別が
付かなくなる。

### 用語の index.md だけを表の形にして正名の列を分ける

ほかの型と行の形が変わり、index.md を読むときの形が型ごとに分かれる。
