---
id: "QA-a92341b9-c873-406e-8b70-a64f56d8a7d6"
type: quality
name: "コード整形と静的解析を自動検証に載せ、同じ判定を手元でも得られるようにする"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
depends_on:
  - "QA-a5f7f088-a459-4bb2-9674-82b1a4a52053"
specification: |
  Code formatting and static analysis SHALL be among the checks that the pre-merge
  verification runs, so that neither rests on a contributor remembering to run it locally.
  A contributor SHALL be able to run the same check the verification runs, and obtain the
  same verdict from it, without reproducing the pipeline that invokes it.
specification_ja: |
  コード整形と静的解析は、マージ前の自動検証が実行する確認に含まれなければならない
  （どちらも、貢献者がローカルでの実行を覚えていることに依存しないようにするため）。貢献者は、
  その検証が実行するのと同じ確認を、それを起動するパイプラインを再現することなく実行し、
  同じ判定を得られなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:03+09:00"
---
# QA-a92341b9-c873-406e-8b70-a64f56d8a7d6: コード整形と静的解析を自動検証に載せ、同じ判定を手元でも得られるようにする

## 仕様

整形の揺れと静的解析が見つける類の誤りは、レビューで指摘できるが指摘に人の注意を払わせるだけの
価値が無い。マージ前の自動検証の側へ回して、人のレビューを意味判断に集中させる。

[マージ前の自動検証を必須にし、プラットフォーム差が効く層は全プラットフォームで通す](cross-platform-verification-before-merge.md) は「マージ前の自動検証を必須にすること」を規範化し、**検証項目の中身は委譲して
いる**。本 item はその委譲を受けて、整形と静的解析を項目に含めることを定める。

**手元で同じ判定を得られること**が要点になる。パイプラインの手順としてしか書かれていない確認は、
手元で実行する手段がパイプラインの再現になり、CI と手元で結果がずれる余地が残る。指摘された
貢献者が同じ確認を手元で回せなければ、修正は CI の往復に頼ることになる。この性質を満たす
実装形（宣言的なチェック定義など）は基盤の担当で、規範として固定するのは性質の側になる。

対象言語ごとの整形器・解析器の選定、チェックの実装形、どのジョブで実行するかは本 item の
規範に含めない。[Go の整形は gofmt、静的解析は go vet と golangci-lint で行う](../adr/gofmt-vet-golangci.md) と `flake.nix` のチェック定義、[CI パイプライン（flake check の os×system マトリクスと E2E ジョブ）](../infrastructure/ci-pipeline.md) が持つ。

## 出典

[Go の整形は gofmt、静的解析は go vet と golangci-lint で行う](../adr/gofmt-vet-golangci.md)（整形は treefmt へ gofmt を追加し、静的解析は go vet + golangci-lint の check
derivation として `nix flake check` に載せる）が置いた方針と、`flake.nix` の checks 定義・
`.github/workflows/test.yml` の flake check ジョブが実運用してきた規範を、Issue #272 で
quality item として立てたもの。
