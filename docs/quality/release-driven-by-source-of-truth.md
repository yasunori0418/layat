---
id: "QA-0949183b-7ef0-4cae-b88f-3ad361576b63"
type: quality
name: "リリースはバージョンの一次情報の変更で駆動し、手作業の工程を挟まない"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
depends_on:
  - "QA-a5f7f088-a459-4bb2-9674-82b1a4a52053"
specification: |
  The version of layat SHALL have a single source of truth in the repository, from which
  every other place that states the version SHALL be derived rather than maintained by
  hand. Releasing SHALL be driven by a change to that source of truth reaching `main`, and
  the tag, the release notes and the GitHub Release SHALL be produced automatically from
  it; no manual tagging or hand-written changelog SHALL be part of the release path.
specification_ja: |
  layat のバージョンはリポジトリ内に唯一の一次情報を持たなければならず、バージョンを述べる
  他のすべての箇所はそこから導出されなければならない（手で二重管理してはならない）。
  リリースはその一次情報の変更が main へ到達することで駆動されなければならず、タグ・
  リリースノート・GitHub Release はそこから自動的に生成されなければならない。手動での
  タグ打ちや手書きの CHANGELOG をリリースの処理に含めてはならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:03+09:00"
---
# QA-0949183b-7ef0-4cae-b88f-3ad361576b63: リリースはバージョンの一次情報の変更で駆動し、手作業の工程を挟まない

## 仕様

バージョンは flake の `version` と Go バイナリに埋め込む文字列の両方が述べるが、それぞれを
手で書き換える運用は片方だけが古くなる。一次情報を 1 箇所に置き、他はそこから導出する。

リリースの駆動も同じ考え方で、**一次情報の変更が main へ入ること**がリリースそのものになる。
人が別途タグを打つ工程を挟むと「マージしたがタグを忘れた」状態が生じ、CHANGELOG を手書き
すればコミット履歴との二重管理になる。

リリースを駆動する変更も main への変更である以上、[マージ前の自動検証を必須にし、プラットフォーム差が出る層は全プラットフォームで通す](cross-platform-verification-before-merge.md) が定めるマージの条件の対象になる。
「main 直接コミット禁止」と「リリースの自動化」がここで整合する。この規範は [マージ前の自動検証を必須にし、プラットフォーム差が出る層は全プラットフォームで通す](cross-platform-verification-before-merge.md) の
担当なので本 item の specification では繰り返さず、`depends_on` で依存だけを示す。

## 出典

[VERSION ファイルをバージョンの一次情報にする](../adr/version-file-single-source.md)、[bump-version.yml でバージョンを上げるコミットを push する](../adr/bump-version-workflow.md)、
[VERSION の変更をきっかけに release を自動で作る](../adr/release-on-version-change.md)、[release に成果物を添付しない](../adr/release-without-artifacts.md) が置いた方針を、
基盤（INF）から分離して
quality item として立てたもの。一次情報のファイル名・workflow の分割・成果物を添付しない
判断は [リリース自動化（VERSION ファイル起点の bump PR・自動タグ・自動リリースノート）](../infrastructure/release-automation.md) が持つ。
