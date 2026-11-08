---
id: "DSG-9f5f949f-8b0e-4679-8f70-e0901a3d63b8"
type: design
name: "engine のエラーは発生源で 1 回だけ op と対象パスを付けて wrap し、伝搬では素通しする"
status: stable
satisfies:
  - "REQ-9341fa5d-836e-4023-af53-cc7d273438d1"
  - "REQ-b74a118a-1272-44eb-944c-7725163211c6"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T22:29:30+09:00"
---
# DSG-9f5f949f-8b0e-4679-8f70-e0901a3d63b8: engine のエラーは発生源で 1 回だけ op と対象パスを付けて wrap し、伝搬では素通しする

## 設計

適用範囲は `internal/engine/` のエラーとする。エラーの組み立ては標準ライブラリの `fmt.Errorf` と
`errors.Is` だけで行う（→ [engine は第三者依存ゼロの stdlib-only とし内部層に閉じる](../requirements/engine-stdlib-only.md)）。

- **発生源で 1 回だけ wrap する**: syscall・exec・parse に最も近い箇所で
  `fmt.Errorf("layat: cannot <op> (<target>): %w", ...)` の形にする。`<op>` は何をしようとしたか、
  `<target>` は対象パスを書く。stderr に出す人間向けテキストの op と対象パスは、ここで付いたものになる
  （→ [エンベロープのエラーは主体の有無で層を分けコードを分類する](../requirements/json-error-layering.md)）
- **伝搬では素通しする**: 呼び出し元は `return err` でそのまま返し、途中で再 wrap しない
- **sentinel は `%w` で透過させる**: `ErrSkipped`（= `lock.ErrLocked`）などの sentinel は常に `%w` で包み、
  呼び出し元が `errors.Is` で判定できる状態を保つ
- **監査の基準**: 利用者に見える失敗が op と対象を 1 回以上含めば規約を満たす。伝搬の途中に
  op と対象が無いことは違反にならない

## 出典

[engine のエラーは発生源で 1 回だけ wrap し、伝搬では再 wrap しない](../adr/engine-errors-wrapped-once-at-origin.md) の決定（GitHub Issue #90 の再計画で定めた規約）を、engine の中での実現として書き下した。
