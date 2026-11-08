---
id: "ADR-964e99a8-4aaa-46f6-b4ee-26270fc04649"
type: adr
name: "生成器の失敗は構造化した値で CLI へ返す"
status: 採用
issues:
  - "#396"
origin: "Issue #396（評価器の診断の扱い）の結論。ADR-0055 §6 の失敗の返し方と表示の主体を書き直した"
justifies:
  - "REQ-7a2f1ecf-4675-45aa-80c0-a8fc58db9edd"
references:
  - "ADR-6687c0dc-957d-4913-99c6-168b0fe552ad"
  - "ADR-28605860-770e-41df-95dd-7dc05b20497a"
  - "ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9"
  - "ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1"
  - "ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246"
  - "ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa"
  - "ADR-6b7becf0-f029-41e3-895e-b672dce94d3d"
  - "ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3"
  - "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-964e99a8-4aaa-46f6-b4ee-26270fc04649: 生成器の失敗は構造化した値で CLI へ返す

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-6687c0dc-957d-4913-99c6-168b0fe552ad, ADR-28605860-770e-41df-95dd-7dc05b20497a, ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9, ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1, ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246, ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa, ADR-6b7becf0-f029-41e3-895e-b672dce94d3d, ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3, ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0, ADR-fe035038-4415-4a1e-aa5c-509fdfce9955, GitHub Issue #396
- 起点: Issue #396（評価器の診断の扱い）の結論。ADR-0055 §6 の失敗の返し方と表示の主体を書き直した

## 背景

生成器の失敗は、人間向けの表示・`--json` のエンベロープ・`--debug` の表示のどれにも現れる。
生成器の診断は実行中に既に stderr へ流れている（→ ADR-6687c0dc-957d-4913-99c6-168b0fe552ad）。

## 決定

生成器の失敗は、段階（discover / roots / build）・種類・要約・解決策の案内・キャプチャした診断を持つ
構造化した値（`generator.Error`）で CLI へ返す。どの形で表示するかは CLI が決める。人間向けの表示は
要約と案内だけにし、生の診断は再掲しない。

フィールドの一覧と各表示形式での扱いは REQ-7a2f1ecf-4675-45aa-80c0-a8fc58db9edd に置く。

## 根拠

表示形式（人間向け・`--json`・`--debug`）は生成器によらないので CLI が持つ。生成器は失敗の中身だけを
返せばよく、生成器を足しても表示の実装は変わらない。生の診断は既に stderr へ流れているので、人間向けに
再掲すると同じ内容が 2 回出る。

## 棄却した代替案

### 生成器が整形済みの文を返す

表示形式ごとの整形を生成器ごとに持つことになり、`--json` と人間向けの書き分けが生成器へ漏れる。

### 失敗を error の文字列だけで返す

CLI が段階や種類を文字列から判定することになり、生成器固有の文字列判定が CLI へ戻る。
