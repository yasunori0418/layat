---
id: "CASE-9e44bd6d-f6b4-4d12-990e-349a4b5c0dda"
type: test_case
name: "namaka: manifest-subdir — subdir 適用形で生成した manifest 文書のスナップショット回帰"
target: "tests/namaka/manifest-subdir/"
covers:
  - "TC-7d713505-5673-4e15-99bf-e98965e18892"
  - "TC-de6514e2-9105-45a6-a5b9-d474911a401b"
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
- 配置元は TP-d3d06fe4-6940-4df8-b111-bb4096d5444f の fake flake-input double イディオムに従い、2 つは store hash を
  違える（`000…-fake-src` と `111…-other`）

スナップショットに委ねてよい範囲の線引きは TC-de6514e2-9105-45a6-a5b9-d474911a401b が持つ。前置と root の不変は
TC-7d713505-5673-4e15-99bf-e98965e18892 の名前付きアサート（CASE-9e8e3151-2c56-4c53-89ec-3e9fc5f1e06f）でも守る。

## 出典

`tests/namaka/manifest-subdir/expr.nix` の実装。規範は REQ-08c8f639-790c-4970-a954-c9567538cf9b。
