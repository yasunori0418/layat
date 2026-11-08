---
id: "CASE-59de34d4-d9d5-4ae0-b1be-44d2e775f031"
type: test_case
name: "nix-unit: defaults.nix — 既定適用・明示上書き・target 辞書順"
status: stable
resource: "tests/nix-unit/defaults.nix"
covers:
  - "TC-d9175bb5-d7ec-41e0-8bee-71de928a71fb"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:28:26+09:00"
---
# CASE-59de34d4-d9d5-4ae0-b1be-44d2e775f031: nix-unit defaults.nix

## 対象

`tests/nix-unit/defaults.nix`

## 検証内容

`normalizeManifest` の正規化を 3 本でアサートする。

- **既定適用**: `src` だけを宣言した entry が `subpath = "."` / `target = 属性キー` /
  `method = "symlink"` を得ること（出力 entry 全体の exact 一致で見る）
- **明示上書き**: `target` / `subpath` / `method` を明示した entry で、既定値ではなく宣言値が
  文書へ載ること
- **決定的な配列化**: 属性キー `b` / `a` / `c` を与えたとき、出力 entries の `target` 列が
  `[a b c]` の辞書順になること

配置元は [評価テストの store-backed な入力は固定 outPath を持つ fake flake-input で与える](../../test-plan/eval-test-double.md) に従う。

## 出典

`tests/nix-unit/defaults.nix` の現行実装からの逆算（→ Issue #273「L1〜L4」節）。既定と
辞書順の設計判断は [入力の確認を normalizeManifest に切り出す](../../adr/normalize-then-mk-two-stage.md) / [entries を target をキーにした attrset にする](../../adr/entries-keyed-by-target.md) /
[apply --all の集約は config 名の辞書順に並べる](../../adr/apply-all-lexicographic-order.md) が持つ。
