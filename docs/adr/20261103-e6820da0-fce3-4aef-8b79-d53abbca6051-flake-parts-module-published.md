---
id: "ADR-e6820da0-fce3-4aef-8b79-d53abbca6051"
type: adr
name: "layat output の flake-parts module を flakeModules.default で公開する"
status: 採用
issues:
  - "#26"
origin: "Issue #26。ADR-0029 §1 と ADR-0029 §4 を書き直した"
justifies:
  - "DSG-16373ec2-3496-4b12-b3b1-ef74e0435b58"
  - "DSG-aeb5e219-4784-4950-845c-35f9bab9179c"
references:
  - "ADR-e683ac05-c0e0-41c1-8b6a-38b605972fe1"
---
# ADR-e6820da0-fce3-4aef-8b79-d53abbca6051: layat output の flake-parts module を flakeModules.default で公開する

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-e683ac05-c0e0-41c1-8b6a-38b605972fe1, GitHub Issue #26
- 起点: Issue #26。ADR-0029 §1 と ADR-0029 §4 を書き直した

## 背景

named manifest は `layat.<system>.<name>` という専用の output に公開する
（ADR-e683ac05-c0e0-41c1-8b6a-38b605972fe1）。`layat` は標準の flake output でも、flake-parts の
module が型を登録した output でもないので、`perSystem` に書いても top-level へ転置されない。

そのため layat 自身の dogfood は、top-level の `flake` ブロックで system ごとに
`legacyPackages.${system}` を直書きして `mkManifest` に渡していた。`pkgs` を `packages.layat` 側と
二重に解決し、dogfood の定義も CLI パッケージと別の層に散っていた。

flake-parts には、`perSystem.<name>` の値をそのまま `flake.<name>.<system>` へ転置する公式の
ヘルパー `mkTransposedPerSystemModule` がある。

## 決定

`mkTransposedPerSystemModule` で `perSystem.layat` を `flake.layat.<system>` へ転置する module を
`modules/flake-parts.nix` に置き、`flakeModules.default` として利用者へ公開する。

layat 自身の flake もこの module で dogfood する。dogfood の import は、`self.flakeModules.default` を
経由せず、公開と同じパスを直接参照する。

outputs の構成は DSG-16373ec2-3496-4b12-b3b1-ef74e0435b58 に、`modules/` の中での位置付けは
DSG-aeb5e219-4784-4950-845c-35f9bab9179c に置く。

## 根拠

公式のヘルパーを使えば、`packages` などの標準 output と同じ転置の仕組みをそのまま使える。
module は数行で済み、flake-parts 側の変更にも追従しやすい。

公開すれば、flake-parts を使う利用者も `pkgs` の二重解決を避けられる。dogfood が公開版と同じ
module を使うので、利用者と同じ使い方を layat 自身が毎回評価して確かめることになる。

同じパスを直接 import するのは、`self` の outputs を評価する途中で `self` の outputs を参照する
循環を避けるためである。

## 棄却した代替案

### dogfood の内部専用にとどめ、公開しない

layat 自身の `pkgs` の二重解決は消えるが、flake-parts を使う利用者には同じ問題が残る。利用者向けの
導線が増える価値の方が大きい。

### `self.flakeModules.default` 経由で自分を import する

公開と import の参照先が 1 つの名前に揃うが、`self` の outputs の評価で循環を招くおそれがある。

### `mkPerSystemOption` と `transposition` を自前で組む

公式ヘルパーの中身を手書きすることになる。得られるものが無く、flake-parts 側の変更に自分で
追従しなければならない。
