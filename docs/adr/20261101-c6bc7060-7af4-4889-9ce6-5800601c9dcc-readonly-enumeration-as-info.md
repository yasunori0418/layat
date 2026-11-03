---
id: "ADR-c6bc7060-7af4-4889-9ce6-5800601c9dcc"
type: adr
name: "読み取り専用の列挙は info で返す"
status: 採用
issues:
  - "#552"
  - "yasunori0418/outturn#1"
origin: "layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §5 を書き直した"
justifies:
  - "REQ-fa181aa6-29a2-48c3-ae07-cc1b9a3b0303"
supersedes:
  - "ADR-0043"
references:
  - "ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71"
  - "ADR-0033"
---
# ADR-c6bc7060-7af4-4889-9ce6-5800601c9dcc: 読み取り専用の列挙は info で返す

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71, ADR-0033, GitHub Issue #552, yasunori0418/outturn#1
- 起点: layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §5 を書き直した

## 背景

`list-generations` の世代一覧と `gitignore` の対象パスは、副作用の無い列挙である。outturn の item は
処理単位の実行結果の記録で、id を identity から導く。世代番号のような値には identity の key に置き場が無い。

`gitignore` は、`--json` が無ければ行指向のテキストを出す。`--json` は、それとは別に opt-in で選ぶ第 2 契約である（ADR-0033）。

## 決定

読み取り専用の列挙は、id を導く item にせず、`result.info` 配下にツール固有のインベントリとして持つ。
これらのコマンドの `items` は空配列にする。

info の形は REQ-fa181aa6-29a2-48c3-ae07-cc1b9a3b0303 に置く。

## 根拠

列挙は実行結果ではないので、id を安定させる仕組みを必要としない。item にすると、世代番号をどの key に
入れるかで無理が生じる。info に置けば、item の定義を変えずに済む。

## 棄却した代替案

### 列挙の各要素を id 付きの item にする

世代番号の key の扱いで無理が生じる。outturn#1 で棄却した。
