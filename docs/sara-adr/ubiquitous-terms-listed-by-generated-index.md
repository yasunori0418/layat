---
id: "SDR-00554aea-e9c1-445a-bdf2-d64f6508fb0b"
type: sara_adr
name: "用語は ubiquitous_term の item で管理し、一覧は生成した index.md が持つ"
description: "用語は 1 語 1 ファイルの ubiquitous_term の item で管理し、語の一覧は `docs/ubiquitous-language/index.md` に生成して、`docs/glossary.md` と glossary-index の契約テストは廃止する"
status: stable
sources:
  - id: "issue-719"
    resource: "https://github.com/yasunori0418/layat/issues/719"
    title: "#719"
origin: "epic #719 の grilling（2026-11-13）で確定"
justifies:
  - "QA-4a623664-650d-4a08-800f-691f4ea6ff91"
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
supersedes:
  - "SDR-d3f88093-dc98-4703-af45-df5b27527ffa"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-13T14:41:56+09:00"
---
# 用語は ubiquitous_term の item で管理し、一覧は生成した index.md が持つ

- ステータス: stable
- 日付: 2026-11-13
- 関連: GitHub Issue #719, `docs/ubiquitous-language/index.md`, `dev/scripts/okf-index.sh`
- 起点: epic #719 の grilling（2026-11-13）で確定

## 背景

用語の正名・避ける綴り・定義は、`CONTEXT.md`（日本語の長い定義）と `docs/glossary.md` /
`docs/glossary.ja.md`（英日の短い定義）の 3 ファイルが重ねて持っていた。3 ファイルの同期は
手作業で、語の数も定義の中身も食い違っていた。用語を 2 語足すだけで 5 ファイルを
直す必要があった（#644）。

親を持たない型（根）は solution と ADR 系の 2 型だけに限ってきた。根の item は構造上いつも接続漏れの
warning になり、根の型が増えるほど本物の張り忘れが埋もれるという前提による。sara 0.12.0 では
この前提が成り立たない。`parent_types` が空で upstream relation を持たない型は根と判定され、その
item は接続漏れにならない。

[用語を ubiquitous_term の item で管理する](ubiquitous-term-type.md) は、3 ファイルを
`docs/glossary.md` の 1 本に縮め、語の一覧と item へのリンクを主題ごとに並べた索引にした。その後、
型ごとの index.md を生成するようになり、`docs/ubiquitous-language/index.md` が語の一覧を主題ごとの
節に分け、日本語の正名を添えて持つようになった。`docs/glossary.md` は同じ一覧を手で二重に持ち、
`dev/tests/glossary-index.sh` が両者の 1 対 1 を確かめていた（#719）。

## 決定

用語は sara の `ubiquitous_term` 型の item（prefix `UBQ`、置き場所 `docs/ubiquitous-language/`）で、
1 語を 1 ファイルとして管理する。

- 英語の正名は組み込みの `name` に置く。日本語の正名は `name_ja` に置き、英語のまま使う語では省く。
  避ける綴りは `avoid` のリストに置く
- `avoid` に入れるのは綴りだけにする。避ける理由や言い換えの説明は本文の `## Avoid` と
  `## 避ける表現` に書き、定義は `## Definition` と `## 定義` に書く
- 型は relation を 1 本も持たない根にする（`parent_types: []`・`allowed_targets: []`）。ADR や
  QA、REQ / DSG のどれともつながず、adr の `justifies` の張り先にも加えない
- 親を持たない型は solution と ADR 系の 2 型（adr / sara_adr）、ubiquitous_term だけにする。それ以外の型は
  `parent_types` を持ち、各 item は upstream relation を 1 本以上張る。型と relation の定義は
  `docs/model.yaml` に置く
- 語の一覧は `docs/ubiquitous-language/index.md` に生成し、`docs/glossary.md` と
  `dev/tests/glossary-index.sh` は廃止する

## 根拠

1 語 1 ファイルにすれば、語を足すのも直すのも 1 ファイルで済む。正名をフィールドに置くと
`sara check --format json` から機械的に取り出せるので、一覧の生成や、綴りの
自動検出に使える。`avoid` を綴りだけに限るのも同じ理由で、説明が入ると、文中の綴りと
比べる用途に使えなくなる。

relation を持たせないのは、この型を特定のプロジェクトの item に依存させないためである。
model.yaml の雛形を他のリポジトリへ切り出すとき（#500）にも、そのまま持ち出せる。sara は
「`parent_types` が空で upstream relation を持たない型」を根と判定するので、strict モードでも
接続漏れの error にならない。

根を solution・ADR 系の 2 型・ubiquitous_term に限るのは、親が無いことにどれも理由があるためである。solution はグラフの根、ADR 系は
規範の外にある決定の記録で `justifies` で規範へつながる。ubiquitous_term は上で述べたとおり、どの
item にも依存させない。それ以外の型に親が無いのは、受け皿の型が欠けているか張り忘れなので、strict
モードの接続漏れとして見えるほうがよい。

語の一覧を生成した index.md に置くのは、[item の一覧は生成した index.md だけが持つ](item-lists-held-by-generated-index.md)
と同じ理由による。index.md は item の `name` と `name_ja` から生成するので、語を足しても一覧を手で
直す必要が無く、item と一覧が食い違わない。手書きの索引が無くなれば、その 1 対 1 を確かめる
`dev/tests/glossary-index.sh` も要らなくなる。

## 棄却した代替案

### 3 ファイルを残し、同期を契約テストで守る

ずれは検出できるが、語を足す手間は減らない。定義を 3 箇所に書く構造そのものが残る。

### 用語 item から ADR や REQ / DSG へ relation を張る

語の由来は追えるようになるが、型がプロジェクト固有の item に依存し、雛形として切り出せなくなる。

### 手書きの `docs/glossary.md` を残し、生成した index.md と併存させる

同じ一覧を 2 か所に持つことになり、語を足すたびに索引も直す必要が残る。
