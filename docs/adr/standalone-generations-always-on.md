---
id: "ADR-74bf4a1f-bb74-44a7-a3cb-3b3e87c24f95"
type: adr
name: "CLI 単体では世代を常時有効にする"
section: generations
status: stable
origin: "ADR-0002 §決定（standalone は世代を常時 ON）を書き直した"
justifies:
  - "REQ-1be4d678-959c-44d7-a346-44bfd95af56e"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:30:55+09:00"
---
# CLI 単体では世代を常時有効にする

- ステータス: stable
- 日付: 2026-11-01
- 起点: ADR-0002 §決定（standalone は世代を常時 ON）を書き直した

## 背景

世代管理を opt-in にし、指定が無いときは世代を取らずに配置だけを行う道もある。世代を使わない利用者には
profile が不要に見えるためである。

## 決定

CLI 単体でも世代管理を切るモードやフラグは持たず、適用は常に profile と世代を経由する。
世代を積む手順の細目は [世代は link farm derivation を layat 自前 profile へコミットして積み、前世代 manifest から古くなった配置を除去する](../requirements/generation-mechanism-link-farm-manifest.md) に置く。

## 根拠

古くなった配置の除去には前回の配置の記録が要るので、世代を取らないモードでも結局は状態の管理が要る。
世代ありとなしの 2 本のコードパスを保守することになり、得るものが無い。

## 棄却した代替案

### 世代を opt-in にする（既定は世代なし）

上の理由で状態管理は避けられず、コードパスが 2 本に分かれる。
