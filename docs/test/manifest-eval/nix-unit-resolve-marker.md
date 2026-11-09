---
id: "CASE-0c9d41d1-33fd-4982-aed9-fa575a56ee01"
type: test_case
name: "nix-unit: resolve-marker.nix — src 解決と marker 判別述語"
status: stable
resource: "tests/nix-unit/resolve-marker.nix"
covers:
  - "TC-81be084d-709f-481b-9b61-5d2d11c317a0"
follows:
  - "TP-36e90d5d-4524-4294-bc72-ee263bb02782"
  - "TP-d3d06fe4-6940-4df8-b111-bb4096d5444f"
  - "TP-403c55c7-d996-4951-8e6b-c3a7dddd387c"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T10:25:27+09:00"
---
# nix-unit resolve-marker.nix

## 対象

`tests/nix-unit/resolve-marker.nix`（[lib.__internal で private helper をテストへ公開する](../../test-plan/lib-internal-test-seam.md) のテスト用の公開点 `layat.__internal.resolveEntry`
と `lib/types.nix` の `isRootMarker` / `isOutOfStoreMarker` / `isRootRelativeMarker` を直接叩く。manifest 全体を介さず
単一 entry の src 種別判定・文字列化だけを見る最小の境界。root 相対 marker だけは、
`normalizeManifest` を通した entries の形も 1 件確かめる）

## 検証内容

src 解決については、`srcKind` 単体・`src` 単体・entry 全体の shape という 3 段のアサートを
store / out-of-store / root 相対それぞれに重ねて置く。どの段で落ちたかで、種別判定・文字列化・余分な
キーの混入のどれが破損したかを見分けられるようにするため。

- **store src**: `srcKind = "store"`、`src` が store パスへ文字列化、entry 全体が
  5 フィールドに exact 一致
- **out-of-store marker**: `srcKind = "outOfStore"`、`src` が marker の絶対パス、entry 全体が
  exact 一致（`method = "copy"` を明示した形で）
- **root 相対 marker**: `mkRootRelativeSymlink` が `_layatMarker = "rootRelative"` と `path` だけを
  持つこと、`srcKind = "rootRelative"`、`src` が marker の root 相対パスのまま、entry 全体が
  exact 一致。`normalizeManifest` を通した entries でも同じ形に正規化されること
- **判別タグの非漏洩**: store / out-of-store / root 相対のいずれの解決結果にも `_layatMarker` が無いこと
- **`isOutOfStoreMarker`**: out-of-store marker には true、root マーカー・root 相対 marker・
  store-backed な attrset には false
- **`isRootRelativeMarker`**: root 相対 marker には true、out-of-store marker・root マーカー・
  store-backed な attrset には false
- **`isRootMarker`**: projectRoot / homeRoot / systemRoot には true、out-of-store marker と
  絶対パス文字列には false

配置元は [評価テストの store-backed な入力は固定 outPath を持つ fake flake-input で与える](../../test-plan/eval-test-double.md) に従う。

## 出典

`tests/nix-unit/resolve-marker.nix` の現行実装からの逆算（→ Issue #273「L1〜L4」節）。
marker の設計判断は [out-of-store のリンクは lib の関数が包むだけにして engine が張る](../../adr/out-of-store-linked-by-engine.md) / [root 相対マーカーを別関数 mkRootRelativeSymlink で作る](../../adr/mkrootrelativesymlink-separate-function.md) /
[marker を判別タグで derivation と見分ける](../../adr/marker-tag-typing.md) / [root は 3 つの marker で表す](../../adr/three-root-markers.md) が持つ。
