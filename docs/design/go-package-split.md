---
id: "DSG-7d354fe0-a333-495b-9f4b-14bba316dc47"
type: design
name: "Go 側は cmd/layat に CLI 面を、internal に配置ロジックを置く 2 パッケージ構成にする"
status: stable
satisfies:
  - "REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645"
  - "REQ-b74a118a-1272-44eb-944c-7725163211c6"
  - "REQ-637599dc-a1ec-4af5-9e97-e882c7df56d0"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:36:11+09:00"
---
# DSG-7d354fe0-a333-495b-9f4b-14bba316dc47: Go 側は cmd/layat に CLI 面を、internal に配置ロジックを置く 2 パッケージ構成にする

## 設計

```
cmd/layat/
└── main.go            # entrypoint 発見 + nix build/eval オーケストレーション + サブコマンド分岐
internal/              # 配置・diff・保守的 stale 除去の純ロジック（ユニットテスト対象）
```

[layat は CLI とエンジンの 2 層で構成する](../requirements/two-layer-architecture.md) が定める CLI と engine の 2 層を、そのまま Go のパッケージ境界へ落とす。
`packages.layat` は両者（`cmd/layat` + `internal`）をまとめてビルドした 1 バイナリになる。

分け方が実現手段になるのは次の点。

- **依存の非対称を層ごとに固定できる**。[CLI が持ち込む依存は許可した第三者ライブラリと pin した Go に限り、いずれも固定する](../requirements/cli-dependency-policy.md) が CLI 層に許す第三者依存（cobra）は
  `cmd/layat` に閉じ、[engine は第三者依存ゼロの stdlib-only とし内部層に閉じる](../requirements/engine-stdlib-only.md) が engine へ課す stdlib-only は `internal/` に閉じる。
  1 パッケージだと同一の import ブロックに両方が含まれ、どちらの制約が及ぶ範囲かが
  コードから読めなくなる
- **`internal/` は Go の言語仕様で外部リポジトリから import できない**ため、engine を
  内部層へ閉じる [engine は第三者依存ゼロの stdlib-only とし内部層に閉じる](../requirements/engine-stdlib-only.md) がコンパイラに強制される。engine をライブラリとして
  公開しない意思が構造で表明される
- **テスト対象が `internal/` に集約される**。engine の振る舞い（[engine が叩く外部コマンドは git のみに限る](../requirements/engine-external-command-constraint.md) が許す
  外部コマンドの範囲を含む）を検証する対象が 1 パッケージに収まるため、engine 側の
  ユニットテスト（[engine は nix を使わない tmpdir 統合テストで検証し、古くなった配置の除去の安全不変条件を table-driven に重点配分する](engine-test-go-unit-tmpdir.md)）が engine の全処理を実行できる。CLI 層と混在していると、
  検証対象と cobra を含む CLI の関心事が同じパッケージに同居する

`main.go` 単一ファイルにサブコマンド分岐まで置くのは、CLI 層が持つのが「entrypoint 発見・
nix オーケストレーション・フラグ解釈」に限られ、分岐の実体が `internal/` の呼び分けに
過ぎないため。

## 出典

`docs/design.md`「プロジェクト構成」の `cmd/layat/` / `internal/` の項と、
同節末尾の「配置ロジックは Go エンジンが単一の源として所有し、`lib/` はデータ生成に
徹する」。
