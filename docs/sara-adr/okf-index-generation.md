---
id: "SDR-52dccd7e-666c-46f8-8a8d-1a50dd0f721b"
type: sara_adr
name: "OKF の index.md を sara check の出力から生成してコミットする"
description: "OKF の index.md を sara check --format json から生成し、コミットして契約テストで差分ゼロを守る"
status: stable
sources:
  - id: "issue-662"
    resource: "https://github.com/yasunori0418/layat/issues/662"
    title: "#662"
origin: "epic #662 の grilling（2026-11-07）で確定"
justifies:
  - "QA-67715bb3-1162-4ccf-8441-2f44257a57da"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T23:10:46+09:00"
---
# SDR-52dccd7e-666c-46f8-8a8d-1a50dd0f721b: OKF の index.md を sara check の出力から生成してコミットする

- ステータス: stable
- 日付: 2026-11-07
- 関連: GitHub Issue #662, `docs/model.md`
- 起点: epic #662 の grilling（2026-11-07）で確定

## 背景

OKF の読み手は、ディレクトリごとの index.md から item を探す。sara の item には index.md が無く、
一覧は `sara query` や `sara report` で取る。sara を持たない読み手は、これらを使えない。

QA-0d42104c-a3c2-4bd1-9cad-a42d9e5a56a1 は、リファレンスドキュメントをソースから生成し、生成物を
リポジトリに持たないと定めている。

## 決定

ディレクトリごとの index.md を、`sara check --format json` の出力から自前のスクリプトで生成する。
生成した index.md はコミットし、生成し直した結果との差分が無いことを契約テストで確かめる。

- index.md には item の `name`・`description`・`status` とファイルへのリンクを載せる
- sara の relation は index.md に出さない。OKF の読み手には item の一覧だけを見せる
- `description` を持たない item は、本文の最初の段落で代用する
- index.md のコミットは QA-0d42104c-a3c2-4bd1-9cad-a42d9e5a56a1 の例外として扱う

## 根拠

OKF の読み手は、リポジトリをそのまま読む。ビルド時に生成する方式では、ビルドを実行しない読み手に
index.md が届かない。そのため QA-0d42104c-a3c2-4bd1-9cad-a42d9e5a56a1 の例外とし、生成物を
コミットする。生成物とソースの食い違いという、例外で生じる問題は、差分ゼロの契約テストで防ぐ。

`sara check --format json` は item の一覧とフィールドを機械可読で返すので、frontmatter を
自前で読み直さずに済む。relation を出さないのは、relation が sara のモデルに固有で、OKF の読み手は
その型と向きを解釈できないためである。

## 棄却した代替案

### index.md を CI で生成して配布物にだけ含める

リポジトリをそのまま読む読み手には届かない。

### relation を Markdown のリンクとして index.md に出す

relation の種類と向きが失われ、OKF の読み手には単なるリンクの束に見える。関係を正しく辿るには
結局 sara が要る。
