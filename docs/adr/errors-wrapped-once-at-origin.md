---
id: "ADR-0ad0321a-97f6-4db9-b46c-c9b01c2c28b9"
type: adr
name: "エラーは op と対象が分かるいちばん下の関数で 1 回だけ wrap し、伝搬では再 wrap しない"
description: "layat の Go のコード全体で、エラーは op と対象の両方が分かるいちばん下の関数で 1 回だけ包み、パッケージを跨ぐときは呼ばれる側の公開関数が包み、伝搬では再 wrap せず、sentinel は %w で透過させる"
section: dev-build
status: stable
sources:
  - id: "issue-677"
    resource: "https://github.com/yasunori0418/layat/issues/677"
    title: "#677 Go のエラー処理を規約に揃える"
origin: "Issue #677 の grilling（2026-11-14）。[engine のエラーは発生源で 1 回だけ wrap し、伝搬では再 wrap しない](engine-errors-wrapped-once-at-origin.md) の発生源での wrap の決定を Go 全体へ広げて書き直した"
justifies:
  - "DSG-9f5f949f-8b0e-4679-8f70-e0901a3d63b8"
supersedes:
  - "ADR-5228fec0-09ad-494f-8c8d-d887fc009999"
references:
  - "ADR-fdb8003f-db36-424b-8484-6f2fd57a3027"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-14T23:00:46+09:00"
---
# エラーは op と対象が分かるいちばん下の関数で 1 回だけ wrap し、伝搬では再 wrap しない

- ステータス: stable
- 日付: 2026-11-14
- 関連: GitHub Issue #677, [engine は JSON とエラーを標準ライブラリで扱う](json-and-errors-via-stdlib.md), [error-wrap convention](../ubiquitous-language/error-wrap-convention.md)
- 起点: Issue #677 の grilling（2026-11-14）。[engine のエラーは発生源で 1 回だけ wrap し、伝搬では再 wrap しない](engine-errors-wrapped-once-at-origin.md) の発生源での wrap の決定を Go 全体へ広げて書き直した

## 背景

エラーを発生源で 1 回だけ包み、伝搬では素通しする規約は、engine の中にしか無かった。engine が呼ぶ
lock・nixprofile・generation・gitutil・paths・manifest・planner と、nixgen や `cmd/layat/` は規約の外に
あった。そのため、同じ失敗を呼ばれる側と engine の両方が包んで op が重なる箇所と、どちらも対象を
付けない箇所が混在していた。発生源を「syscall・exec・parse に最も近い箇所」としていたので、op を
知らないヘルパーの中で包むのか、呼び出し元で包むのかも決まらなかった。

## 決定

layat の Go のコード全体（`internal/` の全パッケージと `cmd/layat/`）で、エラーを次のように扱う。

- 発生源は op と対象の両方が分かるいちばん下の関数とし、そこで 1 回だけ包む。その下のヘルパーは
  包まずに返す
- パッケージを跨ぐときは、呼ばれる側の公開関数が包み、呼ぶ側は素通しする
- 伝搬の途中で再 wrap しない
- sentinel は `%w` で透過させ、`errors.Is` で判定できる状態を保つ
- 監査では、利用者に見える失敗が op と対象を 1 回以上含むかを基準にする

規範の細目は [エラーは op と対象が分かるいちばん下の関数で 1 回だけ wrap し、伝搬では素通しする](../design/error-wrap-once-at-origin.md) に置く。

## 根拠

op と対象がいちばん正確に分かるのは、その両方を引数や変数として持つ関数である。そこで 1 回包めば
失敗の文脈は足り、syscall の直上のヘルパーまで下りて包む必要はない。パッケージの境界では、呼ばれる側が
自分の op と対象を知っている。呼ばれる側が包むと決めておけば、呼ぶ側が同じ文脈を重ねずに済む。
sentinel を `%w` で包めば、文字列を加えても `errors.Is` による判別は崩れない。

## 棄却した代替案

### 規約を engine の中に留める

engine が呼ぶパッケージの包み方が揃わず、境界ごとに文脈の重なりと欠けが残る。engine の側で
呼ばれる側の失敗を包み直すことになり、伝搬では再 wrap しないという規則と両立しない。

### syscall に最も近い箇所で包む

syscall の直上のヘルパーは対象のパスを持っていても、どの操作の一部として呼ばれたかを知らない。
再帰で下りるヘルパーごとに包むと、1 つの失敗に同じ op が何段も重なる。

### 全境界で再 wrap する

境界を越えるたびに文脈を足すと、同じ op と対象が二重に並んで読みにくくなる。エラーの文字列を
比べている既存の assert も通らなくなる。
