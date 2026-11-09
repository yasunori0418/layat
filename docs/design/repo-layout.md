---
id: "DSG-1361df1a-31c4-46ed-a6ce-92938f94ac02"
type: design
name: "リポジトリを lib / cmd / internal / templates / modules のトップレベル 5 ディレクトリへ分ける"
status: stable
satisfies:
  - "REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645"
  - "REQ-d85f0cef-0f1e-4897-a841-41b61a8dae51"
  - "REQ-b74a118a-1272-44eb-944c-7725163211c6"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:36:11+09:00"
---
# リポジトリを lib / cmd / internal / templates / modules のトップレベル 5 ディレクトリへ分ける

## 設計

`flake.nix` を唯一の entrypoint とし、その下にトップレベル 5 ディレクトリを並置する。

```
<project>/
├── flake.nix              # outputs を定義（packages.layat / lib / templates / modules）
├── flake.lock             # flake 入力のロック
├── lib/                   # 純データ生成（nixpkgs.lib のみ依存）
├── cmd/layat/              # layat CLI のエントリポイント（packages.layat）
├── internal/              # 配置・diff・保守的 stale 除去の純ロジック（Go ライブラリ）
├── templates/             # nix flake init -t 用
└── modules/               # 各モジュールシステムへの統合層
```

各ディレクトリの内訳は個別の `design` item が持つ（`lib/` は [lib は公開 API・型定義・manifest 生成・マーカー構築子・apply コマンド組み立ての 5 ファイルへ分割する](lib-module-split.md)、`cmd/layat` と
`internal/` は [Go 側は cmd/layat に CLI 面を、internal に配置ロジックを置く 2 パッケージ構成にする](go-package-split.md)、`modules/` は [modules は common.nix に共通オプションを集約し、統合層ごとに 1 ファイルへ分ける](modules-file-split.md)）。

この分け方が満たしているのは、[layat は CLI とエンジンの 2 層で構成する](../requirements/two-layer-architecture.md) が要求する CLI と engine の 2 層構成、
[lib は nixpkgs.lib のみに依存する純データ生成器である](../requirements/lib-pure-data-generator.md) が要求する lib の依存制約、[engine は第三者依存ゼロの stdlib-only とし内部層に閉じる](../requirements/engine-stdlib-only.md) が要求する engine の内部層への
封じ込めの 3 つで、それぞれが**別ディレクトリとして物理的に分かれている**ことが
実現手段になる。

- `lib/`（Nix）と `internal/`（Go）が別ツリーなので、[lib は nixpkgs.lib のみに依存する純データ生成器である](../requirements/lib-pure-data-generator.md) の「lib は配置
  ロジックを持たない」が構造として保たれる。lib から engine を呼ぶ処理がそもそも無い
- `internal/` は Go の言語機能で外部モジュールから import できないため、[engine は第三者依存ゼロの stdlib-only とし内部層に閉じる](../requirements/engine-stdlib-only.md) の
  「engine を内部層に閉じる」がコンパイラに強制される。公開 API 面は `cmd/layat` のみ
- `cmd/layat` と `internal/` を分けることで、[layat は CLI とエンジンの 2 層で構成する](../requirements/two-layer-architecture.md) の 2 層境界（`manifest.json`）が
  パッケージ境界と一致する

## 出典

`docs/design.md`「プロジェクト構成」のツリー。
