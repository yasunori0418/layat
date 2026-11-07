---
id: "ADR-f58e9981-bacc-4937-98b8-bddce0a057db"
type: adr
name: "モジュールの activation は apply --all --manifest を 1 回だけ起動する"
status: stable
issues:
  - "#585"
origin: "ADR-0060 §1 と ADR-0066 §3a を書き直した"
justifies:
  - "REQ-c847d1af-a437-46bb-bd64-42083810d034"
  - "REQ-8085f194-c903-4ecb-abd8-c719fe7b3292"
  - "REQ-e1e1114b-ba07-4d57-8e04-6e30e39a5da3"
  - "REQ-a220bb28-89de-42b0-a854-e7468755ecb1"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T21:51:56+09:00"
---
# ADR-f58e9981-bacc-4937-98b8-bddce0a057db: モジュールの activation は apply --all --manifest を 1 回だけ起動する

- ステータス: stable
- 日付: 2026-11-03
- 関連: GitHub Issue #585, `docs/spec.md`
- 起点: ADR-0060 §1 と ADR-0066 §3a を書き直した

## 背景

モジュールは config ごとにビルド済みの link-farm を持つ。これを engine へ渡す形として、config ごとに
`apply --manifest` を起動し、部分失敗を bash で集約する形も考えられる。一方で `apply --all` は、
config の集合について前段の衝突の確認・build と配置の並列化・部分失敗の集約をすでに CLI に持つ。

## 決定

home-manager / NixOS / nix-darwin の activation は、config が 1 つ以上あるとき、全 config の
link-farm を `layat apply --all --manifest <name>=<link-farm> …` の 1 回の起動で渡す。`--jobs` は
常に明示して付け、`--backup` や `--recopy` などの修飾も同じ起動に付ける。config が無いときは
起動しない。activation の側には config のループも部分失敗の集約も書かない。

組の並びと報告の順は REQ-c847d1af-a437-46bb-bd64-42083810d034、home-manager の起動の契約は
REQ-8085f194-c903-4ecb-abd8-c719fe7b3292、NixOS / nix-darwin の起動の契約は
REQ-a220bb28-89de-42b0-a854-e7468755ecb1、`--backup` の配線は
REQ-e1e1114b-ba07-4d57-8e04-6e30e39a5da3 に置く。

## 根拠

- 集合の扱いは CLI に寄せる。bash でループと集約を書くと、終了コードの優先度を含めて `apply --all` の
  機能を写し直すことになる
- 世代の単位は曖昧にならない。`apply --all` は config ごとに独立した profile へ世代をコミットし、
  全体を atomic にしない。`<name>=<link-farm>` の 1 組が 1 つの profile の世代になる
- activation も `--jobs` による並列化に乗れる
- 3 つのモジュールで起動の形を揃えるので、起動の規則がモジュールごとに分岐しない
- config が無いときに起動しないのは、`--manifest` の無い `apply --all` が entrypoint から link-farm を得る動作になるからである

## 棄却した代替案

### config ごとに `apply --manifest` を起動し、bash で集約する

`apply --all` のループ・集約・並列化を bash で再実装することになる。

### モジュールごとに起動の形を組む

規則が 3 か所に分かれ、どれか 1 つだけが変わって分岐する。
