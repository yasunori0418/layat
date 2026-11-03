---
id: "ADR-52423842-bacf-4773-bbca-bcb34ae71013"
type: adr
name: "last-applied 記録は manifest を内容ごと持つ"
status: 採用
issues:
  - "#580"
  - "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0058 §2a を書き直した"
justifies:
  - "REQ-9d622539-a1e5-4719-af51-e0cd85b66f35"
  - "DSG-3c156bbe-e72b-43dd-8cb0-2cc105314ad5"
supersedes:
  - "ADR-0058"
references:
  - "ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d"
  - "ADR-d97154cd-a2e5-4853-a682-800ff9435c89"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
  - "ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93"
---
# ADR-52423842-bacf-4773-bbca-bcb34ae71013: last-applied 記録は manifest を内容ごと持つ

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d, ADR-d97154cd-a2e5-4853-a682-800ff9435c89, ADR-fe035038-4415-4a1e-aa5c-509fdfce9955, ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93, GitHub Issue #580, #406
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0058 §2a を書き直した

## 背景

比較の基準を store の外に記録する案は、以前「store 外の可変 JSON」として却下したことがある。
その案は manifest を store パスの参照で持ったので、参照先が GC で消えると比較の基準を読めなくなり、
rollback ができなくなった。記録を持つなら、この問題を避ける形にする必要がある。

## 決定

last-applied 記録は、配置した manifest を参照ではなく内容ごと埋め込む。あわせて、その manifest
を組んだ link-farm の store パス（`linkFarm`）と、記録した時刻（`appliedAt`）を持つ。
フィールドの一覧と版は REQ-9d622539-a1e5-4719-af51-e0cd85b66f35 と
DSG-3c156bbe-e72b-43dd-8cb0-2cc105314ad5 に置く。

## 根拠

manifest を内容ごと持てば、参照していた store パスが GC で消えても比較の基準を読める。
記録の有効性は GC と無関係になる。GC root は引き続き世代が担い、
記録は比較の基準の読み出しだけを受け持つ。記録があるから GC が安全になるわけではない。

`linkFarm` は profile との比較に使う。記録と profile が食い違ったときに、
それを見つける手掛かりになる。

## 棄却した代替案

### manifest を store パスの参照で持つ

参照先が GC で消えると比較の基準を読めず、以前に却下した案と同じ問題が出る。
