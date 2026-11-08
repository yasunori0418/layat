---
id: "UBQ-fbb72ad3-f0d8-4403-8775-705fb0275ade"
type: ubiquitous_term
name: "prebuilt"
status: stable
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T01:45:24+09:00"
---

# UBQ-fbb72ad3-f0d8-4403-8775-705fb0275ade: prebuilt

## Definition

The generator behind `apply --manifest <link-farm>`: Discover returns the given path, Roots reads that link-farm's `manifest.json`, and Build / DryBuild return the path. It is not one of the `--generator` choices (→ [ADR-1ddf424a-db2c-4bcb-8823-bf7e4326b980](../adr/manifest-flag-as-prebuilt-generator.md)).

## Avoid

Listing it as a `--generator` value; describing `--manifest` as a route separate from the generators.

## 定義

`apply --manifest <link-farm>` の実体である生成器。Discover は与えられたパス、Roots はその link-farm の `manifest.json` を読み、Build / DryBuild はそのパスを返す。`--generator` の選択肢には出ない（→ [ADR-1ddf424a-db2c-4bcb-8823-bf7e4326b980](../adr/manifest-flag-as-prebuilt-generator.md)）。

## 避ける表現

`--generator` の値として挙げること、`--manifest` を生成器とは別の取得手段と説明すること。
