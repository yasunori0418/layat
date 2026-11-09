---
id: "TC-e7ff0e6d-32d7-4ed6-8c2f-449dba34b2f6"
type: test_condition
name: "拒否すべき入力が評価時に throw することをアサートする"
status: stable
mitigates:
  - "RISK-09df40d3-2752-433e-9ab0-2816fbd14969"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T13:39:04+09:00"
---
# 拒否すべき入力が評価時に throw することをアサートする

## テスト条件

「この入力は評価が失敗する」を、エラー種別と拒否されたフィールドの名指しで直接アサートする。
正常系のアサートでは、拒否条件が外れたことを検出できないため。判定ロジックそのものを private
helper 越しに単体で見るのは [パス脱出判定の深さ境界を private helper 越しに内側・外側の対でアサートする](path-escape-boundary.md) の担当で、検証境界が異なる。

主たる処理は `normalizeManifest`（evalModules）だが、型の共有を確かめる 1 件だけは
`entriesType` を直接 `evalModules` して見る（`normalizeManifest` を通さない）。同じ型定義が
モジュールという使い方でも strict であることは、公開された処理のアサートからは導けないため。

**条件ごとの拒否**:

- `method = "copy"` と out-of-store marker の組み合わせ（意図の矛盾）
- 絶対パスの target / `..` で root の外へ出る target / `..` で src の外へ出る subpath
- 別キーから同一 target を指す衝突
- 未知キー・旧名（strict submodule）。`modules/common.nix` が共有する entriesType でも
  同じく eval エラーになること
- 素の文字列 src（out-of-store は marker による opt-in であって暗黙変換しない）

## 覆う CASE

- [nix-unit: gates.nix — 評価時に throw する拒否条件群](nix-unit-gates.md)（`tests/nix-unit/gates.nix`）

## 出典

`tests/nix-unit/gates.nix` の現行実装からの逆算（→ Issue #273「L1〜L4」節）。
