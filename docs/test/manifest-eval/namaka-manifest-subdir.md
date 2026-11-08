---
id: "CASE-9e44bd6d-f6b4-4d12-990e-349a4b5c0dda"
type: test_case
name: "namaka: manifest-subdir — subdir 適用形で生成した manifest 文書のスナップショット回帰"
status: stable
resource: "tests/namaka/manifest-subdir/"
covers:
  - "TC-7d713505-5673-4e15-99bf-e98965e18892"
  - "TC-de6514e2-9105-45a6-a5b9-d474911a401b"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:49+09:00"
---
# CASE-9e44bd6d-f6b4-4d12-990e-349a4b5c0dda: namaka manifest-subdir

## 対象

`tests/namaka/manifest-subdir/expr.nix`（スナップショット実体は
`tests/namaka/_snapshots/manifest-subdir`）

## 検証内容

`root = homeRoot { subdir = ".config"; }` で生成した `normalizeManifest` の出力全体を
1 件スナップショットに固定し、以後の差分を回帰として検出する。

- 入力は 2 entry。1 つは `subpath` を明示（`nvim`）、もう 1 つは `subpath` / `target` /
  `method` を全て省略して既定適用を通す
- 両 entry の target が `.config/` 前置済みで、`root` が `{"rootKind":"home"}` だけである状態を
  固定する。`subdir` が文書のどこかへ現れれば差分になる
- 配置元は [評価テストの store-backed な入力は固定 outPath を持つ fake flake-input で与える](../../test-plan/eval-test-double.md) の fake flake-input double イディオムに従い、2 つは store hash を
  違える（`000…-fake-src` と `111…-other`）

スナップショットに委ねてよい範囲をどこまでとするかは [manifest 文書全体をスナップショットに固定し、どの不変条件も名指ししない変更を差分として出す](manifest-document-regression.md) が持つ。前置と root の不変は
[subdir 適用形が全 target に正規化済みの前置を施し、root と単体 marker を変えず、不正な subdir を評価時に拒否する](root-subdir-sugar.md) の名前付きアサート（[nix-unit: root-subdir.nix — root マーカーの subdir 適用形](nix-unit-root-subdir.md)）でも守る。

## 出典

`tests/namaka/manifest-subdir/expr.nix` の実装。規範は [root マーカーの subdir 適用形は全 target への前置と等価な評価時の糖衣で、root を動かさない](../../requirements/root-marker-subdir-sugar.md)。
