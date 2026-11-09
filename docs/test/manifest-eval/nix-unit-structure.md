---
id: "CASE-7a6c4957-9365-4490-b48f-9725f42162f2"
type: test_case
name: "nix-unit: structure.nix — manifest 構造の不変条件"
status: stable
resource: "tests/nix-unit/structure.nix"
covers:
  - "TC-4e7cfae7-72bc-4af6-a1f5-1ead7db564b1"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:28:26+09:00"
---
# CASE-7a6c4957-9365-4490-b48f-9725f42162f2: nix-unit structure.nix

## 対象

`tests/nix-unit/structure.nix`（`tests/nix-unit.nix` のディレクトリ列挙で自動搭載。[評価テストは nix-unit で不変条件を、namaka で manifest 全体のスナップショットを見る](../../test-plan/nix-unit-namaka-split.md)
が定めるとおり、テスト名はファイル横断で一意でなければならない）

## 検証内容

`normalizeManifest` の出力構造（と `mkManifest` の passthru `targets`）を、値を固定してアサートする。

- `schemaVersion` が 1
- `root.rootKind` が `"project"`
- `systemRoot` を与えた `root.rootKind` が `"system"`（評価時に拒否されず正規の値として通る）
- project root では固定 root パス（`root.root`）を持たない
- store-backed entry が `{ srcKind = "store"; src; subpath; target; method }` に exact 一致
  （余分なキーが残れば落ちる）
- out-of-store marker を与えた entry が `srcKind = "outOfStore"` と marker の絶対パスへ
  変換され、`method` は既定の `"symlink"`
- out-of-store entry に判別タグ `_layatMarker` が残っていないことを単独でも明示アサート
- `mkManifest` の passthru `targets` が正規化後 target を属性キーの辞書順で返し、明示 `target`
  上書きが値に反映されること（fake pkgs で passthru だけを純評価で取り出す。→ [config 間の target 衝突は CLI が判定し engine に持ち込まない](../../adr/cross-config-conflict-judged-by-cli.md)）
- `identity` を与えなければトップレベルに `identity` キーが無く、与えればその値が載り、
  空文字では `builtins.tryEval` が失敗する

本 CASE が検証する root 種別は project と system の `rootKind` で、文書の形の不変条件は project root で
見る（→ [manifest 文書の形（schemaVersion / root / entry 5 フィールド）を不変条件としてアサートする](manifest-shape-invariants.md)）。system の `rootKind` 変換は [manifest 文書の形（schemaVersion / root / entry 5 フィールド）を不変条件としてアサートする](manifest-shape-invariants.md) の適用範囲外で、
これを受ける TC はまだ無い。fixed root の絶対パス併記と
`homeRoot` marker の評価層での扱いは [nix-unit: fixed-root.nix — fixed root の絶対パス併記](nix-unit-fixed-root.md)（→ [fixed root で rootKind が fixed になり絶対パスを併記する（home marker が倒れないことを含む）](fixed-root-absolute-path.md)）、HM 統合での
`homeRoot` は `checks.hm-module`（`integration` 区分）が担当する。
配置元は [評価テストの store-backed な入力は固定 outPath を持つ fake flake-input で与える](../../test-plan/eval-test-double.md) に従う。

## 出典

`tests/nix-unit/structure.nix` の現行実装からの逆算（→ Issue #273「L1〜L4」節）。manifest
構造の設計判断は [manifest.json を schemaVersion 1 の唯一の契約として固定する](../../adr/manifest-json-contract-schemaversion-v1.md) / [manifest.json へ marker を `clean` enum で写す](../../adr/wire-format-clean-enum.md) /
[manifest.json の entry は target を識別子にして name を持たない](../../adr/manifest-entry-identity-is-target.md) が持つ。
