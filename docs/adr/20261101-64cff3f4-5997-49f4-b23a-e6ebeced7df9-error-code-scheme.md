---
id: "ADR-64cff3f4-5997-49f4-b23a-e6ebeced7df9"
type: adr
name: "エラーコードは共通コードを再利用し固有の意味だけを E_LAYAT_* にする"
status: 採用
issues:
  - "#552"
  - "yasunori0418/outturn#1"
origin: "layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §8 のエラーコードを書き直した"
justifies:
  - "REQ-9341fa5d-836e-4023-af53-cc7d273438d1"
  - "REQ-2ea19863-eaa2-466b-b1ed-3f56f6417c62"
supersedes:
  - "ADR-0043"
references:
  - "ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71"
  - "ADR-9e328362-2e90-441a-8eca-40081ce71b1b"
---
# ADR-64cff3f4-5997-49f4-b23a-e6ebeced7df9: エラーコードは共通コードを再利用し固有の意味だけを E_LAYAT_* にする

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71, ADR-9e328362-2e90-441a-8eca-40081ce71b1b, GitHub Issue #552, yasunori0418/outturn#1
- 起点: layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §8 のエラーコードを書き直した

## 背景

outturn §6 はエラーコードを二層で命名する。複数のツールで共有する共通コードと、`E_<TOOL>_<NAME>` の
形のツール別コードである。layat は、失敗の種類ごとにどちらを使うかを決める必要がある。

## 決定

共通レジストリにある意味（ロック・I/O・未検出・権限・入力の不正など）には共通コードを使い、layat に
固有の意味（配置の衝突・manifest 生成器の失敗）だけを `E_LAYAT_*` で定義する。警告も同じく固有の
ものを `W_LAYAT_*` とする。分類するのは sentinel や型で判定できる失敗に限り、それ以外はツールの総称
コード `E_LAYAT_FAILED` にする。

エラーコードの一覧と割り当ては REQ-9341fa5d-836e-4023-af53-cc7d273438d1 に、配置の衝突と警告のコードの
割り当ては REQ-2ea19863-eaa2-466b-b1ed-3f56f6417c62 に置く。

## 根拠

共通コードを使えば、消費側はツールを問わず同じ意味として扱える。ツール別のコードを固有の意味に限れば、
共通レジストリと意味の重なる定義が増えない。

## 棄却した代替案

元の検討では、代替案を記録していない。
