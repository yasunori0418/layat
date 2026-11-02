---
id: "ADR-451679d5-473b-4faf-bf2d-a6f6d3e1e131"
type: adr
name: "src の型で素の文字列を拒否する"
status: 採用
origin: "manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §2（型定義）の srcType の項を書き直した"
justifies:
  - "REQ-a33a11e3-830d-4142-88ed-4c1fc35e7f74"
  - "REQ-99ca5381-6c53-426c-b145-7b4297c53868"
references:
  - "ADR-0001"
---
# ADR-451679d5-473b-4faf-bf2d-a6f6d3e1e131: src の型で素の文字列を拒否する

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-0001, `docs/spec.md`
- 起点: manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §2（型定義）の srcType の項を書き直した

## 背景

out-of-store の配置は、明示の関数が返す marker を渡したときだけ選べる（ADR-0001）。`src` に文字列を
直接渡すと暗黙に out-of-store になる形は提供しない。この取り決めを、文書だけでなく型でも守らせるか
を決める必要があった。

`src` に渡せる値のうち、path と derivation（set）は、どちらも store のパスに解決されて同じように
配置される。

## 決定

`src` の型は store 由来の値（path・derivation・`outPath` を持つ attrset）と marker だけを受け付け、
素の文字列は評価時のエラーにする。path と set は型の上で分けない。

`src` のフィールドの型は REQ-a33a11e3-830d-4142-88ed-4c1fc35e7f74 に、`src` が取る値の種別は
REQ-99ca5381-6c53-426c-b145-7b4297c53868 に置く。

## 根拠

型で拒否すれば、文字列を渡して暗黙に out-of-store を選ぶ書き方が、評価の段で機械的に止まる。

path と set は `manifest.json` に渡る段でどちらも store パスの文字列になり、区別が消える。型で
分けても得るものが無い。

## 棄却した代替案

### `src` の型で path と set を分ける

`manifest.json` の段では区別が消えるので、分ける意味が無い。
