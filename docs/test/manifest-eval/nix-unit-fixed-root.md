---
id: "CASE-823a65c6-e9d2-404f-99ed-85d09376afea"
type: test_case
name: "nix-unit: fixed-root.nix — fixed root の絶対パス併記"
status: stable
resource: "tests/nix-unit/fixed-root.nix"
covers:
  - "TC-f9e927d0-8e10-4b8e-9870-5b5486949af6"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:28:26+09:00"
---
# CASE-823a65c6-e9d2-404f-99ed-85d09376afea: nix-unit fixed-root.nix

## 対象

`tests/nix-unit/fixed-root.nix`（`tests/nix-unit.nix` のディレクトリ列挙で自動搭載。[評価テストは nix-unit で不変条件を、namaka で manifest 全体のスナップショットを見る](../../test-plan/nix-unit-namaka-split.md)
が定めるとおり、テスト名はファイル横断で一意でなければならない。本ファイルは
`testFixedRoot*` 接頭辞を使う）

## 検証内容

`normalizeManifest` の `root` に marker でない絶対パス文字列を渡し、出力の `root`
オブジェクトを値を固定してアサートする。

- `root.rootKind` が `"fixed"`
- `root.root` に渡した絶対パスがそのまま写る
- `root` オブジェクトが `{ rootKind = "fixed"; root = <絶対パス>; }` に exact 一致
  （余分なキーが残れば落ちる）
- 別の絶対パスでも同じ形になる（特定の値に依存した通り方をしない）。entries は同条件に
  揃え、差分がパス値の 1 軸だけになるようにする
- 同じ entries を fixed root と project root へ与えたときの正規化結果が一致する
  （既定値そのものは [nix-unit: defaults.nix — 既定適用・明示上書き・target 辞書順](nix-unit-defaults.md) が持つので、ここは同値比較で独立性だけを見る）。
  現行実装では `resolveEntry` が root を引数に取らないため、この結合は構造的に起こり
  得ない。将来 entry の解決へ root を渡す設計変更が入ったときに初めて働く回帰網である
- `homeRoot` marker を渡した場合は `{ rootKind = "home"; }` に exact 一致し、fixed へ
  倒れず絶対パスも持たない（`root` フィールドの不在まで見る）

project root の否定側（固定 root を持たない）は [nix-unit: structure.nix — manifest 構造の不変条件](nix-unit-structure.md) が持つ。本 CASE はその対に
なる肯定側で、両者で [manifest.json の root は rootKind を持ち fixed のときだけ絶対パスを併記する](../../requirements/manifest-root-object.md) の評価層分を挟む（→ [fixed root で rootKind が fixed になり絶対パスを併記する（home marker が倒れないことを含む）](fixed-root-absolute-path.md)）。

配置元は [評価テストの store-backed な入力は固定 outPath を持つ fake flake-input で与える](../../test-plan/eval-test-double.md) に従う。

## 出典

Issue #288（[manifest.json の root は rootKind を持ち fixed のときだけ絶対パスを併記する](../../requirements/manifest-root-object.md) の肯定側が 8 区分のどこにも無い。検出元は Issue #273 L1 レーンの
逆算）。root オブジェクトの設計判断は [root の型を mkManifest だけに持たせる](../../adr/roottype-mkmanifest-only.md) が持つ。
