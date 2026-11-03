---
id: "ADR-0d8f691f-304a-4b36-8393-d6aeddc4b32a"
type: adr
name: "モジュールの entries に lib と同じ型を使う"
status: 採用
origin: "ADR-0014 §4 を書き直した"
justifies:
  - "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
---
# ADR-0d8f691f-304a-4b36-8393-d6aeddc4b32a: モジュールの entries に lib と同じ型を使う

- ステータス: 採用
- 日付: 2026-11-03
- 関連: `docs/spec.md`
- 起点: ADR-0014 §4 を書き直した

## 背景

entries は lib の `mkManifest` に渡すだけでなく、HM / NixOS / nix-darwin のモジュールの
オプションからも宣言する。lib 側の entries を target をキーにした attrset にすると、モジュールの
オプションの型をどう揃えるかが残る。モジュールの config の階層は、もともと名前をキーにした
attrset である。

## 決定

モジュールのオプションで宣言する entries は、`lib/types.nix` が定義する lib と同じ
`attrsOf (submodule …)` の型を共有する。モジュールの共通オプションの集合は
REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10 に置く。

## 根拠

型を 1 か所で定義すれば、lib から使うときとモジュールから使うときで entry のフィールドや既定値、
検査が食い違わない。config の階層と entry の階層がどちらも attrset のキーで識別する形に揃い、
構造の非対称が消える。

## 棄却した代替案

### モジュールでは別の型（リストなど）を受けて lib の形へ変換する

変換のコードと型の定義が 2 つになり、片方だけを直したときに挙動がずれる。モジュールと lib で
entries の書き方が変わり、記述を持ち回れなくなる。
