---
id: "ADR-c2f0cdc9-284e-4425-8f60-15f3aa423ed4"
type: adr
name: "release に成果物を添付しない"
section: ci-release-docs
status: stable
origin: "次期マイルストーン追加計画の grilling（2026-07-04）。参照実装: https://github.com/yasunori0418/cryoflow のリリース自動化（bump workflow と release workflow）。ADR-0042 §4 を書き直した"
justifies:
  - "INF-9878e9f5-1ec0-4ef2-879b-859ea538cc43"
references:
  - "ADR-2fcf26b4-bce6-4ed6-a374-3afe560ead21"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:26:59+09:00"
---
# release に成果物を添付しない

- ステータス: stable
- 日付: 2026-11-03
- 関連: [VERSION の変更をきっかけに release を自動で作る](release-on-version-change.md)
- 起点: 次期マイルストーン追加計画の grilling（2026-07-04）。参照実装: https://github.com/yasunori0418/cryoflow のリリース自動化（bump workflow と release workflow）。ADR-0042 §4 を書き直した

## 背景

[VERSION の変更をきっかけに release を自動で作る](release-on-version-change.md)。作るのはタグと GitHub Release である。Release にバイナリなどの成果物を添付するかを
決める必要がある。

## 決定

GitHub Release に成果物を添付しない。配布の正規の手段は、タグで pin できる flake ref とする。
需要が出たら `release.yml` にステップを足せる余地は残し、添付を禁じはしない。

成果物の扱いは [リリース自動化（VERSION ファイル起点の bump PR・自動タグ・自動リリースノート）](../infrastructure/release-automation.md) に置く。

## 根拠

layat は実行時に `nix` と `git` を呼ぶ nix 前提のツールで、nix の無い環境ではバイナリ単体では
動かない。添付しても使える場面が無く、価値が薄い。

## 棄却した代替案

### バイナリを Release に添付する

nix 前提のツールに、単体のバイナリの需要は薄い。必要になった時点で `release.yml` に追加できる。
