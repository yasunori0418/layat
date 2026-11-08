---
id: "ADR-c6bc7060-7af4-4889-9ce6-5800601c9dcc"
type: adr
name: "読み取り専用の列挙は info で返す"
status: stable
sources:
  - id: "issue-552"
    resource: "https://github.com/yasunori0418/layat/issues/552"
    title: "#552"
  - id: "issue-outturn-1"
    resource: "https://github.com/yasunori0418/outturn/issues/1"
    title: "yasunori0418/outturn#1"
origin: "layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §5 を書き直した"
justifies:
  - "REQ-fa181aa6-29a2-48c3-ae07-cc1b9a3b0303"
references:
  - "ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71"
  - "ADR-feb60906-1517-46b8-8fa5-e007df4d3e11"
  - "ADR-9e328362-2e90-441a-8eca-40081ce71b1b"
  - "ADR-5d28818e-f7e1-45e8-a7fd-bbb6157b8586"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:00:56+09:00"
---
# ADR-c6bc7060-7af4-4889-9ce6-5800601c9dcc: 読み取り専用の列挙は info で返す

- ステータス: stable
- 日付: 2026-11-01
- 関連: [--json は outturn 規約に常時準拠する](outturn-conformance-principle.md), [--json は常に results\[\] エンベロープを出す](results-envelope-always.md), [エラーはエンベロープに構造化し stderr のテキストと併存させる](structured-errors-with-stderr.md), [--json は items を全件、changes を差分で持ち可逆性を reversible で表す](items-changes-reversible.md), GitHub Issue #552, yasunori0418/outturn#1
- 起点: layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §5 を書き直した

## 背景

`list-generations` の世代一覧と `gitignore` の対象パスは、副作用の無い列挙である。outturn の item は
処理単位の実行結果の記録で、id を identity から導く。世代番号のような値には identity の key に置き場が無い。

`gitignore` は、`--json` が無ければ行指向のテキストを出す。`--json` は、それとは別に opt-in で選ぶ第 2 契約である（[gitignore は行指向の出力を既定のまま JSON を opt-in にする](gitignore-text-default-json-opt-in.md)）。

## 決定

読み取り専用の列挙は、id を導く item にせず、`result.info` 配下にツール固有のインベントリとして持つ。
これらのコマンドの `items` は空配列にする。

info の形は [読み取り系の JSON ペイロードは dryRun パリティと info インベントリで表す](../requirements/json-readonly-payload.md) に置く。

## 根拠

列挙は実行結果ではないので、id を安定させる仕組みを必要としない。item にすると、世代番号をどの key に
入れるかで無理が生じる。info に置けば、item の定義を変えずに済む。

## 棄却した代替案

### 列挙の各要素を id 付きの item にする

世代番号の key の扱いで無理が生じる。outturn#1 で棄却した。
