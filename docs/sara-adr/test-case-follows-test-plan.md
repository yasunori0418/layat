---
id: "SDR-3f524795-70f3-4bca-ba25-ae1e40e94dd4"
type: sara_adr
name: "test_case から test_plan へ follows を張る"
description: "test_plan とテストをつなぐ辺を test_case から test_plan への新設 relation follows とし、test_case の親に test_plan を足す。"
status: stable
sources:
  - id: "issue-648"
    resource: "https://github.com/yasunori0418/layat/issues/648"
    title: "epic: 要求からテストまでの連鎖を sara のモデルでつなぐ — test_plan とテストケース・設計とリスクの切れ目"
origin: "epic #648 の grilling セッション（2026-11-08）で確定"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
references:
  - "SDR-f3f9491f-d6f4-4742-82da-4933ad3d8db1"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T22:59:08+09:00"
---
# test_case から test_plan へ follows を張る

- ステータス: stable
- 日付: 2026-11-09
- 関連: [risk の親に test_plan を含めない](risk-parent-excludes-test-plan.md), GitHub Issue #648, `docs/model.yaml`, `docs/agents/sara-graph.md`
- 起点: epic #648 の grilling セッション（2026-11-08）で確定

## 背景

`test_plan` はテストのスコープ・レベル・アプローチを定める。2026-11-06 の監査（sara 0.12.0）で、`test_plan` から個々のテストへ届く辺がモデルに無く、どの `test_case` がどの計画に従っているかをグラフで辿れないと分かった。

[risk の親に test_plan を含めない](risk-parent-excludes-test-plan.md) は、`risk` と `test_plan` の間の peer relation を「不足が見えないうちに relation を足さない」として棄却している。今回不足が見えたのは `risk` との間でなく、`test_plan` と `test_case` の間である。

`test_plan` 9 本の specification はテスト資産の種類で書かれている。`test_case` は `resource` にテスト資産を持つので、主担当の `test_plan` は `resource` から機械的に決まる。

## 決定

relation `follows`（inverse は `is_followed_by`、upstream の primary）を新設し、`test_case` から `test_plan` へ張る。`test_case` の `parent_types` は `test_condition` と `test_plan` の 2 つにする。

`risk` の親は `requirement` と `design` のままにする。

## 根拠

計画に従って実行されるのはテスト資産であり、グラフ上でテスト資産に 1:1 で対応するのは `test_case` である。辺の起点を `test_case` にすれば、`resource` を見るだけで張り先を決められる。

`covers` はどの検証テーマを確かめるかを表し、`follows` はどの計画に従って確かめるかを表す。軸が違うので、1 件の `test_case` が両方を持つ。

`test_plan` から `is_followed_by` で `test_case` へ下り、そこから `covers`・`mitigates` を上れば、テスト計画がどの `risk` を扱うかも読める。`risk` の親に `test_plan` を加える必要は生じない。

## 棄却した代替案

### `test_condition` から `test_plan` へ辺を張る

`test_condition` は検証テーマの単位で、束ねる `test_case` の `resource` が Go テストと E2E シナリオに分かれることがある。その場合、`test_condition` の単位では主担当の `test_plan` が決まらない。

### `derives_from` を再利用する

`derives_from` は `requirement` から `use_case` へ、`quality`・`test_plan` から `solution` へ張り、上位の item から規範を導く関係を表す。`test_case` は `test_plan` から規範を導く item でなく、計画の定めたやり方で実行される実体である。同じ relation にすると、`sara query` の上流を読むときに 2 つの意味が混ざる。
