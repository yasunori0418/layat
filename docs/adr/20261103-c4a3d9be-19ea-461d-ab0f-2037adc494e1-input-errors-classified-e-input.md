---
id: "ADR-c4a3d9be-19ea-461d-ab0f-2037adc494e1"
type: adr
name: "入力の不正を E_INPUT に分類する"
status: 採用
issues:
  - "#395"
origin: "Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0056 §5 を書き直した"
justifies:
  - "REQ-9341fa5d-836e-4023-af53-cc7d273438d1"
  - "REQ-badc7e10-0ba4-40d4-b334-6e40193119db"
  - "REQ-dec58330-6dad-47f7-8f56-2402764a89c7"
supersedes:
  - "ADR-0056"
references:
  - "ADR-64cff3f4-5997-49f4-b23a-e6ebeced7df9"
  - "ADR-36cb1e78-90c0-47d2-bbb1-4bfec5db2964"
  - "ADR-f58e9981-bacc-4937-98b8-bddce0a057db"
  - "ADR-24488788-a5e6-47a5-9074-97192880642a"
  - "ADR-1338a293-3f69-48f5-9164-efb4ec6994d6"
  - "ADR-ef9e4099-0cf4-409a-9e73-518866aea676"
---
# ADR-c4a3d9be-19ea-461d-ab0f-2037adc494e1: 入力の不正を E_INPUT に分類する

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-64cff3f4-5997-49f4-b23a-e6ebeced7df9, ADR-36cb1e78-90c0-47d2-bbb1-4bfec5db2964, ADR-f58e9981-bacc-4937-98b8-bddce0a057db, ADR-24488788-a5e6-47a5-9074-97192880642a, ADR-1338a293-3f69-48f5-9164-efb4ec6994d6, ADR-ef9e4099-0cf4-409a-9e73-518866aea676, GitHub Issue #395
- 起点: Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0056 §5 を書き直した

## 背景

エラーコードは複数のツールで共有する共通コードを再利用し、layat 固有の意味だけを `E_LAYAT_*` にする
（→ ADR-64cff3f4-5997-49f4-b23a-e6ebeced7df9）。生成器の選択（→ ADR-36cb1e78-90c0-47d2-bbb1-4bfec5db2964）で、
未知の生成器名・不正な設定ファイル・矛盾するフラグという、利用者の入力の誤りが新たに生じる。
`--manifest` は `--all` と合成できるようになり（→ ADR-24488788-a5e6-47a5-9074-97192880642a）、その形の誤りも入力の誤りに加わった。

## 決定

利用者の入力の誤りは、人間向けには exit 1 と 1 行のメッセージで報告し、`--json` では共通コード
`E_INPUT` に分類する。対象は、生成器の未知の値（どの段で指定しても）、設定ファイルの不正（パースの
失敗・未知のキー）、フラグの組み合わせの不正である。`--manifest` は、`--all` と合成するときは
`<name>=<link-farm>` を繰り返す形だけを正当とし、`-f` や `--generator` との併用、`--all` と
`<link-farm>` の形の合成などは入力の不正にする。

選択に関わる入力の不正は REQ-badc7e10-0ba4-40d4-b334-6e40193119db、`--manifest` の組み合わせの一覧は
REQ-dec58330-6dad-47f7-8f56-2402764a89c7、エンベロープでの分類は REQ-9341fa5d-836e-4023-af53-cc7d273438d1 に置く。

## 根拠

入力の誤りを生成器の失敗（`E_LAYAT_BUILD`）や汎用の失敗（`E_LAYAT_FAILED`）と区別できれば、消費者は
入力を直すべきか、再試行してよいかを判断できる。入力の誤りは layat に固有の意味ではないので、
共通コードの `E_INPUT` を使う。

## 棄却した代替案

### 汎用の `E_LAYAT_FAILED` に分類する

入力の誤りと実行時の失敗が同じコードになり、消費者が再試行の要否を判断できない。

### layat 固有のコード（`E_LAYAT_*`）を新設する

複数のツールに共通する意味に固有のコードを割り当てることになり、共通コードを再利用する方針に反する。
