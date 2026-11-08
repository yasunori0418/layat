---
id: "UBQ-3f190b7f-08dc-4b64-ab48-4be319a7d48f"
type: ubiquitous_term
name: "starter"
status: stable
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T01:40:14+09:00"
---

# UBQ-3f190b7f-08dc-4b64-ab48-4be319a7d48f: starter

## Definition

The two templates that `layat init` expands into the CWD: `templates/standalone` and `templates/project`. Each is a minimal scaffold that works right after expansion; the user trims what they do not need and grows it into their own config. layat itself generates no config; the nix `templates` mechanism does the expansion (→ [ADR-d15a216f-b5d6-4df9-9825-b23a341bec1d](../adr/templates-and-init-fixed-ref.md), [ADR-7b9ecd2c-6b77-43c4-8fb4-3ad034e618e1](../adr/templates-minimal-example-with-comments.md)).

## Avoid

なし。

## 定義

`layat init` が CWD へ展開する 2 つのテンプレート（`templates/standalone` と `templates/project`）。展開してすぐ使える最小の雛形で、利用者は不要な部分を削って自分の config に育てる。layat 自身は設定を生成せず、展開は nix の `templates` の仕組みが行う（→ [ADR-d15a216f-b5d6-4df9-9825-b23a341bec1d](../adr/templates-and-init-fixed-ref.md), [ADR-7b9ecd2c-6b77-43c4-8fb4-3ad034e618e1](../adr/templates-minimal-example-with-comments.md)）。

## 避ける表現

なし。
