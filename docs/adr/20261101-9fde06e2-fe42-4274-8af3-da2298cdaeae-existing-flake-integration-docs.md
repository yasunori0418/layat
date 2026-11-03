---
id: "ADR-9fde06e2-fe42-4274-8af3-da2298cdaeae"
type: adr
name: "既存 flake への組込手順を docs に置く"
status: 採用
origin: "垂直トレーサー弾の着手前の横断検査（実装前残セマンティクスの第 6 巡）。ADR-0024 §6 を書き直した"
justifies:
  - "REQ-6be1cbf1-6c6e-498b-8acb-7f4b80037169"
supersedes:
  - "ADR-0024"
---
# ADR-9fde06e2-fe42-4274-8af3-da2298cdaeae: 既存 flake への組込手順を docs に置く

- ステータス: 採用
- 日付: 2026-11-01
- 関連: `README.md`
- 起点: 垂直トレーサー弾の着手前の横断検査（実装前残セマンティクスの第 6 巡）。ADR-0024 §6 を書き直した

## 背景

layat は、プロジェクトに組み込んで repo 内へ配置する使い方を中心に据えている。導入の手段は
`layat init` だが、これは `nix flake init -t` のラッパーで、既存のファイルを上書きしない。既に
`flake.nix` がある repo へ後から組み込む方法は、`init` では賄えない。

## 決定

既に `flake.nix` がある repo へ layat を後から組み込む手順を、docs に通しで置く。CLI には既存の
flake を自動でマージする機構（`init --merge` など）を持たせない。

`layat init` がファイルを生成しないことは REQ-6be1cbf1-6c6e-498b-8acb-7f4b80037169 に置く。手順そのものは
README の「既存の flake に layat を追加する」節が持つ。

## 根拠

既存の repo への後付けは、project mode の主な導入方法になる。手順を docs に置けば、CLI の機構を
増やさずにこの導入方法を示せ、「設定を生成しない」方針も保てる。

## 棄却した代替案

### `layat init --merge` で既存の flake を自動でマージする

後付けは楽になるが、layat が利用者の flake を書き換えることになり、「設定を生成しない」方針を損なう。
