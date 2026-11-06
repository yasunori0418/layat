---
id: "UBQ-f5a31977-edde-4c0f-881f-28f709b0ef0b"
type: ubiquitous_term
name: "error-wrap convention"
name_ja: "エラー wrap 規約"
---

# UBQ-f5a31977-edde-4c0f-881f-28f709b0ef0b: error-wrap convention

## Definition

Errors in `internal/engine/` are wrapped exactly once, at their origin (the point closest to the syscall / exec / parse), in the form `fmt.Errorf("layat: cannot <op> (<target>): %w", ...)`, attaching the op (what was being attempted) and the target path. Callers pass errors through unchanged (bare `return err`) by default; re-wrapping along the propagation path is not done, because it produces double-context noise such as `layat: ...: layat: ...` and breaks existing assertions on error strings. Sentinel errors (`ErrSkipped` = `lock.ErrLocked`, etc.) are always passed through with `%w`, so that callers can still test them with `errors.Is`. The audit criterion is "does the user-visible failure contain the op and target at least once"; if it does, the absence of op and target along the propagation path is not a violation of the convention.

## Avoid

Piling up context every time an error crosses a boundary; swallowing a sentinel in a wrap so that `errors.Is` can no longer find it.

## 定義

`internal/engine/` のエラーは発生源（syscall / exec / parse に最も近い箇所）で 1 回だけ `fmt.Errorf("layat: cannot <op> (<target>): %w", ...)` の形で op（何をしようとしたか）と対象パスを付けて wrap する。呼び出し元は原則として素通し（bare `return err`）にし、伝搬の途中での再 wrap は `layat: ...: layat: ...` の二重 context ノイズと既存のエラー文字列 assert 破壊を招くため行わない。sentinel エラー（`ErrSkipped` = `lock.ErrLocked` 等）は常に `%w` で透過させ、呼び出し元が `errors.Is` で判定できる状態を保つ。監査の判定基準は「ユーザーに見える失敗が op＋対象を少なくとも 1 回含むか」であり、含まれていれば伝搬の途中に op＋対象が無くても規約違反ではない。

## 避ける表現

境界を跨ぐたびに context を積み増すこと、sentinel を wrap で握り潰し `errors.Is` を辿れなくすること。
