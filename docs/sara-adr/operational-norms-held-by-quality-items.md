---
id: "SDR-8777581f-a6d8-4564-8d01-48b4c4c47170"
type: sara_adr
name: "sara の運用規範は quality の item が持ち、規約文書は図と索引に縮退させる"
description: "sara の運用規範を規約文書の散文から 1 主張 1 本の quality item へ移し、規約文書は図と item への索引だけにする。"
status: stable
sources:
  - id: "issue-649"
    resource: "https://github.com/yasunori0418/layat/issues/649"
    title: "#649"
origin: "Issue #649 の grilling で確定"
justifies:
  - "QA-83707ee2-5937-405d-ba7c-2d19109155d0"
  - "QA-d9cf5fcb-6363-49eb-9511-48b233c0c978"
  - "QA-b14e0a14-b9c9-4cc6-9b7a-0ce95d30d289"
  - "QA-1f2fac5e-cfc3-46d7-89eb-34a9be7a07bf"
  - "QA-bf975ac7-1e84-4f75-9fe7-8d6154d69ba1"
  - "QA-1b4601e5-33cb-435f-812e-052cc159ce8b"
  - "QA-1dd63789-24ed-4a86-bf12-b62d8c526ba5"
  - "QA-6de13779-f0bc-4e47-8ffd-3d58de265762"
  - "QA-030f926b-5ae7-4543-b8cb-e66aae5e2b5e"
supersedes:
  - "SDR-441fb006-a2f4-44f2-bc69-bb0eeee81893"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-11T00:07:44+09:00"
---
# sara の運用規範は quality の item が持ち、規約文書は図と索引に縮退させる

- ステータス: stable
- 日付: 2026-11-11
- 関連: GitHub Issue #649, `docs/agents/sara-graph.md`, `docs/agents/adr-rules.md`, `docs/adr/README.md`
- 起点: Issue #649 の grilling で確定

## 背景

sara の運用規範は `docs/agents/sara-graph.md`・`docs/agents/adr-rules.md`・CLAUDE.md・sara-docs の
SKILL.md の散文に書かれていた。規範の型の判別、risk の張り先と採点、specification の様式、
テスト系 item の粒度、ID の書き場所、コメント規約、ADR の書き方がこれに当たる。
[規範は 1 ファイル 1 item のグラフが持ち、概要文書は索引に縮退させる](../quality/normative-items-placement-and-authoring.md) に反して、同じ規範が
複数の文書に重なり、RFC2119 の検査範囲のように食い違う記述もあった。規範を持つ item が無いので、
決定を記録した SDR は機構の item にしか `justifies` を張れなかった。

## 決定

sara の運用規範は quality の item が持つ。

- 規範は 1 主張 1 本の quality item として起こす。規範文は `specification` に置き、判定の手引き
  （灰色領域・例の表・採点の読み方）は本文に置く。既存の quality item の `specification` は変えない
- `docs/agents/sara-graph.md` は型グラフの図と quality item への索引に縮退させる。CLAUDE.md と
  スキルは規範の節を item へのリンクに置き換える
- ADR の汎用の規則は quality の item が持つ。[ADR は 1 本に 1 つの決定を 4 節で書き、justifies は規範の形成に寄与した item へ張る](../quality/adr-one-decision-four-sections-justifies-by-contribution.md) と
  [設計判断は ADR に記録し、改訂の書き戻しと item への接続を同じ変更の中で完了させる](../quality/adr-decisions-recorded-and-connected.md) がこれに当たり、`docs/agents/adr-rules.md` は
  この 2 本への索引に縮退させる
- ADR の規則の layat への当てはめ（型とディレクトリ・frontmatter の対応表・検算用の実データ・
  採番の手順）は `docs/adr/README.md` が持つ
- 規範を決めた SDR の `justifies` はその規範を持つ quality item へ張る

## 根拠

規範が item にあれば、決定と規範を `justifies` でつなげ、`sara query` で辿れる。規約文書が規範を
持つと、規範を変えるたびに重複した箇所を残らず直す必要があり、食い違いを見つける仕組みも無い。

置き換え前の決定が汎用の規則と layat への当てはめを分けたのは、規則をそのまま他のリポジトリへ
切り出すためだった。この狙いは変わらない。quality item は layat 固有の語を持たずに 1 本ずつ
切り出せ、当てはめは `docs/adr/README.md` に残る。

## 棄却した代替案

### 既存の quality item を改訂して規範を取り込む

規範の置き場所を定める既存の quality item の本文に、各規範を書き足す案。1 本が多くの主張を抱え、
決定を主張ごとに張れない。

### 規約文書に規範を残す

規範が item の外に残り、決定を張る先が無いままになる。重複と食い違いも解けない。
