---
id: "SDR-d3f88093-dc98-4703-af45-df5b27527ffa"
type: sara_adr
name: "用語を ubiquitous_term の item で管理する"
status: 採用
issues:
  - "#647"
origin: "epic #647 の grilling（2026-11-06〜07）で確定"
justifies:
  - "QA-4a623664-650d-4a08-800f-691f4ea6ff91"
---
# SDR-d3f88093-dc98-4703-af45-df5b27527ffa: 用語を ubiquitous_term の item で管理する

- ステータス: 採用
- 日付: 2026-11-07
- 関連: GitHub Issue #647, `docs/glossary.md`, `docs/glossary.ja.md`, `CONTEXT.md`
- 起点: epic #647 の grilling（2026-11-06〜07）で確定

## 背景

用語の正名・避ける綴り・定義は、`CONTEXT.md`（日本語の長い定義）と `docs/glossary.md` /
`docs/glossary.ja.md`（英日の短い定義）の 3 ファイルが重ねて持っていた。3 ファイルの同期は
手作業で、語の数も定義の中身もすでに食い違っている。用語を 2 語足すだけで 5 ファイルを
直す必要があった（#644）。

## 決定

用語は sara の `ubiquitous_term` 型の item（prefix `UBQ`、置き場所 `docs/ubiquitous-language/`）で、
1 語を 1 ファイルとして管理する。

- 英語の正名は組み込みの `name` に置く。日本語の正名は `name_ja` に置き、英語のまま使う語では省く。
  避ける綴りは `avoid` のリストに置く
- `avoid` に入れるのは綴りだけにする。避ける理由や言い換えの説明は本文の `## Avoid` と
  `## 避ける表現` に書き、定義は `## Definition` と `## 定義` に書く
- 型は relation を 1 本も持たない根にする（`parent_types: []`・`allowed_targets: []`）。ADR や
  QA、REQ / DSG のどれともつながず、adr の `justifies` の張り先にも加えない
- 3 ファイルは `docs/glossary.md` の 1 本に縮める。中身は英語の正名・日本語の正名・item への
  リンクを主題ごとに並べた索引だけにする

## 根拠

1 語 1 ファイルにすれば、語を足すのも直すのも 1 ファイルで済む。正名をフィールドに置くと
`sara check --format json` から機械的に取り出せるので、索引との突き合わせや、将来の綴りの
自動検出に使える。`avoid` を綴りだけに限るのも同じ理由で、説明が入ると、文中の綴りと
比べる用途に使えなくなる。

relation を持たせないのは、この型を特定のプロジェクトの item に依存させないためである。
model.yaml の雛形を他のリポジトリへ切り出すとき（#500）にも、そのまま持ち出せる。sara は
「`parent_types` が空で upstream relation を持たない型」を根と判定するので、strict モードでも
orphan の error にならない。

索引を残すのは、`docs/spec.md` と同じく、語の全体を通して読む場所が要るためである。
索引は定義を持たないので、item と内容が食い違うことはない。

## 棄却した代替案

### 3 ファイルを残し、同期を契約テストで守る

ずれは検出できるが、語を足す手間は減らない。定義を 3 箇所に書く構造そのものが残る。

### 用語 item から ADR や REQ / DSG へ relation を張る

語の由来は追えるようになるが、型がプロジェクト固有の item に依存し、雛形として切り出せなくなる。
