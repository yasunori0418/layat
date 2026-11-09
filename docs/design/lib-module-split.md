---
id: "DSG-e4d5db6b-5bac-4fd6-b2fb-ed07bdec30f5"
type: design
name: "lib は公開 API・型定義・manifest 生成・マーカー構築子・apply コマンド組み立ての 5 ファイルへ分割する"
status: stable
satisfies:
  - "REQ-d85f0cef-0f1e-4897-a841-41b61a8dae51"
  - "REQ-97c1e088-a17e-46d9-a9a1-83d1757d0f7d"
  - "REQ-60e6b49c-9ba1-4552-a0ec-d340421ec281"
  - "REQ-eb363122-385a-499c-a074-c95efb949d07"
  - "REQ-37b56673-6e40-4a1b-a2a7-5d3c084e3e66"
  - "TP-403c55c7-d996-4951-8e6b-c3a7dddd387c"
  - "REQ-b232ec98-af3b-41f3-a050-29d417322002"
  - "REQ-6c51facc-fa7a-4b5e-939b-44e7578d7a4f"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:22:37+09:00"
---
# lib は公開 API・型定義・manifest 生成・マーカー構築子・apply コマンド組み立ての 5 ファイルへ分割する

## 設計

```
lib/
├── default.nix        # 公開 API のまとめ（mkManifest / mkApplyCommand / mkOutOfStoreSymlink / projectRoot / homeRoot / systemRoot）
├── types.nix          # entries の型定義（各モジュールで共有）
├── manifest.nix       # mkManifest（manifest.json + symlink farm derivation を生成する純粋関数）
├── out-of-store.nix   # mkOutOfStoreSymlink / projectRoot / homeRoot / systemRoot（マーカー構築子・__functor による subdir 適用形）
└── apply-command.nix  # mkApplyCommand（link-farm の集合から apply --all --manifest の 1 行を組む）
```

分割の軸は「公開 API 面 / 型 / 生成ロジック / マーカー / コマンド組み立て」の 5 つで、次の理由による。

- **`default.nix` を公開 API のまとめに限る**ことで、[mkManifest の引数は必須の pkgs / entries / root と任意の identity とする](../requirements/mkmanifest-arguments.md)・[mkManifest は manifest.json と symlink farm を含む store オブジェクトを返す](../requirements/mkmanifest-return-value.md)・[mkOutOfStoreSymlink は out-of-store symlink を表すマーカーを返す](../requirements/mkoutofstoresymlink-marker.md)・[root は 3 マーカーと絶対パス文字列の union を取る](../requirements/root-marker-kinds.md) の 4 要求が定める API 面が 1 ファイルを
  読めば分かる。`__internal` はここから露出するが安定 API ではなく、[lib.__internal で private helper をテストへ公開する](../test-plan/lib-internal-test-seam.md) が定める、
  テストから内部へ到達するための公開点に限る
- **`types.nix` を独立させる**のは、entries の型定義を `lib` と `modules/` の双方が
  共有するため。共有できる形にしておかないと、モジュール側が型を再定義して二重管理になる
  （`modules/common.nix` からの共有は [modules は common.nix に共通オプションを集約し、統合層ごとに 1 ファイルへ分ける](modules-file-split.md)）
- **`manifest.nix` と `out-of-store.nix` を分ける**のは、前者が `normalizeManifest` と derivation 組み立て
  （[normalizeManifest が確認・デフォルト適用・marker 変換を行い mkManifest が derivation を組む](../requirements/normalize-manifest-defaults.md)）を持つのに対し、
  後者はマーカーという不活性なデータ構造を返すだけで、依存の向きが一方向（manifest → marker）に
  なるため
- **`apply-command.nix` を分ける**のは、`mkApplyCommand`（[lib.mkApplyCommand は事前展開した apply --all --manifest のコマンド行を返す](../requirements/lib-apply-command.md)）が
  manifest を作らず、出来上がった link-farm の集合から CLI のコマンド行を組むだけで、manifest の
  生成ロジックにもマーカーにも依存しないため。`modules/kick.nix` もこの関数を使う

いずれのファイルも nixpkgs.lib のみに依存し、[lib は nixpkgs.lib のみに依存する純データ生成器である](../requirements/lib-pure-data-generator.md) の依存制約をファイル単位で保つ。

## 出典

分割時点の原文（縮退前）の `docs/design.md`「プロジェクト構成」の `lib/` 内訳。

> **元の 4 ファイルの分割の根拠は ADR ではなく同原文のツリーそのもの**: その判断を決めた
> ADR は存在しない。`apply-command.nix` の追加は [lib.mkApplyCommand で事前展開した apply コマンドを組む](../adr/lib-mkapplycommand.md) が根拠で、`justifies` はこの追加分に
> 対して張られている。
