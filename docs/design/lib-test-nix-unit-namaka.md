---
id: "DSG-fb49e36c-eb20-4efa-8cac-88ef0873db8e"
type: design
name: "lib は nix-unit の評価テストと namaka のスナップショットの 2 手段で検証する"
section: test-strategy
status: stable
satisfies:
  - "REQ-2b0c2bb8-964f-4e36-a121-c6ea0d4be1c4"
  - "REQ-60e6b49c-9ba1-4552-a0ec-d340421ec281"
  - "REQ-b232ec98-af3b-41f3-a050-29d417322002"
  - "REQ-d1b5b3f5-10a0-400d-9f03-ba00c63d1c34"
  - "TP-403c55c7-d996-4951-8e6b-c3a7dddd387c"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:22:37+09:00"
---
# lib は nix-unit の評価テストと namaka のスナップショットの 2 手段で検証する

## 設計

| 手段 | 重点 |
|---|---|
| **nix-unit**（評価テスト）| `mkManifest` の不変条件 — 個別の入力に対して期待する出力・拒否すべき入力がエラーになること |
| **namaka**（スナップショット）| `manifest.json` 全体の回帰 — 出力構造がまるごと変わっていないこと |

2 手段を併用するのは、lib が生成するものが構造化データ 1 個であることに由来する。

- **nix-unit が担うのは「点」の検証**。[normalizeManifest が確認・デフォルト適用・marker 変換を行い mkManifest が derivation を組む](../requirements/normalize-manifest-defaults.md) と
  [mkManifest 自身が evalModules で入力を確認する単一の箇所になる](../requirements/mkmanifest-input-validation.md) の 2 要求は、正常系より拒否されるべき入力がきちんとエラーになるかが
  肝になる。評価テストなら「この入力は eval が失敗する」を直接書ける
- **namaka が担うのは「面」の検証**。返り値（[mkManifest は manifest.json と symlink farm を含む store オブジェクトを返す](../requirements/mkmanifest-return-value.md)）は、フィールドを 1 つ書き換えただけでも
  engine 側の契約が守られなくなりうる。全体をスナップショットに固定しておけば、
  意図しない構造変化が差分として出る。個別アサーションを網羅する形では
  「書き忘れたフィールド」を捕まえられない

いずれも実際のビルドを伴わない評価だけで完結するのが選択理由でもある。
[mkManifest は配置データを生成する純粋関数である](../requirements/mkmanifest-pure-function.md) という要求がある以上、検証も評価層で閉じるのが自然で、
FS も nix daemon も要らない。テスト対象の private helper へは、テスト用の公開点 `lib.__internal`（[lib.__internal で private helper をテストへ公開する](../test-plan/lib-internal-test-seam.md)）から到達する。

engine 側の検証は [engine は nix を使わない tmpdir 統合テストで検証し、古くなった配置の除去の安全不変条件を table-driven に重点配分する](engine-test-go-unit-tmpdir.md) が、実際の使い方の一気通貫は [E2E は tests/e2e の bash ハーネスに置き、シナリオの詳細は同ディレクトリの README へ委譲する](e2e-harness-shape.md) が担う。

## 出典

`docs/design.md`「テスト戦略」のテーブル 1 行目。
