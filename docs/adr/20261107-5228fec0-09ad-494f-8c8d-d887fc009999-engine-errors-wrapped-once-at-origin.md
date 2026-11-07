---
id: "ADR-5228fec0-09ad-494f-8c8d-d887fc009999"
type: adr
name: "engine のエラーは発生源で 1 回だけ wrap し、伝搬では再 wrap しない"
status: 採用
issues:
  - "#90"
origin: "Issue #90 の再計画（2026-07-03）"
justifies:
  - "DSG-9f5f949f-8b0e-4679-8f70-e0901a3d63b8"
references:
  - "ADR-fdb8003f-db36-424b-8484-6f2fd57a3027"
---
# ADR-5228fec0-09ad-494f-8c8d-d887fc009999: engine のエラーは発生源で 1 回だけ wrap し、伝搬では再 wrap しない

- ステータス: 採用
- 日付: 2026-11-07
- 関連: GitHub Issue #90, ADR-fdb8003f-db36-424b-8484-6f2fd57a3027, UBQ-f5a31977-edde-4c0f-881f-28f709b0ef0b
- 起点: Issue #90 の再計画（2026-07-03）

## 背景

engine はエラーを `fmt.Errorf` の `%w` で wrap し、`errors.Is` で判別する
（→ ADR-fdb8003f-db36-424b-8484-6f2fd57a3027）。ただ、どこで何を付けて wrap するかは決まっていなかった。
利用者に見える失敗には、何をしようとしたか（op）と対象パスが要る。

## 決定

`internal/engine/` のエラーの扱いを次の 4 点に定める。

- 発生源（syscall・exec・parse に最も近い箇所）で 1 回だけ、`fmt.Errorf("layat: cannot <op> (<target>): %w", ...)`
  の形で op と対象パスを付けて wrap する
- 呼び出し元は `return err` で素通しし、伝搬の途中で再 wrap しない
- sentinel（`ErrSkipped` = `lock.ErrLocked` など）は `%w` で透過させ、`errors.Is` で判定できる状態を保つ
- 監査では、利用者に見える失敗が op と対象を 1 回以上含むかを基準にする

規範の細目は DSG-9f5f949f-8b0e-4679-8f70-e0901a3d63b8 に置く。

## 根拠

op と対象パスがいちばん正確に分かるのは発生源なので、そこで 1 回付ければ失敗の文脈は足りる。
伝搬のたびに付け足しても情報は増えず、同じ接頭辞が重なるだけになる。sentinel を `%w` で包めば、
文字列を加えても `errors.Is` による判別は崩れない。

## 棄却した代替案

### 全境界で再 wrap する

境界を越えるたびに context を足すと、`layat: ...: layat: ...` のように同じ context が二重に並び、
読みにくくなる。エラー文字列を比べている既存の assert も通らなくなる。
