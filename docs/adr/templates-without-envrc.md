---
id: "ADR-653e6a28-76dd-44b3-99e2-3be189a3c1c4"
type: adr
name: "template に .envrc を同梱しない"
section: ci-release-docs
status: stable
origin: "ADR-0018 §3c を書き直した"
justifies:
  - "REQ-196ddabf-6569-4303-942e-050872972501"
references:
  - "ADR-baf6346d-783b-4607-ab1c-c2c93c09e064"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:30+09:00"
---
# template に .envrc を同梱しない

- ステータス: stable
- 日付: 2026-11-03
- 関連: [--all は config ごとの SubjectResult を積む](all-as-multiple-subject-results.md)
- 起点: ADR-0018 §3c を書き直した

## 背景

project template の devShell は、direnv から読み込むとシェルに入る手間が省ける。direnv を使うための
`.envrc` を template に入れるかどうかは決まっていなかった。

## 決定

template に `.envrc` を同梱しない。direnv を使う場合の手順は `flake.nix` のコメントで案内する。

project template のファイル構成は [template は動く example を 1 config だけ置きバリエーションはコメントで示す](../requirements/init-template-contents.md) に置く。

## 根拠

direnv を採るかどうかは利用者の環境の選択で、使わない利用者には不要なファイルになる。direnv を
使う利用者にとっても、`.envrc` は `direnv allow` で許可するまで働かない。手順をコメントで示せば
足りる。

## 棄却した代替案

### `.envrc` を同梱する

direnv の利用者はすぐ使えるが、使わない利用者に不要なファイルを強いる。
