---
id: "TC-f9e927d0-8e10-4b8e-9870-5b5486949af6"
type: test_condition
name: "fixed root で rootKind が fixed になり絶対パスを併記する（home marker が倒れないことを含む）"
status: stable
mitigates:
  - "RISK-5df2d02b-e5d4-40eb-86ad-e8bc96e4c34d"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:49+09:00"
---
# fixed root で絶対パスを併記する

## テスト条件

`normalizeManifest` の `root` に marker でない絶対パス文字列を渡したとき、出力の `root`
オブジェクトが [manifest.json の root は rootKind を持ち fixed のときだけ絶対パスを併記する](../../requirements/manifest-root-object.md) の fixed 側の規定どおりになることをアサートする。

- `root.rootKind` が `"fixed"` になること
- `root.root` に渡した絶対パスがそのまま写ること。特定の値でだけ通る実装にならないよう、
  異なるパスでも同じ形になることを確かめる
- `root` オブジェクトが `rootKind` / `root` のちょうど 2 フィールドであること
  （exact 一致で見るので余分なキーの混入は落ちる）
- root 種別を fixed にしても entry の正規化結果は変わらないこと
- root marker（`homeRoot`）は fixed へ倒れず絶対パスも持たないこと。fixed 判定が
  marker 側へ誤って広がらないことの担保で、exact 一致で見るため [manifest.json の root は rootKind を持ち fixed のときだけ絶対パスを併記する](../../requirements/manifest-root-object.md) の home 側
  （`rootKind = "home"` で `root` フィールドを持たない）の形もここが担保する

**この条件は [manifest.json の root は rootKind を持ち fixed のときだけ絶対パスを併記する](../../requirements/manifest-root-object.md) の肯定側（fixed）を持つ**。同じ REQ の否定側のうち project root
分は [manifest 文書の形（schemaVersion / root / entry 5 フィールド）を不変条件としてアサートする](manifest-shape-invariants.md) が持ち、`homeRoot` を実際の HM 統合で見る分は `checks.hm-module`
（`integration` 区分）が担当する。

射程は `normalizeManifest` が返す文書に限る。同じ絶対パスが `mkManifest` の
`passthru.root`（CLI が build 前に `nix eval` で読む処理・[mkManifest の返り値は passthru で `rootKind` を露出する](../../requirements/mkmanifest-passthru-rootkind.md)）へ写る側は
derivation 層の話で、本条件は見ない。fixed root で絶対パスを渡しても相対パスを渡しても
`rootKind = "fixed"` になる（絶対性を強制する確認が実装に無い）点も同じく射程外。
どちらも [manifest.json の形が engine の読む契約から気付かれずにずれる](../../risks/manifest-contract-drift.md) の残余として記録してあり、passthru 側の受け皿は `cli-json` 区分に
なる。

## 覆う CASE

- [nix-unit: fixed-root.nix — fixed root の絶対パス併記](nix-unit-fixed-root.md)（`tests/nix-unit/fixed-root.nix`）

## 出典

[manifest.json の root は rootKind を持ち fixed のときだけ絶対パスを併記する](../../requirements/manifest-root-object.md) の肯定側が 8 区分のどこにも無いという検出（→ Issue #288。検出元は Issue #273
L1 レーンの逆算）。[manifest.json の形が engine の読む契約から気付かれずにずれる](../../risks/manifest-contract-drift.md) が残余として記録していたものを本条件で埋める。
