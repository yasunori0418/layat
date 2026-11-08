---
id: "ADR-f607b605-c69e-4c7e-ac89-75b7bce5bee1"
type: adr
name: "終了コードは 0 / 1 / 2 の 3 値にする"
status: stable
sources:
  - id: "issue-550"
    resource: "https://github.com/yasunori0418/layat/issues/550"
    title: "#550"
origin: "実装前残セマンティクス確定の第5巡（2026-06-14）で確定。ADR-0023 §2 と ADR-0006「dry-run」節を書き直した"
justifies:
  - "REQ-2c5a10d8-112b-4f96-947a-aba7164779c4"
  - "REQ-02a33511-0941-4813-b289-a05eb8e9aa57"
references:
  - "ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71"
  - "ADR-feb60906-1517-46b8-8fa5-e007df4d3e11"
  - "ADR-79af04fe-7ba0-4953-bb0d-75bda1707cd7"
  - "ADR-5d28818e-f7e1-45e8-a7fd-bbb6157b8586"
  - "ADR-c6bc7060-7af4-4889-9ce6-5800601c9dcc"
  - "ADR-9e328362-2e90-441a-8eca-40081ce71b1b"
  - "ADR-baf6346d-783b-4607-ab1c-c2c93c09e064"
  - "ADR-bd8a4c8a-5e7e-485a-bccf-20dab63125f2"
  - "ADR-64cff3f4-5997-49f4-b23a-e6ebeced7df9"
  - "ADR-acc31876-df36-485e-b356-92ae60e26db7"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# ADR-f607b605-c69e-4c7e-ac89-75b7bce5bee1: 終了コードは 0 / 1 / 2 の 3 値にする

- ステータス: stable
- 日付: 2026-11-01
- 関連: ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71, ADR-feb60906-1517-46b8-8fa5-e007df4d3e11, ADR-79af04fe-7ba0-4953-bb0d-75bda1707cd7, ADR-5d28818e-f7e1-45e8-a7fd-bbb6157b8586, ADR-c6bc7060-7af4-4889-9ce6-5800601c9dcc, ADR-9e328362-2e90-441a-8eca-40081ce71b1b, ADR-baf6346d-783b-4607-ab1c-c2c93c09e064, ADR-bd8a4c8a-5e7e-485a-bccf-20dab63125f2, ADR-64cff3f4-5997-49f4-b23a-e6ebeced7df9, ADR-acc31876-df36-485e-b356-92ae60e26db7, GitHub Issue #550
- 起点: 実装前残セマンティクス確定の第5巡（2026-06-14）で確定。ADR-0023 §2 と ADR-0006「dry-run」節を書き直した

## 背景

layat は CI の事前の確認と devShell の shellHook の両方から呼ばれる。CI は `--dryrun` で
「apply できるか」を判定したい。shellHook は、別の apply の実行中に待たずスキップした場合を
失敗として扱いたくない。呼び出し側がこれらを終了コードだけで区別できる必要がある。

## 決定

終了コードは 0 / 1 / 2 の 3 値にする。0 は成功（no-op と正常なスキップを含む）、1 は一般エラー、
2 は `--dryrun` で conflict を見つけたときに使う。

どの状況にどの値を割り当てるかの表は REQ-2c5a10d8-112b-4f96-947a-aba7164779c4 に、`--dryrun` が
conflict で非 0 になる規範は REQ-02a33511-0941-4813-b289-a05eb8e9aa57 に置く。

## 根拠

conflict を一般エラーと別の値にすると、CI は「apply すれば止まる箇所がある」と「layat 自体が
失敗した」を区別できる。`--no-wait` のスキップを 0 にすれば、shellHook でシェルに入ることが失敗扱いに
ならない。

値を 3 つに絞るのは、呼び出し側が分岐に使う区別がこの 3 つで足りるためである。`--json` の
エンベロープ（ADR-feb60906-1517-46b8-8fa5-e007df4d3e11）は status を終了コードに連動させるだけで、この体系を変えない。

## 棄却した代替案

### conflict も一般エラーと同じ 1 にする

CI が apply できない理由を終了コードから区別できず、`--dryrun` を事前の判定に使えなくなる。
