---
id: "ADR-cd139de9-06a4-49a2-8bea-4801c8acc518"
type: adr
name: "nix 固有の失敗判定と案内文を nix 生成器の中に閉じる"
status: 採用
issues:
  - "#396"
origin: "Issue #396（評価器の診断の扱い）の結論。ADR-0055 §6 の nix 固有の判定の置き場所を書き直した"
justifies:
  - "REQ-f9920c87-8551-4aa3-bf03-26fdf4191ed6"
references:
  - "ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa"
  - "ADR-964e99a8-4aaa-46f6-b4ee-26270fc04649"
---
# ADR-cd139de9-06a4-49a2-8bea-4801c8acc518: nix 固有の失敗判定と案内文を nix 生成器の中に閉じる

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa, ADR-964e99a8-4aaa-46f6-b4ee-26270fc04649, GitHub Issue #396
- 起点: Issue #396（評価器の診断の扱い）の結論。ADR-0055 §6 の nix 固有の判定の置き場所を書き直した

## 背景

nix の失敗は、エラー文字列を見て初めて種類が分かる。experimental-features が無効なときの案内
（→ ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa）や、config 名が見つからないときの前置文も nix に固有である。
生成器の失敗は種類と案内を持つ構造化した値で CLI へ返す（→ ADR-964e99a8-4aaa-46f6-b4ee-26270fc04649）。

## 決定

nix 固有のエラー文字列による判定・案内文・config 名が見つからないときの前置文は、nix 生成器
（`internal/generator/nixgen`）の中に閉じ、外へ出さない。CLI は構造化した値の種類と案内だけを見る。

experimental-features の前提と案内の規範は REQ-f9920c87-8551-4aa3-bf03-26fdf4191ed6 に置く。

## 根拠

文字列判定と案内文は生成器ごとに異なる。CLI が持つと、生成器を足すたびに CLI が変わり、生成器の
実装事情が CLI へ漏れる。

## 棄却した代替案

### 判定と案内文を CLI に残す

nix のエラー文字列への依存が CLI に残り、生成器を足すたびに CLI に判定を書き足すことになる。
