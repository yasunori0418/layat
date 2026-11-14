---
id: "ADR-5228fec0-09ad-494f-8c8d-d887fc009999"
type: adr
name: "engine のエラーは発生源で 1 回だけ wrap し、伝搬では再 wrap しない"
section: dev-build
status: deprecated
sources:
  - id: "issue-90"
    resource: "https://github.com/yasunori0418/layat/issues/90"
    title: "#90"
origin: "Issue #90 の再計画（2026-07-03）"
justifies:
  - "DSG-9f5f949f-8b0e-4679-8f70-e0901a3d63b8"
references:
  - "ADR-fdb8003f-db36-424b-8484-6f2fd57a3027"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-14T23:02:03+09:00"
---
# engine のエラーは発生源で 1 回だけ wrap し、伝搬では再 wrap しない

> **廃止（2026-11-14）**: 本 ADR の決定は次の後継へ移った。以下の本文は当時の記録として読む。
>
> - 発生源での 1 回の wrap・伝搬での素通し・sentinel の透過・監査の基準（適用範囲を Go 全体へ広げた） → [エラーは op と対象が分かるいちばん下の関数で 1 回だけ wrap し、伝搬では再 wrap しない](errors-wrapped-once-at-origin.md)
> - `cannot <op> (<target>)` の書式 → [エラーの文字列は `cannot <op>` を保ち、原因が対象を持たないときだけ対象を括弧で添える](error-messages-formatted-by-failure-kind.md)
> - エラーの文字列を `layat:` で始める決定 → [`layat:` はエラーの文字列に持たせず、stderr へ書く場所で付ける](layat-prefix-added-where-stderr-is-written.md)

- ステータス: deprecated
- 日付: 2026-11-07
- 関連: GitHub Issue #90, [engine は JSON とエラーを標準ライブラリで扱う](json-and-errors-via-stdlib.md), [error-wrap convention](../ubiquitous-language/error-wrap-convention.md)
- 起点: Issue #90 の再計画（2026-07-03）

## 背景

engine はエラーを `fmt.Errorf` の `%w` で wrap し、`errors.Is` で判別する
（→ [engine は JSON とエラーを標準ライブラリで扱う](json-and-errors-via-stdlib.md)）。ただ、どこで何を付けて wrap するかは決まっていなかった。
利用者に見える失敗には、何をしようとしたか（op）と対象パスが要る。

## 決定

`internal/engine/` のエラーの扱いを次の 4 点に定める。

- 発生源（syscall・exec・parse に最も近い箇所）で 1 回だけ、`fmt.Errorf("layat: cannot <op> (<target>): %w", ...)`
  の形で op と対象パスを付けて wrap する
- 呼び出し元は `return err` で素通しし、伝搬の途中で再 wrap しない
- sentinel（`ErrSkipped` = `lock.ErrLocked` など）は `%w` で透過させ、`errors.Is` で判定できる状態を保つ
- 監査では、利用者に見える失敗が op と対象を 1 回以上含むかを基準にする

規範の細目は [エラーは op と対象が分かるいちばん下の関数で 1 回だけ wrap し、伝搬では素通しする](../design/error-wrap-once-at-origin.md) に置く。

## 根拠

op と対象パスがいちばん正確に分かるのは発生源なので、そこで 1 回付ければ失敗の文脈は足りる。
伝搬のたびに付け足しても情報は増えず、同じ接頭辞が重なるだけになる。sentinel を `%w` で包めば、
文字列を加えても `errors.Is` による判別は崩れない。

## 棄却した代替案

### 全境界で再 wrap する

境界を越えるたびに context を足すと、`layat: ...: layat: ...` のように同じ context が二重に並び、
読みにくくなる。エラー文字列を比べている既存の assert も通らなくなる。
