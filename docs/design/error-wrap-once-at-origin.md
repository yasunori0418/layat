---
id: "DSG-9f5f949f-8b0e-4679-8f70-e0901a3d63b8"
type: design
name: "エラーは op と対象が分かるいちばん下の関数で 1 回だけ wrap し、伝搬では素通しする"
description: "layat の Go のコード（internal/ の全パッケージと cmd/layat/）では、エラーを op と対象の両方が分かるいちばん下の関数で 1 回だけ包み、呼び出し元は素通しし、sentinel は %w で透過させる"
section: core-logic
status: stable
satisfies:
  - "REQ-9341fa5d-836e-4023-af53-cc7d273438d1"
  - "REQ-b74a118a-1272-44eb-944c-7725163211c6"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-14T22:59:02+09:00"
---
# エラーは op と対象が分かるいちばん下の関数で 1 回だけ wrap し、伝搬では素通しする

## 設計

適用範囲は layat の Go のコード全体で、`internal/` の全パッケージと `cmd/layat/` を含む。エラーは
標準ライブラリの `fmt.Errorf` の `%w`・`errors.Is`・`errors.As` で組み立てて判別し、エラーを作る
helper や独自の枠組みは足さない。engine はこれを標準ライブラリだけで行う
（→ [engine は第三者依存ゼロの stdlib-only とし内部層に閉じる](../requirements/engine-stdlib-only.md)）。

- **発生源で 1 回だけ wrap する**: 発生源は、op（何をしようとしたか）と対象の両方が分かる
  いちばん下の関数である。そこで `fmt.Errorf` により 1 回だけ包み、その下のヘルパーは包まずに返す。
  包む文字列の書式は [エラーの文字列は失敗の種類ごとの書式で書き、括弧の対象は原因が対象を持たないときだけ添える](error-message-format-by-failure-kind.md) に従う。
  stderr に出す人間向けテキストの op と対象は、ここで付いたものになる
  （→ [エンベロープのエラーは主体の有無で層を分けコードを分類する](../requirements/json-error-layering.md)）
- **パッケージを跨ぐときは呼ばれる側が包む**: 呼ばれる側の公開関数が自分の op と対象で包み、
  呼ぶ側は素通しする
- **伝搬では素通しする**: 呼び出し元は `return err` でそのまま返し、途中で再 wrap しない
- **sentinel は `%w` で透過させる**: `lock.ErrLocked` などの sentinel は常に `%w` で包み、
  呼び出し元が `errors.Is` で判定できる状態を保つ
- **`layat:` は持たない**: エラーの文字列は `layat:` で始めない。`layat:` は stderr へ書く場所で付ける
  （→ [stderr へ書く行は cmd の 1 箇所で `layat:` を付け、単発の失敗・`--all` の失敗行・status の行・警告が同じ整形を通る](stderr-lines-prefixed-in-one-place.md)）
- **監査の基準**: 利用者に見える失敗が op と対象を 1 回以上含めば規約を満たす。伝搬の途中に
  op と対象が無いことは違反にならない

## 出典

[エラーは op と対象が分かるいちばん下の関数で 1 回だけ wrap し、伝搬では再 wrap しない](../adr/errors-wrapped-once-at-origin.md) の決定を、Go のコード全体での実現として書き下した。
