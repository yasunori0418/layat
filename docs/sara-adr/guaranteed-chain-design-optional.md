---
id: "SDR-eff65cfa-2fd4-449a-94b2-d4b9948b76e1"
type: sara_adr
name: "保証する連鎖は requirement → risk → test_condition → test_case とし、design は任意の中継点にする"
description: "要件からテストまでで保証する連鎖を requirement → risk → test_condition → test_case に定め、requirement に design を必須にせず、risk を持たない design を正常とする。"
status: stable
sources:
  - id: "issue-648"
    resource: "https://github.com/yasunori0418/layat/issues/648"
    title: "epic: 要求からテストまでの連鎖を sara のモデルでつなぐ — test_plan とテストケース・設計とリスクの切れ目"
origin: "epic #648 の grilling セッション（2026-11-08）で確定"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T23:31:26+09:00"
---
# 保証する連鎖は requirement → risk → test_condition → test_case とし、design は任意の中継点にする

- ステータス: stable
- 日付: 2026-11-09
- 関連: GitHub Issue #648, `docs/model.yaml`, `docs/agents/sara-graph.md`
- 起点: epic #648 の grilling セッション（2026-11-08）で確定

## 背景

`docs/model.yaml` の辺をどこまでつなげば「要件がテストで確かめられている」と言えるのかが、決まっていなかった。

[型グラフの規約](../agents/sara-graph.md)では、`risk` は原則として `requirement` へ張り、`design` へ張るのは設計を差し替えると消える懸念に限られる。この規約に従うと、`risk` を持たない `design` が生じる。一方で `sara-gap` は、threatens されていない `design` を未カバーとして数えている。

## 決定

保証する連鎖は `requirement` → `risk` → `test_condition` → `test_case` とする。

`design` はこの連鎖の任意の中継点にする。`requirement` に `design` を必須にしない。`design` は設計に固有の `risk` を持つときだけ `threatens` を受け、`risk` を持たない `design` は正常とする。未カバーの指標に `design` を数えない。

## 根拠

テストが確かめるのは懸念であり、懸念を表す型は `risk` である。`requirement` から `risk` を経て `test_case` まで届いていれば、設計の有無にかかわらず、その要件の懸念はテストまでつながっている。

`risk` を `requirement` へ張る既定の規約と一致するので、規約に従って書いた item が未カバーに数えられることがなくなる。

## 棄却した代替案

### `requirement` に `design` を必須にする

設計の判断が無い `requirement` にも、形だけの `design` を起こすことになる。テストまでの到達は `risk` の辺で決まるので、`design` を必須にしても検証の範囲は広がらない。

### `risk` を持たない `design` を未カバーに数え続ける

設計に固有の懸念が無い `design` が、着手すべき項目として一覧に残り続ける。一覧から外すには、規約に反して `requirement` 水準の懸念を `design` へ張るしかない。
