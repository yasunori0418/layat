---
id: "ADR-521765d7-7e31-42f4-b548-979185a7586a"
type: adr
name: "lib は純データ生成に徹する"
section: concept
status: stable
origin: "初期設計の生成 bash を実装前に正す見直し。ADR-0006「決定」節の第 2 項・ADR-0002「決定」節の「純粋性と副作用の両立」を書き直した"
justifies:
  - "REQ-d85f0cef-0f1e-4897-a841-41b61a8dae51"
  - "REQ-2b0c2bb8-964f-4e36-a121-c6ea0d4be1c4"
  - "REQ-60e6b49c-9ba1-4552-a0ec-d340421ec281"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-01T22:59:52+09:00"
---
# lib は純データ生成に徹する

- ステータス: stable
- 日付: 2026-11-01
- 関連: `docs/concept.md`
- 起点: 初期設計の生成 bash を実装前に正す見直し。ADR-0006「決定」節の第 2 項・ADR-0002「決定」節の「純粋性と副作用の両立」を書き直した

## 背景

layat は Nix の純粋な評価と、FS への配置という副作用の両方を扱う。初期の設計では、lib の純粋関数が
link farm の derivation と、配置を行う activation スクリプトの両方を生成していた。配置のロジックが
lib の生成物に入るため、lib は実質的に shell script の生成ツールになっていた。

## 決定

lib は配置データの生成だけを行う純粋関数にとどめ、配置のロジックを持たない。`lib.mkManifest` は
`manifest.json` と symlink farm を含む derivation を返し、配置はそれを入力に取る engine が行う。

lib の依存の制約は [lib は nixpkgs.lib のみに依存する純データ生成器である](../requirements/lib-pure-data-generator.md) に、`mkManifest` が純粋関数であることは
[mkManifest は配置データを生成する純粋関数である](../requirements/mkmanifest-pure-function.md) に、返す store オブジェクトの中身は
[mkManifest は manifest.json と symlink farm を含む store オブジェクトを返す](../requirements/mkmanifest-return-value.md) に置く。

## 根拠

純粋な評価と副作用を lib と engine の境界で分けられる。lib は Nix の評価テストで、engine は Go の
テストで、それぞれ検証でき、両者の間の契約は `manifest.json` の 1 点に閉じる。

## 棄却した代替案

### lib が配置スクリプトも生成する

初期の設計そのもの。配置のロジックが config ごとの生成物に散らばり、テストしにくい。
