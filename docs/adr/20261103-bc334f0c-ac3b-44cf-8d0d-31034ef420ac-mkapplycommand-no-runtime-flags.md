---
id: "ADR-bc334f0c-ac3b-44cf-8d0d-31034ef420ac"
type: adr
name: "mkApplyCommand は実行時フラグを引数に取らない"
status: stable
sources:
  - id: "issue-530"
    resource: "https://github.com/yasunori0418/layat/issues/530"
    title: "#530"
origin: "Issue #530 の grilling（2026-10-25）。ADR-0068 §4 を書き直した"
justifies:
  - "REQ-6c51facc-fa7a-4b5e-939b-44e7578d7a4f"
references:
  - "ADR-db293e0c-4789-45b7-93a9-adedc021344f"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T15:54:50+09:00"
---
# ADR-bc334f0c-ac3b-44cf-8d0d-31034ef420ac: mkApplyCommand は実行時フラグを引数に取らない

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-db293e0c-4789-45b7-93a9-adedc021344f, GitHub Issue #530
- 起点: Issue #530 の grilling（2026-10-25）。ADR-0068 §4 を書き直した

## 背景

`mkApplyCommand`（ADR-db293e0c-4789-45b7-93a9-adedc021344f）が組む `apply` の行には、並列度や
copy の再実行、退避といった実行時のフラグを足す場面がある。呼び出し側によって必要なフラグは違う。

## 決定

`mkApplyCommand` は実行時フラグを引数に取らない。必要なフラグは、呼び出し側が戻り値の文字列の
後ろに足す。

引数に取らないフラグの範囲は REQ-6c51facc-fa7a-4b5e-939b-44e7578d7a4f に置く。

## 根拠

フラグを引数にすると、CLI にフラグが増えるたびに関数の引数も追従させ続けることになる。戻り値は
コマンド行の文字列なので、後ろに足すだけで任意のフラグを組み合わせられる。

## 棄却した代替案

### 実行時フラグを引数に取る

呼び出し側が型付きの引数でフラグを渡せるが、関数の引数が CLI のフラグ追加に追従し続けることに
なる。
