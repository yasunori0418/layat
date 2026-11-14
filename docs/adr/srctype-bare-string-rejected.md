---
id: "ADR-451679d5-473b-4faf-bf2d-a6f6d3e1e131"
type: adr
name: "src の型で素の文字列を拒否する"
section: lib-entries
status: stable
origin: "manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §2（型定義）の srcType の項を書き直した"
justifies:
  - "REQ-a33a11e3-830d-4142-88ed-4c1fc35e7f74"
  - "REQ-99ca5381-6c53-426c-b145-7b4297c53868"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T09:15:16+09:00"
---
# src の型で素の文字列を拒否する

- ステータス: stable
- 日付: 2026-11-01
- 関連: `docs/spec.md`
- 起点: manifest の型チェックを NixOS モジュールシステムで行えるかという確認。ADR-0010 §2（型定義）の srcType の項を書き直した

## 背景

[store link を既定にし、out-of-store は明示関数でのみ選ばせる](store-link-default-out-of-store-opt-in.md)。選ぶときは明示の関数が返す marker を渡す。`src` に文字列を
直接渡すと暗黙に out-of-store になる形は提供しない。この取り決めを、文書だけでなく型でも守らせるか
を決める必要があった。

`src` に渡せる値のうち、path と derivation（set）は、どちらも store のパスに解決されて同じように
配置される。

## 決定

`src` の型は store 由来の値（path・derivation・`outPath` を持つ attrset）と marker だけを受け付け、
素の文字列は評価時のエラーにする。path と set は型の上で分けない。

`src` のフィールドの型は [entry submodule のフィールドは src / subpath / target / method の 4 つとする](../requirements/entry-submodule-fields.md) に、`src` が取る値の種別は
[src は path / set / marker の 3 種を取り store link を既定として out-of-store は marker で opt-in する](../requirements/src-kinds-store-default.md) に置く。

## 根拠

型で拒否すれば、文字列を渡して暗黙に out-of-store を選ぶ書き方が、評価の段で機械的に止まる。

path と set は `manifest.json` に渡る段でどちらも store パスの文字列になり、区別が消える。型で
分けても得るものが無い。

## 棄却した代替案

### `src` の型で path と set を分ける

`manifest.json` の段では区別が消えるので、分ける意味が無い。
