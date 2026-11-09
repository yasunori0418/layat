---
id: "UBQ-f5a31977-edde-4c0f-881f-28f709b0ef0b"
type: ubiquitous_term
name: "error-wrap convention"
status: stable
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T22:24:21+09:00"
---

# error-wrap convention

## Definition

The convention that an error in the engine is wrapped exactly once, at its origin, with the op and the target path attached, and is not re-wrapped as it propagates.

## Avoid

Piling up context every time an error crosses a boundary; swallowing a sentinel in a wrap so that `errors.Is` can no longer find it.

## 定義

engine のエラーを発生源で 1 回だけ op と対象パスを付けて wrap し、伝搬の途中では再 wrap しない規約。

## 避ける表現

境界を跨ぐたびに context を積み増すこと、sentinel を wrap で握り潰し `errors.Is` を辿れなくすること。
