---
id: "ADR-cd139de9-06a4-49a2-8bea-4801c8acc518"
type: adr
name: "nix 固有の失敗判定と案内文を nix 生成器の中に閉じる"
status: stable
issues:
  - "#396"
origin: "Issue #396（評価器の診断の扱い）の結論。ADR-0055 §6 の nix 固有の判定の置き場所を書き直した"
justifies:
  - "REQ-f9920c87-8551-4aa3-bf03-26fdf4191ed6"
references:
  - "ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa"
  - "ADR-964e99a8-4aaa-46f6-b4ee-26270fc04649"
  - "ADR-28605860-770e-41df-95dd-7dc05b20497a"
  - "ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9"
  - "ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1"
  - "ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246"
  - "ADR-6b7becf0-f029-41e3-895e-b672dce94d3d"
  - "ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3"
  - "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-cd139de9-06a4-49a2-8bea-4801c8acc518: nix 固有の失敗判定と案内文を nix 生成器の中に閉じる

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa, ADR-964e99a8-4aaa-46f6-b4ee-26270fc04649, ADR-28605860-770e-41df-95dd-7dc05b20497a, ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9, ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1, ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246, ADR-6b7becf0-f029-41e3-895e-b672dce94d3d, ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3, ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0, ADR-fe035038-4415-4a1e-aa5c-509fdfce9955, GitHub Issue #396
- 起点: Issue #396（評価器の診断の扱い）の結論。ADR-0055 §6 の nix 固有の判定の置き場所を書き直した

## 背景

nix の失敗は、エラー文字列を見て初めて種類が分かる。experimental-features が無効なときの案内
（→ ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa）や、config 名が見つからないときの前置文も nix に固有である。
生成器の失敗は種類と案内を持つ構造化した値で CLI へ返す（→ ADR-964e99a8-4aaa-46f6-b4ee-26270fc04649）。

## 決定

nix 固有のエラー文字列による判定・案内文・config 名が見つからないときの前置文は、nix 生成器
（`internal/generator/nixgen`）の中に閉じ、外へ出さない。CLI はエラー文字列を判定せず、生成器が
返す構造化した値だけを見る。

experimental-features の前提と案内の規範は REQ-f9920c87-8551-4aa3-bf03-26fdf4191ed6 に置く。

## 根拠

文字列判定と案内文は生成器ごとに異なる。CLI が持つと、生成器を足すたびに CLI が変わり、生成器の
実装事情が CLI へ漏れる。

## 棄却した代替案

### 判定と案内文を CLI に残す

nix のエラー文字列への依存が CLI に残り、生成器を足すたびに CLI に判定を書き足すことになる。
