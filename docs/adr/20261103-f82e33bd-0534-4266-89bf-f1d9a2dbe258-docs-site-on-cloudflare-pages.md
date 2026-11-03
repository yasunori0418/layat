---
id: "ADR-f82e33bd-0534-4266-89bf-f1d9a2dbe258"
type: adr
name: "ドキュメントサイトを GitHub Actions から Cloudflare Pages へ配備する"
status: 採用
issues:
  - "#57"
origin: "Issue #57 のサイト設計。ADR-0037 §2 を書き直した"
justifies:
  - "INF-0865477b-21b2-4feb-9c06-1a882a4e595a"
references:
  - "ADR-993423c8-6374-4b72-aa2d-5d5fdbe7a3e4"
  - "ADR-da7b74a6-bab4-43ea-8f1c-cf843b8b58cc"
---
# ADR-f82e33bd-0534-4266-89bf-f1d9a2dbe258: ドキュメントサイトを GitHub Actions から Cloudflare Pages へ配備する

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-993423c8-6374-4b72-aa2d-5d5fdbe7a3e4, ADR-da7b74a6-bab4-43ea-8f1c-cf843b8b58cc, GitHub Issue #57
- 起点: Issue #57 のサイト設計。ADR-0037 §2 を書き直した

## 背景

ドキュメントサイトのソースは同じリポジトリの `site/` に置く（ADR-993423c8-6374-4b72-aa2d-5d5fdbe7a3e4）。
サイトのビルドでは nixdoc と gomarkdoc を nix から動かしてリファレンスを生成する
（ADR-da7b74a6-bab4-43ea-8f1c-cf843b8b58cc）。これをどこでホストし、どこから配備するかを決める必要があった。

## 決定

ドキュメントサイトは Cloudflare Pages でホストする。配備は Cloudflare の Git 連携ビルドに任せず、
GitHub Actions から wrangler で行う。あわせて、CI でサイトのビルドを検証する。

本番と preview の打ち分けや workflow の分け方は INF-0865477b-21b2-4feb-9c06-1a882a4e595a に置く。

## 根拠

Cloudflare Pages は PR ごとにプレビューを配備できるので、ドキュメントのレビューを URL で回せる。

ビルドには nix が要り、Cloudflare 側のビルド環境では賄えない。そのため配備の主体は GitHub Actions に
なる。wrangler なら本番と preview の配備を同じ仕組みで扱える。CI でビルドを検証すれば、doc-comment の
構文エラーのようなリファレンス抽出の破損をマージ前に見つけられる。

## 棄却した代替案

### GitHub Pages でホストする

リポジトリの中で完結して手軽だが、PR ごとのプレビューを標準では持たない。ドキュメントを
レビューする体験で Cloudflare Pages に劣る。

### Cloudflare の Git 連携ビルドで配備する

配備の設定は簡単になるが、Cloudflare 側のビルド環境では nix を動かせない。使うには生成した
リファレンスをコミットするしかなく、ビルド時に生成する決定と両立しない。
