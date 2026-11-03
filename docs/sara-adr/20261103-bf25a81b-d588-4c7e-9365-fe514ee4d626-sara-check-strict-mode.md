---
id: "SDR-bf25a81b-d588-4c7e-9365-fe514ee4d626"
type: sara_adr
name: "sara check を strict モードで回す"
status: 採用
issues:
  - "#583"
  - "#255"
  - "#203"
origin: "epic #203 の移行完了（Batch 2・2026-08-08）を受けた残論点判断（Issue #255）で確定。ADR-0050 §1・§3 を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
references:
  - "SDR-a0095ee8-570c-4f77-a6c9-bbc162a5c4cb"
---
# SDR-bf25a81b-d588-4c7e-9365-fe514ee4d626: sara check を strict モードで回す

- ステータス: 採用
- 日付: 2026-11-03
- 関連: SDR-a0095ee8-570c-4f77-a6c9-bbc162a5c4cb, GitHub Issue #583, #255, #203, `sara.toml`
- 起点: epic #203 の移行完了（Batch 2・2026-08-08）を受けた残論点判断（Issue #255）で確定。ADR-0050 §1・§3 を書き直した

## 背景

`sara check` は orphan を warning に留める設定で導入された。移行の途中では、まだ親を持たない item が
正常に存在し、orphan を error にすると移行そのものが止まるためである。

移行が終わり、`sara check` は warning 0 件で通るようになった。「接続漏れ = orphan warning」は全型で
成り立つ（SDR-a0095ee8-570c-4f77-a6c9-bbc162a5c4cb）が、それを担保する仕組みは無く、接続漏れの item を
足しても警告だけで通っていた。

## 決定

`sara.toml` の `strict_mode` を `true` にする。orphan を含む全 warning が error になり、`sara check` は
接続漏れで exit 1 を返す。型を限った別のゲート（ADR の orphan だけを見る CI の grep など）や、frontmatter を
単体で検証するテストは設けない。

## 根拠

`strict_mode` は全 warning を一括で切り替える 0/1 のスイッチで、型ごとの粒度を持たない。warning が
残っていた時期には、これが strict 化を退ける理由だった。warning 0 件が成り立った今は、一括で全部落とす
挙動がそのまま望む挙動になる。既存の状態を壊さずに、以後の接続漏れだけを機械的に落とせる。

strict にすれば orphan は型を問わず error になる。型を限ったゲートはその部分集合を二重に検査するだけになる。

## 棄却した代替案

### warning に留めたまま運用で守る

接続漏れの item が警告だけで通り、不変条件が宣言のうえでしか成り立たない。

### ADR の orphan だけを CI の grep で落とす

strict の検査に含まれる。同じ判定を 2 箇所で持つと、片方だけ直したときに食い違う。
