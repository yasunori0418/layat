---
id: "UBQ-f5a31977-edde-4c0f-881f-28f709b0ef0b"
type: ubiquitous_term
name: "error-wrap convention"
section: errors
status: stable
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-14T23:03:48+09:00"
---

# error-wrap convention

## Definition

The convention that an error in layat's Go code is wrapped exactly once, at the lowest function that knows both the op and the target, and is not re-wrapped as it propagates.

## Avoid

Piling up context every time an error crosses a boundary; swallowing a sentinel in a wrap so that `errors.Is` can no longer find it.

## 定義

layat の Go のコードのエラーを、op と対象の両方が分かるいちばん下の関数で 1 回だけ wrap し、伝搬の途中では再 wrap しない規約。

## 避ける表現

境界を跨ぐたびに context を積み増すこと、sentinel を wrap で握り潰し `errors.Is` を辿れなくすること。
