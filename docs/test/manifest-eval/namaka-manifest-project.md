---
id: "CASE-7fb33044-b68c-4518-bd93-b258072235ee"
type: test_case
name: "namaka: manifest-project — manifest 文書全体のスナップショット回帰"
status: stable
resource: "tests/namaka/manifest-project/"
covers:
  - "TC-de6514e2-9105-45a6-a5b9-d474911a401b"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:28:26+09:00"
---
# CASE-7fb33044-b68c-4518-bd93-b258072235ee: namaka manifest-project

## 対象

`tests/namaka/manifest-project/expr.nix`（スナップショット実体は
`tests/namaka/_snapshots/manifest-project`）

## 検証内容

`normalizeManifest` の出力全体を project root で 1 件スナップショットに固定し、以後の差分を
回帰として検出する。

- 入力は 2 entry。1 つは `subpath` を明示（`skills/nix`）、もう 1 つは `subpath` / `target` /
  `method` を全て省略して既定適用を通す。既定と明示の双方が文書へ現れる状態を固定する
- 配置元は [評価テストの store-backed な入力は固定 outPath を持つ fake flake-input で与える](../../test-plan/eval-test-double.md) に従い、2 つは store hash を
  違える（`000…-fake-src` と `111…-other`）
- どの不変条件も名指ししていないフィールドの追加・削除・改名が、ここで差分として現れる

スナップショットに委ねてよい範囲をどこまでとするかは [manifest 文書全体をスナップショットに固定し、どの不変条件も名指ししない変更を差分として出す](manifest-document-regression.md) が持つ。

## 出典

`tests/namaka/manifest-project/expr.nix` の現行実装からの逆算（→ Issue #273「L1〜L4」節）。
namaka を併用する設計判断は [テストを sandbox 可否で flake check と E2E に分ける](../../adr/test-strategy-layers.md) / [lib は nix-unit の評価テストと namaka のスナップショットの 2 手段で検証する](../../design/lib-test-nix-unit-namaka.md) が持つ。
