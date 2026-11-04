---
id: "QA-6bf957d9-17d9-4660-92b7-ebd6eeb71a8c"
type: quality
name: "ドキュメントグラフのトレーサビリティは機械的に検証する"
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
  その検証は要求されたときではなく CI で実行されなければならない。参照先を解決しない検査を
  もってこれを満たしたとみなしてはならない。
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
INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71 が、そう決めた根拠は SDR-bf25a81b-d588-4c7e-9365-fe514ee4d626 が持つ。

**スコープ境界**: 本 item が対象とするのは**ドキュメント間**のグラフ整合のみ。公開契約と
コードの対応関係（cobra の `Use:` や `mkOption` からの母集団の機械抽出と比較）はドキュメント
グラフの視野外で、quality-observability 側の課題として残る（→ Issue #203 のスコープ境界）。

## 出典

SDR-5ea31001-96f8-4324-a40b-647cdfa5cc9d、SDR-536848bb-8901-4e73-8c66-861de3f91bcd、
SDR-2457aa00-fab7-4c99-af53-2e6643c21c79、SDR-49733f2e-ed79-480d-bfb9-6361a7f95aed、
SDR-c15921a9-7aea-407c-a34b-7465783b1af1、SDR-e093c6b2-75fd-4c91-a8f0-5a89e005e4ed、
SDR-a0095ee8-570c-4f77-a6c9-bbc162a5c4cb、SDR-bf25a81b-d588-4c7e-9365-fe514ee4d626、
SDR-34916a83-0cff-4524-a8ff-5a3195af52c3、SDR-f3f9491f-d6f4-4742-82da-4933ad3d8db1、
SDR-29ef8bb4-9349-47db-9e79-926bc71bbdd0、SDR-933a9191-f60e-45bc-bebc-bc09af8154cc、
SDR-35636015-710a-4cf8-b6ba-48ca047efed0、SDR-1d295034-9bab-4e30-9ead-173e5f56c396、
SDR-9bec5976-67cd-407a-afa9-3d394a6204ca、SDR-9f2ef734-a9a7-4b93-92a7-e8ed1c7cf8fe、
SDR-d5304533-00ae-423b-a459-7c67c1cf093c、SDR-fa7eff8e-788f-48e7-951d-4cfbf95d5a1f、
SDR-3ff9f355-a667-476e-b10d-087248cdbfcf、SDR-722ce265-2aaf-4a90-915d-32d26b7a38bb、
SDR-6cc54665-c239-4235-92a8-f73e26c5a5ef が置いた方針を、
基盤（INF）から分離して
quality item として立てたもの。model.yaml の型定義・`strict_mode` の設定・採番コマンドは
INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71 が持つ。当初はリファレンスの生成（QA-0d42104c-a3c2-4bd1-9cad-a42d9e5a56a1）と 1 件にまとめていたが、
根拠 ADR も支える基盤も別で片方だけ改訂され得るため、Issue #242 のレビューで 2 件へ分けた。
