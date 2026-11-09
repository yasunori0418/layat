---
id: "ADR-1e010fa2-a3d9-42b5-9641-d68d9db9f405"
type: adr
name: "ドキュメントサイトは英語をルートのロケールにする"
status: stable
sources:
  - id: "issue-57"
    resource: "https://github.com/yasunori0418/layat/issues/57"
    title: "#57"
origin: "Issue #57 のサイト設計。ADR-0037 §4 を書き直した"
justifies:
  - "INF-0865477b-21b2-4feb-9c06-1a882a4e595a"
references:
  - "ADR-993423c8-6374-4b72-aa2d-5d5fdbe7a3e4"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T02:02:56+09:00"
---
# ドキュメントサイトは英語をルートのロケールにする

- ステータス: stable
- 日付: 2026-11-03
- 関連: [ドキュメントサイトのソースを同じリポジトリの site/ に置く](docs-site-in-same-repository.md), GitHub Issue #57, `docs/glossary.md`
- 起点: Issue #57 のサイト設計。ADR-0037 §4 を書き直した

## 背景

ドキュメントサイトは英語と日本語の両方で出す（[ドキュメントサイトのソースを同じリポジトリの site/ に置く](docs-site-in-same-repository.md)）。
Starlight の i18n で、どちらの言語をルートに置くかを決める必要があった。

## 決定

Starlight の i18n では英語をルートのロケール（`/`）にし、日本語を `/ja/` に置く。訳していない
ページは英語にフォールバックして表示する。

用語の準拠先は [ドキュメントサイト（Astro Starlight + Cloudflare Pages・ビルド時リファレンス生成）](../infrastructure/docs-site.md) に置く。

## 根拠

README の正式な版を英語とした公開方針と揃い、外部の利用者が最初に触れる言語が英語になる。作者の
主な言語である日本語は `/ja/` で出す。英語へのフォールバックがあれば、日本語訳が
遅れてもリンク切れにはならない。

## 棄却した代替案

### 日本語をルートにし、英語を `/en/` に置く

英語を外向きの一次言語とする公開方針と逆になる。
