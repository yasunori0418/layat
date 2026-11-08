---
id: "ADR-796508e6-2845-467d-83f8-386930a48bfb"
type: adr
name: "src が target に重なる root 相対 entry を評価時に拒否する"
status: stable
sources:
  - id: "issue-362"
    resource: "https://github.com/yasunori0418/layat/issues/362"
    title: "#362"
origin: "Issue #362（src 側のプロジェクトルート実行時解決）の grilling（2026-10-12）。ADR-0059 §6c を書き直した"
justifies:
  - "REQ-a77181dd-a59c-40bb-855e-3430ad970eb6"
references:
  - "ADR-5e46ebd6-e44b-48dc-a93f-a8a3a3903967"
  - "ADR-6d2327f7-f513-42c3-89f8-808467db1914"
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
  - "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-796508e6-2845-467d-83f8-386930a48bfb: src が target に重なる root 相対 entry を評価時に拒否する

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-5e46ebd6-e44b-48dc-a93f-a8a3a3903967, ADR-6d2327f7-f513-42c3-89f8-808467db1914, ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd, ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0, `docs/spec.md`
- 起点: Issue #362（src 側のプロジェクトルート実行時解決）の grilling（2026-10-12）。ADR-0059 §6c を書き直した

## 背景

root 相対マーカーでは、`src/subpath` と target が同じ root の中にある。そのため、両者が同じパスを
指したり、一方が他方の祖先になったりする記述が書けてしまう。この形で symlink を張ると、リンクが
自分自身や自分の祖先を指し、symlink のループになる。

## 決定

root 相対マーカーの entry で、`src/subpath` と target が一致する場合と、一方が他方の祖先である
場合を、`normalizeManifest` が評価時に拒否する。拒否の規則は
REQ-a77181dd-a59c-40bb-855e-3430ad970eb6 に置く。

## 根拠

重なりは root の実際の値に依らず、root 相対の 2 つの文字列だけで判定できる。評価時に止めれば、
ループする symlink がディスクに作られる前に、`--dryrun` を待たず記述の誤りとして見つかる。

## 棄却した代替案

### engine の実行時に検出する

静的に判定できる誤りの発見を実行時まで遅らせ、engine に root 相対の場合だけの確認を足すことになる。

### 確認せずに張る

ループする symlink ができ、配置先を辿る操作が失敗する。原因が記述にあると気付きにくい。
