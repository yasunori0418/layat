---
id: "ADR-ad1dd4ef-0b71-42cf-bc30-03eb11e5485c"
type: adr
name: "status は記録の確認を apply より緩めない"
status: 採用
issues:
  - "#580"
  - "#198"
origin: "Issue #198（`layat status [name]`）の実装前に、2026-07-19 の grilling 以後の実装の変化を差分 grilling で反映した（2026-10-23）。ADR-0065 §1b を書き直した"
justifies:
  - "DSG-85439cb5-a149-422e-8596-fc4e860bc3a5"
supersedes:
  - "ADR-0065"
references:
  - "ADR-8e253f2e-c255-482a-a0f8-0837e122e80f"
---
# ADR-ad1dd4ef-0b71-42cf-bc30-03eb11e5485c: status は記録の確認を apply より緩めない

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-8e253f2e-c255-482a-a0f8-0837e122e80f, GitHub Issue #580, #198
- 起点: Issue #198（`layat status [name]`）の実装前に、2026-07-19 の grilling 以後の実装の変化を差分 grilling で反映した（2026-10-23）。ADR-0065 §1b を書き直した

## 背景

status は表示するだけで記録をもとに FS を変えることはないので、不正な記録でも、
読めるところまで表示した方が便利に見える。

## 決定

status は、apply が比較の基準を読むときと同じ記録の確認を使い、status 用に緩めない。不正な記録は、
その config だけを error にする。ほかの config の列挙は続ける。

## 根拠

記録の不正は apply を止めるので、status がそれを警告なしに飛ばすと、利用者は apply
の時点で初めて問題に気付く。同じ確認を使えば、status で読めた記録は apply でも読める。

不正な記録の config だけを error にするのは、1 件の不正で全体を止めると、ほかの正常な config
の状態まで見えなくなるからである。

## 棄却した代替案

### 確認を緩めて、読めるところまで表示する

不正な記録が見えなくなる。apply の比較の基準と、status の表示が食い違う。

### 不正な記録が 1 件でもあれば status 全体を失敗させる

ほかの config の状態を確かめられなくなる。
