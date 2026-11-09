---
id: "SDR-bf25a81b-d588-4c7e-9365-fe514ee4d626"
type: sara_adr
name: "sara check を strict モードで回す"
status: stable
sources:
  - id: "issue-583"
    resource: "https://github.com/yasunori0418/layat/issues/583"
    title: "#583"
  - id: "issue-255"
    resource: "https://github.com/yasunori0418/layat/issues/255"
    title: "#255"
  - id: "issue-203"
    resource: "https://github.com/yasunori0418/layat/issues/203"
    title: "#203"
origin: "epic #203 の移行完了（Batch 2・2026-08-08）を受けた残論点判断（Issue #255）で確定。ADR-0050 §1・§3 を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
references:
  - "SDR-a0095ee8-570c-4f77-a6c9-bbc162a5c4cb"
  - "SDR-2457aa00-fab7-4c99-af53-2e6643c21c79"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T22:44:41+09:00"
---
# sara check を strict モードで回す

- ステータス: stable
- 日付: 2026-11-03
- 関連: [root は solution と ADR 系の型だけにする](roots-solution-and-adr-types-only.md), [model.yaml で組み込みモデルを全面置換し、汎用の型名を使う](model-yaml-full-replacement.md), GitHub Issue #583, #255, #203, `sara.toml`
- 起点: epic #203 の移行完了（Batch 2・2026-08-08）を受けた残論点判断（Issue #255）で確定。ADR-0050 §1・§3 を書き直した

## 背景

`sara check` は接続漏れを warning に留める設定で導入された。移行の途中では、まだ親を持たない item が
正常に存在し、接続漏れを error にすると移行そのものが止まるためである。

移行が終わり、`sara check` は warning 0 件で通るようになった。「張り忘れ = 接続漏れの warning」は全型で
成り立つ（[root は solution と ADR 系の型だけにする](roots-solution-and-adr-types-only.md)）が、それを担保する仕組みは無く、張り忘れの item を
足しても警告だけで通っていた。

## 決定

`sara.toml` の `strict_mode` を `true` にする。接続漏れを含む全 warning が error になり、`sara check` は
張り忘れで exit 1 を返す。型を限った別の確認（ADR の接続漏れだけを見る CI の grep など）や、frontmatter を
単体で検証するテストは設けない。

## 根拠

`strict_mode` は全 warning を一括で切り替える 0/1 のスイッチで、型ごとの粒度を持たない。warning が
残っていた時期には、これが strict 化を退ける理由だった。warning 0 件が成り立った今は、一括で全部落とす
挙動がそのまま望む挙動になる。既存の状態を壊さずに、以後の張り忘れだけを機械的に落とせる。

strict にすれば接続漏れは型を問わず error になる。型を限った確認はその部分集合を二重に確認するだけになる。

## 棄却した代替案

### warning に留めたまま運用で守る

張り忘れの item が警告だけで通り、不変条件が宣言のうえでしか成り立たない。

### ADR の接続漏れだけを CI の grep で落とす

strict の確認に含まれ、冗長になる。
