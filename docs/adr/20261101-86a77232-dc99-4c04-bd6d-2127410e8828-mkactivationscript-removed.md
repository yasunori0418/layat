---
id: "ADR-86a77232-dc99-4c04-bd6d-2127410e8828"
type: adr
name: "mkActivationScript を廃止する"
status: 採用
issues:
  - "#3"
origin: "Issue #3（CLI 化）。ADR-0007 §3 を書き直した"
justifies:
  - "REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645"
---
# ADR-86a77232-dc99-4c04-bd6d-2127410e8828: mkActivationScript を廃止する

- ステータス: 採用
- 日付: 2026-11-01
- 関連: GitHub Issue #3
- 起点: Issue #3（CLI 化）。ADR-0007 §3 を書き直した

## 背景

lib は、config ごとに engine を呼び出すラッパー derivation を作る `mkActivationScript` を公開していた。
利用者はこのラッパーを `nix run .#<config>` で起動して配置していた。ラッパーは
`layat apply --manifest <manifest>` を埋め込んだだけのスクリプトで、配置のロジックは持たない。

CLI が一次 UX になると、どの manifest を適用するかは、CLI が entrypoint の公開する名前から引ける。

## 決定

`mkActivationScript` を lib から取り除く。entrypoint の公開面は `lib.mkManifest` が返す manifest だけにし、
engine の起動は CLI とモジュールの配線が担う。

CLI と engine の 2 層構成は REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645 に置く。

## 根拠

ラッパーは engine の呼び出しを config の数だけ複製するだけで、CLI が manifest を名前で引けるなら
要らない。取り除けば、lib の公開面は配置データの生成だけに揃い、lib が engine のパッケージを
参照する経路も消える。

## 棄却した代替案

### ラッパーを主経路のまま残す

`nix run .#<config>` による config ごとの適用は保てるが、ラッパーを作らないと layat を使えないという
#3 の不満と、flake 以外の entrypoint への対応の要求を満たせない。
