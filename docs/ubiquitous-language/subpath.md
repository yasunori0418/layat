---
id: "UBQ-12218337-953b-40c8-b9ec-6d8ee8ec0d4e"
type: ubiquitous_term
name: "subpath"
status: stable
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T01:45:24+09:00"
---

# UBQ-12218337-953b-40c8-b9ec-6d8ee8ec0d4e: subpath

## Definition

The relative path selecting which path inside src to take, for an entry. Works for both files and directories; default `"."`. Omitting `subpath` is the canonical way to select the whole repository (`subpath = "."` is the explicit form).

## Avoid

The old names `source` / `dir`; confusing it with src; assuming a dedicated token/marker is needed for "whole repo" (omission expresses it).

`src` and `subpath` are distinct. `src` = which thing (store path / repository); `subpath` = which path inside it. Similar names, orthogonal concepts. Do not use the old name `source` (= today's `subpath`).

## 定義

entry で src 内のどのパスを取り出すかを選ぶ相対パス。file / dir 両対応。デフォルト `"."`。リポジトリ全体を選ぶ canonical な方法は subpath を省略すること（`subpath = "."` は明示形）。

## 避ける表現

旧名 `source` / `dir` を使うこと、src と混同すること、「全体を表すには専用のトークン / marker が要る」と考えること（省略で表せる）。

`src` と `subpath` は別概念で、`src` はどの物（store パス / リポジトリ）か、`subpath` はその中のどのパスかを表す。名前が似ているが直交する。旧名 `source`（= 現 `subpath`）は使わない。
