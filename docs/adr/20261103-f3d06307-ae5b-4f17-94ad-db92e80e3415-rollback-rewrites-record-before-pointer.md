---
id: "ADR-f3d06307-ae5b-4f17-94ad-db92e80e3415"
type: adr
name: "rollback はポインタ移動の前に戻り先の manifest で記録を書き直す"
status: 採用
issues:
  - "#580"
  - "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0058 §4b を書き直した"
justifies:
  - "REQ-9d622539-a1e5-4719-af51-e0cd85b66f35"
  - "REQ-0e341430-17f0-498b-9439-65491652163a"
supersedes:
  - "ADR-0058"
references:
  - "ADR-d97154cd-a2e5-4853-a682-800ff9435c89"
  - "ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
  - "ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93"
---
# ADR-f3d06307-ae5b-4f17-94ad-db92e80e3415: rollback はポインタ移動の前に戻り先の manifest で記録を書き直す

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-d97154cd-a2e5-4853-a682-800ff9435c89, ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d, ADR-fe035038-4415-4a1e-aa5c-509fdfce9955, ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93, GitHub Issue #580, #406
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0058 §4b を書き直した

## 背景

rollback は FS を N-1 に収束させてから、最後に profile
のポインタを動かす（ADR-d97154cd-a2e5-4853-a682-800ff9435c89）。記録は世代ではなく、
「最後に置いたもの」である。世代のポインタを戻しても、記録は巻き戻らない。rollback で FS が N-1
になったら、記録も N-1 の内容にする必要がある。

## 決定

rollback は、FS の収束後、profile のポインタを動かす前に、N-1 の manifest で記録を書き直す。
書けなければ rollback は失敗とする。undo journal で FS を巻き戻す。

## 根拠

書く時機を apply と同じ「FS の成功直後、世代操作の前」に揃える。ポインタを動かす前に書けば、
書き込みに失敗したときに FS と profile をどちらも元のまま残せる。

記録は明示的に書き直す。そのため、記録が世代に捕捉されている必要が無い。

## 棄却した代替案

### ポインタを動かした後に記録を書き直す

ポインタが N-1 に動いた後で記録が書けないと、記録は N のまま残り、次の apply が N
を比較の基準にして古くなった配置を除去する。

### rollback では記録を書き直さない

記録が N のまま残り、N-1 に戻った FS と食い違う。
