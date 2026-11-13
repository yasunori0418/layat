---
id: "UBQ-6e764ddd-03ce-4bd1-be53-64d651783241"
type: ubiquitous_term
name: "src"
section: input-output
status: stable
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T01:45:24+09:00"
---

# src

## Definition

The placement source of an entry (which store path or repository). The default is a store link into the Nix store; out-of-store is opt-in via an explicit marker. It is orthogonal to subpath; do not read `src` as an abbreviation of `source`.

## Avoid

Confusing it with subpath; calling it `source`.

`src` and `subpath` are distinct. `src` = which thing (store path / repository); `subpath` = which path inside it. Similar names, orthogonal concepts. Do not use the old name `source` (= today's `subpath`).

## 定義

entry の配置元（どの store パス / リポジトリか）。デフォルトは Nix ストアへの store link。out-of-store は明示マーカーで opt-in する。subpath とは直交する別概念であり、`src` を `source` の短縮形と読まない。

## 避ける表現

subpath と混同すること、`source` と呼ぶこと。

`src` と `subpath` は別概念で、`src` はどの物（store パス / リポジトリ）か、`subpath` はその中のどのパスかを表す。名前が似ているが直交する。旧名 `source`（= 現 `subpath`）は使わない。
