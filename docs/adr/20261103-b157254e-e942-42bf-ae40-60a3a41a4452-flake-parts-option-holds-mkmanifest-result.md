---
id: "ADR-b157254e-e942-42bf-ae40-60a3a41a4452"
type: adr
name: "flake-parts module の option に mkManifest の結果を入れる"
status: 採用
issues:
  - "#26"
origin: "Issue #26。ADR-0029 §2 と ADR-0029 §3 を書き直した"
justifies:
  - "REQ-c50df875-2cb0-4e72-8a21-858359a11cae"
references:
  - "ADR-e6820da0-fce3-4aef-8b79-d53abbca6051"
---
# ADR-b157254e-e942-42bf-ae40-60a3a41a4452: flake-parts module の option に mkManifest の結果を入れる

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-e6820da0-fce3-4aef-8b79-d53abbca6051, GitHub Issue #26
- 起点: Issue #26。ADR-0029 §2 と ADR-0029 §3 を書き直した

## 背景

flake-parts module は `perSystem.layat` を `flake.layat.<system>` へ転置する
（ADR-e6820da0-fce3-4aef-8b79-d53abbca6051）。転置先の `flake.layat.<system>.<name>` は、CLI が
`nix build .#layat.<system>.<name>` でビルドする対象なので、ビルドできる derivation でなければ
ならない。`mkTransposedPerSystemModule` は option の値を加工せずに転置する。

## 決定

option の型は package の lazy な attrset にし、利用者が perSystem の `pkgs` で `mkManifest` を
呼んだ結果（derivation）をそのまま入れる。module は `mkManifest` を内部で呼ばず、`mkManifest` の
シグネチャも flake-parts のために変えない。

flake-parts 経路の規範は REQ-c50df875-2cb0-4e72-8a21-858359a11cae に置く。

## 根拠

転置先が derivation でなければならないという制約から、option に derivation を入れる形が導かれる。
この形なら公式のヘルパーがそのまま使え、module に変換の処理が要らない。

`mkManifest` が直書きの経路と flake-parts の経路で共通の唯一の公開 API になり、利用者が覚える
API の面が 1 つで済む。`pkgs` は perSystem から来るので、`packages.layat` と同じ `pkgs`（将来の
overlay や config を含む）を使い、二重に解決しない。

## 棄却した代替案

### option に `{ root, entries }` を入れ、module が `mkManifest` を呼ぶ

利用者は `pkgs` を書かずに済み、HM モジュールの `layat.configs` に近い書き味になる。しかし転置先を
derivation にするために、公式ヘルパーを使わず自前の転置と submodule から derivation への変換が
要る。`mkManifest` の直呼びと並んで API の面が 2 つになる。

### flake-parts 経路のために `mkManifest` のシグネチャを変える

`pkgs` を module から渡す形にすれば利用者の記述は減るが、直書きの経路と引数が食い違い、
`mkManifest` が経路ごとに別の API になる。
