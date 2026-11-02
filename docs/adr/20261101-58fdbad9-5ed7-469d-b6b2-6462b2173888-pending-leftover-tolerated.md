---
id: "ADR-58fdbad9-5ed7-469d-b6b2-6462b2173888"
type: adr
name: ".pending の残留を許容する"
status: 採用
origin: "実装前レビューの第 2 巡（2026-06-14）。ADR-0016 §6 を書き直した"
justifies:
  - "REQ-840b3641-6e76-46da-82e9-680cabd65abe"
supersedes:
  - "ADR-0016"
references:
  - "ADR-0057"
---
# ADR-58fdbad9-5ed7-469d-b6b2-6462b2173888: .pending の残留を許容する

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-0057
- 起点: 実装前レビューの第 2 巡（2026-06-14）。ADR-0016 §6 を書き直した

## 背景

apply は link-farm を profileDir の `.pending` に out-link を張って build し、適用に成功したら
`.pending` を消す。コミットの前に apply が失敗すると `.pending` が残り、使われない link-farm を GC から
守り続ける。

## 決定

失敗で残った `.pending` を回収する処理は持たない。次の apply が同じパスに out-link を張り直して上書きする。

残留が config あたり最大 1 個に収まることは REQ-840b3641-6e76-46da-82e9-680cabd65abe に置く。

## 根拠

`.pending` は config ごとの固定パスなので、残っても config あたり 1 個を超えない。次の apply が成功すれば
上書きされ、コミット後に消えるので、掴んでいた link-farm も GC で回収できるようになる。残留の害が
有界なので、回収処理を持つ複雑さに見合わない。

## 棄却した代替案

### 失敗時の片付けを自前で実装する

残留は消えるが、次の apply の上書きで済む問題に対して回収処理の複雑さを負う。
