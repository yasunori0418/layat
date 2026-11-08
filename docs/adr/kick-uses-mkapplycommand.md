---
id: "ADR-ffbb755c-f548-4ee4-8cea-f6a9b6d98adb"
type: adr
name: "モジュールの kick も mkApplyCommand で組む"
status: stable
sources:
  - id: "issue-530"
    resource: "https://github.com/yasunori0418/layat/issues/530"
    title: "#530"
origin: "Issue #530 の grilling（2026-10-25）。ADR-0068 §5a を書き直した"
justifies:
  - "DSG-e4d5db6b-5bac-4fd6-b2fb-ed07bdec30f5"
references:
  - "ADR-db293e0c-4789-45b7-93a9-adedc021344f"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T15:55:37+09:00"
---
# ADR-ffbb755c-f548-4ee4-8cea-f6a9b6d98adb: モジュールの kick も mkApplyCommand で組む

- ステータス: stable
- 日付: 2026-11-03
- 関連: [lib.mkApplyCommand で事前展開した apply コマンドを組む](lib-mkapplycommand.md), GitHub Issue #530
- 起点: Issue #530 の grilling（2026-10-25）。ADR-0068 §5a を書き直した

## 背景

HM / NixOS / nix-darwin のモジュールは、`modules/kick.nix` で link-farm を作り、engine を起動する
`apply --all --manifest` の行を組む。`lib` に公開した `mkApplyCommand`
（[lib.mkApplyCommand で事前展開した apply コマンドを組む](lib-mkapplycommand.md)）も同じ形の行を組む。

## 決定

`modules/kick.nix` も `mkApplyCommand` を `root = "all"` で呼んで行を組み、モジュールが使う
実行時フラグを後ろに足す。HM / NixOS / nix-darwin モジュールが組むコマンド行は変えない。

`kick.nix` が使う関数の置き場所は [lib は公開 API・型定義・manifest 生成・マーカー構築子・apply コマンド組み立ての 5 ファイルへ分割する](../design/lib-module-split.md) に置く。

## 根拠

コマンド行の組み立てを 1 か所にまとめれば、モジュールと devShell の間で行の形がずれない。
`--manifest` の並べ方を変えるときも、直すのは `mkApplyCommand` だけで済む。

## 棄却した代替案

### `kick.nix` が独自の組み立てを持ち続ける

モジュールの実装には手を入れずに済むが、同じ形の行を組む処理が 2 か所に残り、片方だけ変えると
モジュールと devShell で行の形が食い違う。
