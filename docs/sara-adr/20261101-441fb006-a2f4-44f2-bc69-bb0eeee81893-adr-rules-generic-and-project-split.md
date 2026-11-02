---
id: "SDR-441fb006-a2f4-44f2-bc69-bb0eeee81893"
type: sara_adr
name: "ADR の規則を汎用とプロジェクト当てはめに分けて書く"
status: 採用
issues:
  - "#499"
  - "#500"
origin: "Issue #499 の grilling（2026-11-01）で確定"
justifies:
  - "QA-030f926b-5ae7-4543-b8cb-e66aae5e2b5e"
---
# SDR-441fb006-a2f4-44f2-bc69-bb0eeee81893: ADR の規則を汎用とプロジェクト当てはめに分けて書く

- ステータス: 採用
- 日付: 2026-11-01
- 関連: GitHub Issue #499, #500, `docs/agents/adr-rules.md`, `docs/adr/README.md`
- 起点: Issue #499 の grilling（2026-11-01）で確定

## 背景

ADR の規則は `docs/adr/README.md` に、layat の ADR 番号・型の区分・実データと混ぜて書かれていた。
#500 で sara 品質管理機構を別のリポジトリへ切り出すと、規則だけを持ち出すことになる。同じ文書にあるままでは
layat 固有の記述を 1 文ずつ取り除く作業が要る。

## 決定

ADR の規則を 2 つの文書に分けて書く。

- 汎用の規則は `docs/agents/adr-rules.md` に置く。ADR 番号・型の区分・ディレクトリ名を含めない
- layat への当てはめ（型とディレクトリ・frontmatter の対応表・検算用の実データ・採番の手順）は
  `docs/adr/README.md` に置く

## 根拠

汎用の規則に layat 固有の語が無ければ、文書をそのまま切り出し先へ移せる。当てはめの側は layat に
残り、実データの更新は規則の文書に波及しない。

## 棄却した代替案

### 1 つの文書の中で節を分ける

書き足すうちに、どの節に何を書くかの区別が曖昧になる。layat 固有の語が汎用の節に書かれても、文書が 1 つでは気付きにくい。
