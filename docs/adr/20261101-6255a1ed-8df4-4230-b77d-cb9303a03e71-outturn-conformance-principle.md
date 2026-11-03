---
id: "ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71"
type: adr
name: "--json は outturn 規約に常時準拠する"
status: 採用
issues:
  - "#552"
  - "yasunori0418/outturn#1"
origin: "layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §1 を書き直した"
justifies:
  - "REQ-a5053191-1c6a-449b-9c5e-5ff49dc5aead"
  - "TP-d3000054-42d9-4bac-912a-dd3abc38d3e9"
supersedes:
  - "ADR-0043"
references:
  - "ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed"
  - "ADR-f85f1afd-9b87-4f7a-922e-97c62554f198"
---
# ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71: --json は outturn 規約に常時準拠する

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed, ADR-f85f1afd-9b87-4f7a-922e-97c62554f198, GitHub Issue #552, yasunori0418/outturn#1, outturn specVersion 1（yasunori0418/outturn）
- 起点: layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §1 を書き直した

## 背景

outturn（specVersion 1）は、エコシステムの各ツールを stdout / stdin の JSON パイプで合成するための
共通規格である。単一責務のツールを outturn のパイプで組み立てることが、ADR-0004 の north-star に
向かう手段になる。各ツールの出力を固定順に合成し、失敗時は逆順に rollback する
オーケストレーションツールが構想されており、layat はその最初の適合実装にあたる。

準拠を機能ごとに決めると、後から足す機能が独自の形に逸れる余地が残る。

## 決定

layat が出す JSON は、今ある機能も今後足す機能も、すべて outturn 規約に準拠させる。これを個々の
機能の判断ではなく、layat の恒常的な設計制約とする。ツール固有の情報は outturn の `info` 配下にだけ置く。

エンベロープの形は REQ-a5053191-1c6a-449b-9c5e-5ff49dc5aead に、適合の検証方法は
TP-d3000054-42d9-4bac-912a-dd3abc38d3e9 に置く。

## 根拠

オーケストレーションツールによる合成は、ツール間のやりとりが規格だけに依存して初めて成り立つ。
1 つのツールが独自の形に逸れると合成の前提が崩れるので、layat 内部の出力の都合より準拠を優先する。

## 棄却した代替案

### 独自のエンベロープを保ち、outturn は参照にとどめる

outturn の消費側が解釈できず、合成に参加できない。

### 準拠を `--json` の初期実装だけの決定にする

将来の機能が独自の形に逸れる余地が残る。
