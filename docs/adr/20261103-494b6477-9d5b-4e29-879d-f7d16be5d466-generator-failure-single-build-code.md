---
id: "ADR-494b6477-9d5b-4e29-879d-f7d16be5d466"
type: adr
name: "生成器の失敗を --json のコード E_LAYAT_BUILD 1 本に分類する"
status: 採用
issues:
  - "#395"
origin: "Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0055 §7 を書き直した"
justifies:
  - "REQ-9341fa5d-836e-4023-af53-cc7d273438d1"
  - "REQ-7a2f1ecf-4675-45aa-80c0-a8fc58db9edd"
references:
  - "ADR-64cff3f4-5997-49f4-b23a-e6ebeced7df9"
  - "ADR-964e99a8-4aaa-46f6-b4ee-26270fc04649"
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
  at: "2026-11-04T09:27:41+09:00"
---
# ADR-494b6477-9d5b-4e29-879d-f7d16be5d466: 生成器の失敗を --json のコード E_LAYAT_BUILD 1 本に分類する

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-64cff3f4-5997-49f4-b23a-e6ebeced7df9, ADR-964e99a8-4aaa-46f6-b4ee-26270fc04649, ADR-28605860-770e-41df-95dd-7dc05b20497a, ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9, ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1, ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246, ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa, ADR-6b7becf0-f029-41e3-895e-b672dce94d3d, ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3, ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0, ADR-fe035038-4415-4a1e-aa5c-509fdfce9955, GitHub Issue #395
- 起点: Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0055 §7 を書き直した

## 背景

layat 固有の失敗は `E_LAYAT_*` のコードで表す（→ ADR-64cff3f4-5997-49f4-b23a-e6ebeced7df9）。
`E_LAYAT_BUILD` は nix の失敗を表していた。生成器の失敗は段階と種類を持つ構造化した値で返る
（→ ADR-964e99a8-4aaa-46f6-b4ee-26270fc04649）ので、それを `--json` のコードへどう対応させるかを決める必要がある。

## 決定

生成器が manifest を評価・生成できなかった失敗は、`--json` で `E_LAYAT_BUILD` 1 本に分類し、新しい
コードは足さない。対象は prebuilt 以外の生成器の roots / build の段階の失敗に限る。discover の段階の
失敗と prebuilt 生成器の失敗は、生成器を導入する前と同じく、原因の error から従来の規則で分類する。
失敗の種類はコードに対応させず、エンベロープに生成器名を出さない。

エンベロープのコードの層分けは REQ-9341fa5d-836e-4023-af53-cc7d273438d1、生成器の失敗の分類の細目は
REQ-7a2f1ecf-4675-45aa-80c0-a8fc58db9edd に置く。

## 根拠

失敗の種類ごとにコードを分けると、消費者が生成器の内部分類に依存するコードが増える。種類の違いは
エラーメッセージと人間向けの案内で伝えれば足りる。

entrypoint が見つからない、`-f` のパスが無い、入力の link-farm を読めない、といった失敗は、生成器が
manifest を作れなかったのではなく入力の問題である。生成器を導入する前の分類を保つほうが、消費者の
判断に合う。

生成器名を出さないのは、機械可読の契約を生成器の実装事情から切り離すためである。

## 棄却した代替案

### 失敗の種類ごとに `--json` のコードを分ける（`E_NOTFOUND` など）

config 名が見つからない失敗を `E_NOTFOUND` にするなど、生成器の内部分類に依存するコードが増える。

### discover と prebuilt の失敗も `E_LAYAT_BUILD` にする

entrypoint の不在や読めない link-farm が build の失敗と同じコードになり、消費者が入力の誤りと
評価の失敗を区別できなくなる。

### エンベロープに生成器名のフィールドを足す

生成器の実装事情が機械可読の契約に現れ、生成器を足すたびに消費者の扱う値が増える。
