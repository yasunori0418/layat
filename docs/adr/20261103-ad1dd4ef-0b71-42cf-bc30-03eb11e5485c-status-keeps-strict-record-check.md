---
id: "ADR-ad1dd4ef-0b71-42cf-bc30-03eb11e5485c"
type: adr
name: "status は記録の検査を apply より緩めない"
status: 採用
issues:
  - "#580"
  - "#198"
origin: "Issue #198（`layat status [name]`）の実装前に、2026-07-19 の grilling 以後の実装の変化を差分 grilling で反映した（2026-10-23）。ADR-0065 §1b を書き直した"
justifies:
  - "DSG-85439cb5-a149-422e-8596-fc4e860bc3a5"
---
# ADR-ad1dd4ef-0b71-42cf-bc30-03eb11e5485c: status は記録の検査を apply より緩めない

- ステータス: 採用
- 日付: 2026-11-03
- 関連: GitHub Issue #580, #198
- 起点: Issue #198（`layat status [name]`）の実装前に、2026-07-19 の grilling 以後の実装の変化を差分 grilling で反映した（2026-10-23）。ADR-0065 §1b を書き直した

## 背景

status は表示するだけで、記録をもとに FS を変えない。壊れた記録でも、
読めるところまで表示した方が便利に見える。

## 決定

status は、apply が baseline を読むときと同じ記録の検査を使い、status 用に緩めない。
不正な記録はその config を error にし、ほかの config の列挙は続ける。

## 根拠

記録の不正は apply を止める。status がそれを警告なしに飛ばすと、利用者は apply
の時点で初めて問題に気付く。同じ検査を使えば、status で読めた記録は apply でも読める。

1 件の不正で全体を止めると、ほかの正常な config の状態まで見えなくなる。

## 棄却した代替案

### 検査を緩めて、読めるところまで表示する

壊れた記録が見えなくなり、apply の baseline と status の表示が食い違う。

### 不正な記録が 1 件でもあれば status 全体を失敗させる

ほかの config の状態を確かめられなくなる。
