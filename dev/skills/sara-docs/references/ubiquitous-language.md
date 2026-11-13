# 用語（ubiquitous_term）の引き方と起こし方

用語の正名は `ubiquitous_term` 型の item（`docs/ubiquitous-language/`、1 語 1 ファイル）が固定する。
`docs/ubiquitous-language/index.md` は `dev/scripts/okf-index.sh` が生成する語の一覧で、手で編集しない。

## 一覧から item を辿る

1. `docs/ubiquitous-language/index.md` を開き、主題の節から語を探す。各行は
   `` * [<英語の正名>（<日本語の正名>）](./<ファイル名>) - <description> `` の形になっている。
   `name_ja` の無い語は括弧ごと省く
2. リンク先の item を読む。正名は frontmatter、定義と避ける表現は本文にある
3. 一覧で見つからないときは、綴りが揺れている可能性があるので item 全体を検索する

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
- `avoid`: 執筆で使わない綴りのリスト。綴りだけを書き、「〜と説明すること」のような説明は本文に書く。
  他の語の正名として正しく使う綴り（`copy` / `mode` など）は入れない。入れると検出が偽陽性になる
- relation は持たない。ADR や QA、REQ / DSG へ辺を張らない

本文は `# <name>` の見出しの下に、次の 4 節をこの順で置く。`sara init` が
生成する `## Overview` と `## Name (ja)` の雛形は消す。

| 見出し | 書くこと |
|---|---|
| `## Definition` | 英語の定義 |
| `## Avoid` | 避ける表現とその理由（英語）。無ければ「なし」 |
| `## 定義` | 日本語の定義 |
| `## 避ける表現` | 避ける表現とその理由（日本語）。無ければ「なし」 |

## 執筆での使い方

- 語は item の正名で書く。`avoid` にある綴りは使わない
- `avoid` の綴りは契約テスト `dev/tests/term-avoid.sh`（textlint + prh）が検出する。対象は docs・README・CLAUDE.md で、ADR・SDR・用語 item は除く
- 必要な語の item が無ければ、その場で綴りを決めずに item を起こすか、既存の語を選ぶ

## 語が無いときに item を起こす

```bash
sara-new ubiquitous_term <slug> docs/ubiquitous-language -- \
  --name '<英語の正名>' --description '<1 文の要約>' --section '<節の値>' \
  [--name-ja '<日本語の正名>'] [--avoid '<綴り>' ...]
```

`--section` は語を載せる index.md の節の値で、省かずに渡す。取りうる値は `docs/model.yaml` の
`ubiquitous_term` の enum にある。`--avoid` は綴り 1 つごとにフラグを繰り返す。起こしたら本文の
4 節を書き、`dev/scripts/okf-index.sh` で index.md を再生成して、`sara check` が通ることを確かめる。
