---
id: "SDR-4a25760b-c309-41ff-96ef-91c5b63602c6"
type: sara_adr
name: "risk の張り先は requirement を既定にし、threatens の 1 本ごとに判定する"
description: "risk は既定で requirement に張り、設計を差し替えると消える懸念だけを design に張る。判定は threatens の 1 本ごとに行う。"
section: risk-test-defect
status: stable
sources:
  - id: "issue-649"
    resource: "https://github.com/yasunori0418/layat/issues/649"
    title: "#649"
origin: "Issue #649 の grilling で確定"
justifies:
  - "QA-d9cf5fcb-6363-49eb-9511-48b233c0c978"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-11T00:07:44+09:00"
---
# risk の張り先は requirement を既定にし、threatens の 1 本ごとに判定する

- ステータス: stable
- 日付: 2026-11-11
- 関連: GitHub Issue #649, `docs/agents/sara-graph.md`
- 起点: Issue #649 の grilling で確定

## 背景

risk は `threatens` で `requirement` と `design` のどちらも指せる。モデルは両方を許し、本数も
組み合わせも縛らない。どちらに張るかは epic #203 以来 `docs/agents/sara-graph.md` の規約として
運用してきたが、決定の記録は無かった。

## 決定

- risk は既定で `requirement` に張る。`design` に張るのは、同じ `requirement` を満たす別の設計へ
  差し替えたら消える risk に限る
- `requirement` そのものが満たされないかもしれないという懸念は、常に `requirement` に張る
- 判定は risk ごとではなく `threatens` の 1 本ごとに行う。1 つの risk が両方の種類の辺を持ってよい

## 根拠

test_condition は risk からぶら下がるので、risk を置き違えると、それを守るテストも置き違える。
`requirement` 側の risk を 1 つの design の下に置くと、同じ懸念を抱える他の設計はその懸念について
テストされていないように見える。逆に、設計に固有の risk を `requirement` に張ると、それが特定の
選択に伴う代償だという情報が失われる。

## 棄却した代替案

### 判定を risk ごとに 1 回行う

1 つの脅威の複数の面を挙げる risk は、設計を差し替えても残る面と、設計に固有の面を併せ持つ。
risk ごとに張り先を 1 つ選ぶと、どちらかの面が辺から落ちる。
