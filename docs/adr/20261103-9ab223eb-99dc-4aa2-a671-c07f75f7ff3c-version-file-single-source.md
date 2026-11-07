---
id: "ADR-9ab223eb-99dc-4aa2-a671-c07f75f7ff3c"
type: adr
name: "VERSION ファイルをバージョンの一次情報にする"
status: 採用
origin: "次期マイルストーン追加計画の grilling（2026-07-04）。参照実装: https://github.com/yasunori0418/cryoflow のリリース自動化（bump workflow と release workflow）。ADR-0042 §1 を書き直した"
justifies:
  - "INF-9878e9f5-1ec0-4ef2-879b-859ea538cc43"
  - "REQ-9ed6b500-a11f-414e-a763-adb47c89f7d4"
  - "REQ-a5053191-1c6a-449b-9c5e-5ff49dc5aead"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T17:00:19+09:00"
---
# ADR-9ab223eb-99dc-4aa2-a671-c07f75f7ff3c: VERSION ファイルをバージョンの一次情報にする

- ステータス: 採用
- 日付: 2026-11-03
- 起点: 次期マイルストーン追加計画の grilling（2026-07-04）。参照実装: https://github.com/yasunori0418/cryoflow のリリース自動化（bump workflow と release workflow）。ADR-0042 §1 を書き直した

## 背景

layat にはバージョン付きのリリースが無く、利用者は flake ref のブランチや rev で pin するしか
なかった。リリースを作るには、バージョンをどこに置くかを決める必要がある。layat には、Python の
pyproject.toml のようにバージョンを書く場所が言語の側に用意されていない。

## 決定

バージョンの一次情報は、リポジトリ直下の `VERSION` ファイル（semver のプレーンテキスト 1 行）に
置く。`flake.nix` は `builtins.readFile` でこれを読んで `packages.layat` の `version` に反映し、
Go のバイナリへは nix build の `ldflags` で `main.version` に埋め込む。`layat --version` と
`--json` の `tool.version` は、この埋め込んだ値を出す。

`--version` の書式は REQ-9ed6b500-a11f-414e-a763-adb47c89f7d4、`tool.version` は
REQ-a5053191-1c6a-449b-9c5e-5ff49dc5aead、ファイルと workflow の構成は
INF-9878e9f5-1ec0-4ef2-879b-859ea538cc43 に置く。

## 根拠

1 つの一次情報から flake とバイナリの両方のバージョンを導くので、二重に管理しない。独立した
ファイルなら言語に依存せず、flake・Go・CI のどこからでも読める。リリースの workflow が発火の
判定に使う `paths` フィルタも、1 つのファイルを指すだけで済む。

## 棄却した代替案

### `flake.nix` の中にバージョンを直書きする

`paths` フィルタが `flake.nix` の全ての変更に反応し、リリースを意図しない変更でも発火の判定が
複雑になる。
