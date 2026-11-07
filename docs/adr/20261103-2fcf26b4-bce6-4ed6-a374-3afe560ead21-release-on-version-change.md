---
id: "ADR-2fcf26b4-bce6-4ed6-a374-3afe560ead21"
type: adr
name: "VERSION の変更をきっかけに release を自動で作る"
status: 採用
origin: "次期マイルストーン追加計画の grilling（2026-07-04）。参照実装: https://github.com/yasunori0418/cryoflow のリリース自動化（bump workflow と release workflow）。ADR-0042 §3 を書き直した"
justifies:
  - "INF-9878e9f5-1ec0-4ef2-879b-859ea538cc43"
references:
  - "ADR-9ab223eb-99dc-4aa2-a671-c07f75f7ff3c"
  - "ADR-0dd4036d-594b-4f36-a509-e5498b231cf7"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T17:26:22+09:00"
---
# ADR-2fcf26b4-bce6-4ed6-a374-3afe560ead21: VERSION の変更をきっかけに release を自動で作る

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-9ab223eb-99dc-4aa2-a671-c07f75f7ff3c, ADR-0dd4036d-594b-4f36-a509-e5498b231cf7
- 起点: 次期マイルストーン追加計画の grilling（2026-07-04）。参照実装: https://github.com/yasunori0418/cryoflow のリリース自動化（bump workflow と release workflow）。ADR-0042 §3 を書き直した

## 背景

バージョンは `VERSION` ファイルを一次情報にし（ADR-9ab223eb-99dc-4aa2-a671-c07f75f7ff3c）、
それを上げる変更は PR で main へ入る（ADR-0dd4036d-594b-4f36-a509-e5498b231cf7）。タグと GitHub
Release をいつ、どうやって作るかを決める必要がある。

## 決定

`release.yml` は、main への push のうち `VERSION` が変わったもので発火する。手動の再実行は
`workflow_dispatch` で行い、どちらの起動でも main だけを対象にする。`VERSION` を
読み、タグ `vX.Y.Z` と GitHub Release を作り、リリースノートは GitHub 標準の自動生成に任せる。
CHANGELOG.md はコミットしない。`release.yml` は main への push の後に実行されるので、required status
check の対象にしない。

workflow の構成は INF-9878e9f5-1ec0-4ef2-879b-859ea538cc43 に置く。

## 根拠

bump の PR のマージがリリースを駆動するので、タグの打ち忘れや、タグと内容の食い違いが構造的に
起きない。マージの条件が掛かるのは bump の PR の側で、リリースの作業はマージの後に実行される。

このリポジトリは Conventional Commits で履歴を書いているので、GitHub 標準の自動生成で十分に
読めるリリースノートになる。

## 棄却した代替案

### 手でタグを push して release を作る

構成は最小だが、タグの打ち忘れやタグの位置の誤りを人の注意力に頼ることになる。bump の PR を
起点にすれば、リリースの判断が既存のレビューを通る。

### release-please や git-cliff で CHANGELOG.md をコミットする

生成物のコミットと bot の PR の運用が増える割に、標準の自動生成ノートより得るものが少ない。
