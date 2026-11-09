---
id: "TC-81be084d-709f-481b-9b61-5d2d11c317a0"
type: test_condition
name: "marker の判別と src 種別への変換を単一 entry の境界でアサートする"
status: stable
mitigates:
  - "RISK-5df2d02b-e5d4-40eb-86ad-e8bc96e4c34d"
generated:
  by: "claude-code/claude-opus-5"
  at: "2026-09-20T08:21:50+09:00"
---
# marker の判別と src 種別への変換をアサートする

## テスト条件

manifest 全体を組み立てずに、単一 entry の解決（`resolveEntry`）と型側の marker 判別述語を
直接突いてアサートする。文書の形（[manifest 文書の形（schemaVersion / root / entry 5 フィールド）を不変条件としてアサートする](manifest-shape-invariants.md)）とは別の境界で、変換ロジックそのものを見る。

- store-backed な src が `srcKind = "store"` と文字列化された store パスへ解決されること
- out-of-store marker が `srcKind = "outOfStore"` と marker の絶対パスへ解決されること
- どちらの解決結果にも判別タグ `_layatMarker` が残らないこと（engine が読む契約は clean enum）
- `isOutOfStoreMarker` / `isRootMarker` が、root マーカー 3 種・out-of-store marker・
  store-backed な attrset・絶対パス文字列を互いに取り違えずに判別すること。marker はタグの
  値で識別され、attrset であること自体では判別しない

## 覆う CASE

- [nix-unit: resolve-marker.nix — src 解決と marker 判別述語](nix-unit-resolve-marker.md)（`tests/nix-unit/resolve-marker.nix`）

## 出典

`tests/nix-unit/resolve-marker.nix` の現行実装からの逆算（→ Issue #273「L1〜L4」節）。
