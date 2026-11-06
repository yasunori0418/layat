# 用語（ubiquitous_term）の引き方と起こし方

用語の正名は `ubiquitous_term` 型の item（`docs/ubiquitous-language/`、1 語 1 ファイル）が固定する。
`docs/glossary.md` は語の一覧と item へのリンクだけを持つ索引で、定義は書かない。

## 索引から item を辿る

1. `docs/glossary.md` を開き、主題の節から語を探す。各行は
   `- [UBQ-<フル UUIDv4>](ubiquitous-language/<ファイル名>) — <英語の正名>（<日本語の正名>）` の形になっている
2. リンク先の item を読む。正名は frontmatter、定義と避ける表現は本文にある
3. 索引で見つからないときは、綴りが揺れている可能性があるので item 全体を検索する

```bash
rg -l -i '<探す語>' docs/ubiquitous-language/
```

## 正名と避ける綴りを機械的に取る

`sara check --format json` の `items[]` から取る。英語の正名は組み込みの `name` で、`fields` の外に
ある。日本語の正名 `name_ja` と避ける綴り `avoid` は `fields` の中にあり、どちらも任意なので
無い item もある。

```bash
sara check --format json | jq -r '
  .items[] | select(.type == "ubiquitous_term")
  | [.id, .name, (.fields.name_ja // ""), ((.fields.avoid // []) | join(", "))] | @tsv'
```

## frontmatter と本文の様式

- `name`: 英語の正名。英語の正名が一次になる
- `name_ja`: 日本語の正名。英語のまま使う語では省く
- `avoid`: 執筆で使わない綴りのリスト。綴りだけを書き、「〜と説明すること」のような説明は本文に書く
- relation は持たない。ADR や QA、REQ / DSG へ辺を張らない

本文は `# UBQ-<フル UUIDv4>: <name>` の見出しの下に、次の 4 節をこの順で置く。`sara init` が
生成する `## Overview` と `## Name (ja)` の雛形は消す。

| 見出し | 書くこと |
|---|---|
| `## Definition` | 英語の定義 |
| `## Avoid` | 避ける表現とその理由（英語）。無ければ「なし」 |
| `## 定義` | 日本語の定義 |
| `## 避ける表現` | 避ける表現とその理由（日本語）。無ければ「なし」 |

## 執筆での使い方

- 語は item の正名で書く。`avoid` にある綴りは使わない
- 必要な語の item が無ければ、その場で綴りを決めずに item を起こすか、既存の語を選ぶ

## 語が無いときに item を起こす

```bash
sara-new ubiquitous_term <slug> docs/ubiquitous-language -- \
  --name '<英語の正名>' [--name-ja '<日本語の正名>'] [--avoid '<綴り>' ...]
```

`--avoid` は綴り 1 つごとにフラグを繰り返す。起こしたら本文の 4 節を書き、`docs/glossary.md` の
該当する節に 1 行を足して、`sara check` が通ることを確かめる。
