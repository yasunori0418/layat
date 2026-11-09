---
id: "TC-de6514e2-9105-45a6-a5b9-d474911a401b"
type: test_condition
name: "manifest 文書全体をスナップショットに固定し、どの不変条件も名指ししない変更を差分として出す"
status: stable
mitigates:
  - "RISK-5df2d02b-e5d4-40eb-86ad-e8bc96e4c34d"
  - "RISK-3de9753f-3fd3-4364-b1ab-64c68c15ec77"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:28:26+09:00"
---
# 文書全体をスナップショットに固定する

## テスト条件

`normalizeManifest` の出力全体をスナップショットとして保存し、以後の変更が差分として現れる
ようにする。個別アサートは名指しした性質しか守らないため、フィールドの追加・削除・改名の
ように「アサートを書き忘れた箇所」の変化はここでしか捕まらない。

- 入力は既定適用の対象（`subpath` / `target` / `method` を省略した entry）と明示指定した
  entry を混在させ、既定と明示の双方が文書へ現れる状態を固定する
- 配置元は [評価テストの store-backed な入力は固定 outPath を持つ fake flake-input で与える](../../test-plan/eval-test-double.md) に従う
- 2 つの配置元は store hash を違えて置き、entry どうしの取り違えが差分として出るようにする

**このスナップショットに委ねてよい範囲**: [評価テストは nix-unit で不変条件を、namaka で manifest 全体のスナップショットを見る](../../test-plan/nix-unit-namaka-split.md) が定めるとおり、規範として述べられて
いる性質は名前付きアサート（[manifest 文書の形（schemaVersion / root / entry 5 フィールド）を不変条件としてアサートする](manifest-shape-invariants.md) / [省略時の既定適用・明示上書き・target 辞書順での決定的配列化をアサートする](entry-normalization-defaults.md) / [marker の判別と src 種別への変換を単一 entry の境界でアサートする](marker-resolution.md)）で守り、スナップショット
だけに委ねない。スナップショットは変わったときに丸ごと再承認されるため、委ねた性質は
「変わってよかったのか」を問われないまま通る。この区別は本 item が
[評価スイート自身の盲点でアサートが空虚に緑になる](../../risks/eval-suite-blind-spot.md)（評価スイート自身の盲点）を緩和する部分そのものである。

## 覆う CASE

- [namaka: manifest-project — manifest 文書全体のスナップショット回帰](namaka-manifest-project.md)（`tests/namaka/manifest-project/expr.nix`）
- [namaka: manifest-subdir — subdir 適用形で生成した manifest 文書のスナップショット回帰](namaka-manifest-subdir.md)（`tests/namaka/manifest-subdir/expr.nix`）

## 出典

`tests/namaka/manifest-project/expr.nix` と `tests/namaka/_snapshots/` の現行実装からの逆算
（→ Issue #273「L1〜L4」節）。役割分担の規範は [評価テストは nix-unit で不変条件を、namaka で manifest 全体のスナップショットを見る](../../test-plan/nix-unit-namaka-split.md)、double の規範は [評価テストの store-backed な入力は固定 outPath を持つ fake flake-input で与える](../../test-plan/eval-test-double.md)。
