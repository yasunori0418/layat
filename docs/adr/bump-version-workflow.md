---
id: "ADR-0dd4036d-594b-4f36-a509-e5498b231cf7"
type: adr
name: "bump-version.yml でバージョンを上げるコミットを push する"
section: ci-release-docs
status: stable
origin: "次期マイルストーン追加計画の grilling（2026-07-04）。参照実装: https://github.com/yasunori0418/cryoflow のリリース自動化（bump workflow と release workflow）。ADR-0042 §2 を書き直した"
justifies:
  - "INF-9878e9f5-1ec0-4ef2-879b-859ea538cc43"
references:
  - "ADR-9ab223eb-99dc-4aa2-a671-c07f75f7ff3c"
  - "ADR-8035e7c6-7dc5-4aa8-9a37-403998fa7f9b"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:26:59+09:00"
---
# bump-version.yml でバージョンを上げるコミットを push する

- ステータス: stable
- 日付: 2026-11-03
- 関連: [VERSION ファイルをバージョンの一次情報にする](version-file-single-source.md), [main への直接 push を ruleset で禁じる](no-direct-push-to-main.md)
- 起点: 次期マイルストーン追加計画の grilling（2026-07-04）。参照実装: https://github.com/yasunori0418/cryoflow のリリース自動化（bump workflow と release workflow）。ADR-0042 §2 を書き直した

## 背景

[VERSION ファイルをバージョンの一次情報にする](version-file-single-source.md)。
[main への直接 push を ruleset で禁じる](no-direct-push-to-main.md)。そのため、
`VERSION` の書き換えも PR を通して main へ入れる必要がある。

## 決定

`workflow_dispatch` で起動する `bump-version.yml` を置く。この workflow はバージョン文字列を
入力に取り、main 以外のブランチの上で `VERSION` を書き換えてコミットし、push する。PR は人が開き、
マージは通常の PR の流れで行う。手元で `VERSION` を編集して PR を作る手動の方法も同じく有効で、
workflow は省力化のためのもので唯一の方法ではない。

workflow の構成は [リリース自動化（VERSION ファイル起点の bump PR・自動タグ・自動リリースノート）](../infrastructure/release-automation.md) に置く。

## 根拠

バージョンを上げる変更が PR を通るので、リリースの意思決定も、既存のマージの条件と
レビューを通る。bump のコミットは workflow がブランチの上に積み、マージは人が PR で行うので、main
への直接の push を禁じる ruleset と矛盾しない。

## 棄却した代替案

### main の上で直接 bump のコミットを積む

main への直接の push を禁じる ruleset に反し、テストの成功というマージの条件も通らない。
