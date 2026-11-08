---
id: "QA-6bf957d9-17d9-4660-92b7-ebd6eeb71a8c"
type: quality
name: "ドキュメントグラフのトレーサビリティは機械的に検証する"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
specification: |
  The traceability of the document graph — that every reference resolves to an item that
  exists, that identifiers are unique, and that the graph is acyclic — SHALL be verified
  mechanically rather than by manual inspection, and that verification SHALL run in CI
  rather than on request. A check that does not resolve the reference it inspects SHALL NOT
  be treated as satisfying this.
specification_ja: |
  ドキュメントグラフのトレーサビリティ（すべての参照が実在する item を指すこと・ID が一意で
  あること・グラフが非循環であること）は、人手の点検ではなく機械的に検証されなければならず、
  その検証は要求されたときではなく CI で実行されなければならない。参照先を解決しない確認を
  もってこれを満たしたとみなしてはならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:03+09:00"
---
# QA-6bf957d9-17d9-4660-92b7-ebd6eeb71a8c: ドキュメントグラフのトレーサビリティは機械的に検証する

## 仕様

文書間の参照が指す先を失う、というずれ方は「気づいた人が直す」運用では検出できない。存在
しない ID への参照・ID の重複・循環は、人が読んで気づく類の誤りではないため、機械検証に置く。

**参照先の実在を確かめる**ことを条件に含めるのが要点になる。ID を文字列として比べる
だけの確認（grep による比較など）は、存在しない ID を書いても検出できず、カバー率だけが
実態より高く出る。

検証の**強度**（未接続を失敗とするか警告に留めるか）は本 item の規範に含めない。検証が満たす
べき条件と、検出した未接続をどの重さで扱うかは切り離して決められる。現在の設定は
[トレーサビリティ検証基盤（sara によるドキュメントグラフの機械検証）](../infrastructure/traceability-verification.md) が、そう決めた根拠は [sara check を strict モードで回す](../sara-adr/sara-check-strict-mode.md) が持つ。

**スコープ境界**: 本 item が対象とするのは**ドキュメント間**のグラフ整合のみ。公開契約と
コードの対応関係（cobra の `Use:` や `mkOption` からの母集団の機械抽出と比較）はドキュメント
グラフの視野外で、quality-observability 側の課題として残る（→ Issue #203 のスコープ境界）。

## 出典

[ドキュメントは sara のナレッジグラフとして item に分割する](../sara-adr/docs-as-sara-knowledge-graph.md)、[概要文書は item へのリンク集に縮退させる](../sara-adr/overview-docs-as-link-index.md)、
[model.yaml で組み込みモデルを全面置換し、汎用の型名を使う](../sara-adr/model-yaml-full-replacement.md)、[欠陥はグラフに置かず GitHub Issues で管理する](../sara-adr/defects-in-github-issues.md)、
[sara 品質管理機構の決定は sara_adr 型に分ける](../sara-adr/sara-adr-type-for-doc-process-decisions.md)、[ADR も UUID で採番し、連番の旧 ADR は legacy_adr 型に凍結する](../sara-adr/adr-uuid-ids-and-legacy-type.md)、
[root は solution と ADR 系の型だけにする](../sara-adr/roots-solution-and-adr-types-only.md)、[sara check を strict モードで回す](../sara-adr/sara-check-strict-mode.md)、
[テスト計画の型を品質方針の型から分けて設ける](../sara-adr/quality-and-test-plan-types.md)、[risk の親に test_plan を含めない](../sara-adr/risk-parent-excludes-test-plan.md)、
[infrastructure を技術基盤として再定義し root から外す](../sara-adr/infrastructure-redefined-non-root.md)、[infrastructure 型を改名しない](../sara-adr/infrastructure-name-kept.md)、
[design の satisfies 対象に test_plan を含める](../sara-adr/design-satisfies-test-plan.md)、[ADR の justifies 対象を specification を持つ全型に広げる](../sara-adr/adr-justifies-quality-and-test-plan.md)、
[欠陥 issue のクローズ時にテスト系 item の欠落を追補する](../sara-adr/defect-close-backfills-risks.md)、[item にできない外部参照は text フィールドへ逃がす](../sara-adr/external-references-as-text-fields.md)、
[提案を棄却した決定は status 不採用 で記録する](../sara-adr/adr-status-rejected.md)、[adr の status_note を持たない](../sara-adr/adr-status-note-removed.md)、
[ADR の失効は supersedes と 廃止 で表す](../sara-adr/adr-supersedes-and-deprecated-status.md)、[ID はフル UUID に統一し 8 文字省略形を使わない](../sara-adr/full-uuid-ids-no-short-form.md)、
[sara check を required status check にしない](../sara-adr/sara-check-not-required.md) が置いた方針を、
基盤（INF）から分離して
quality item として立てたもの。model.yaml の型定義・`strict_mode` の設定・採番コマンドは
[トレーサビリティ検証基盤（sara によるドキュメントグラフの機械検証）](../infrastructure/traceability-verification.md) が持つ。当初はリファレンスの生成（[リファレンスドキュメントは記述対象のソースから生成し、生成物を持たない](reference-docs-generated-from-source.md)）と 1 件にまとめていたが、
根拠 ADR も支える基盤も別で片方だけ改訂され得るため、Issue #242 のレビューで 2 件へ分けた。
