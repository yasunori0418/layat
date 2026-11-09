---
id: "QA-0d42104c-a3c2-4bd1-9cad-a42d9e5a56a1"
type: quality
name: "リファレンスドキュメントは記述対象のソースから生成し、生成物を持たない"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
specification: |
  Reference documentation SHALL be generated from the source it documents at build time and
  SHALL NOT be committed in generated form, so that it cannot drift from that source. Its
  buildability SHALL be verified in CI rather than on request, so that a break in the
  extraction is caught before it is merged.
specification_ja: |
  リファレンスドキュメントはビルド時に記述対象のソースから生成されなければならず、生成物を
  コミットしてはならない（ソースとの乖離が起こり得ないようにするため）。そのビルド可能性は、
  要求されたときではなく CI で検証されなければならない（抽出の破損をマージ前に捕まえる
  ため）。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:22:37+09:00"
---
# QA-0d42104c-a3c2-4bd1-9cad-a42d9e5a56a1: リファレンスドキュメントは記述対象のソースから生成し、生成物を持たない

## 仕様

記述対象を変えたのに記述を直し忘れる、というずれ方は「気づいた人が直す」運用では検出できない。
ソースの doc-comment から毎回生成し、生成物をコミットしないことで、乖離が起こり得る状態
そのものを作らない。

生成できなくなる形（doc-comment の構文エラーなど）は残るため、ビルド可能性の検証を CI に
置いてマージ前に捕まえる。

抽出元・抽出ツール・掲載範囲・サイトの構成は本 item の規範に含めない。[ドキュメントサイト（Astro Starlight + Cloudflare Pages・ビルド時リファレンス生成）](../infrastructure/docs-site.md) が持つ。

## 出典

[ドキュメントサイトのソースを同じリポジトリの site/ に置く](../adr/docs-site-in-same-repository.md)、[ドキュメントサイトを GitHub Actions から Cloudflare Pages へ配備する](../adr/docs-site-on-cloudflare-pages.md)、
[リファレンスをビルド時に生成する](../adr/reference-generated-at-build-time.md)、[ドキュメントサイトは英語をルートのロケールにする](../adr/docs-site-english-root-locale.md) が
置いた方針を、基盤（INF）から分離して quality item として立てたもの。当初は
[ドキュメントグラフのトレーサビリティは機械的に検証する](document-graph-mechanically-verified.md) と 1 件にまとめていたが、根拠 ADR も支える基盤も
別で片方だけ改訂され得るため、Issue #242 のレビューで 2 件へ分けた。
