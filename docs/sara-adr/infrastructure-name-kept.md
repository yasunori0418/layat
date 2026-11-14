---
id: "SDR-933a9191-f60e-45bc-bebc-bc09af8154cc"
type: sara_adr
name: "infrastructure 型を改名しない"
section: graph-model
status: stable
sources:
  - id: "issue-583"
    resource: "https://github.com/yasunori0418/layat/issues/583"
    title: "#583"
  - id: "issue-237"
    resource: "https://github.com/yasunori0418/layat/issues/237"
    title: "#237"
  - id: "issue-203"
    resource: "https://github.com/yasunori0418/layat/issues/203"
    title: "#203"
origin: "epic #203 の grilling セッション（2026-08-08）で確定。ADR-0049 §3b を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
references:
  - "SDR-29ef8bb4-9349-47db-9e79-926bc71bbdd0"
  - "SDR-2457aa00-fab7-4c99-af53-2e6643c21c79"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# infrastructure 型を改名しない

- ステータス: stable
- 日付: 2026-11-03
- 関連: [infrastructure を技術基盤として再定義し root から外す](infrastructure-redefined-non-root.md), [model.yaml で組み込みモデルを全面置換し、汎用の型名を使う](model-yaml-full-replacement.md), GitHub Issue #583, #237, #203
- 起点: epic #203 の grilling セッション（2026-08-08）で確定。ADR-0049 §3b を書き直した

## 背景

infrastructure を技術基盤として定義し直した（[infrastructure を技術基盤として再定義し root から外す](infrastructure-redefined-non-root.md)）とき、型名が
内容に合っているかが問題になった。違和感の主な原因は、品質系の文書と基盤系の文書の同居だった。

## 決定

型名 infrastructure と prefix `INF` を変えない。

## 根拠

同居の違和感は、品質系の文書を quality へ分けたことで解消した。改名で残る効果は表記の問題だけである。

一方、`INF-<UUID>` は ADR の `justifies` などから参照されている。改名は ID の変更と全参照の張り替えを伴う。
得られるものに比べて壊す範囲が大きい。

## 棄却した代替案

### `platform` などへ改名する

表記は内容に近づくが、全参照の張り替えが要る。同居の解消という主な目的は改名なしで達成済みである。
