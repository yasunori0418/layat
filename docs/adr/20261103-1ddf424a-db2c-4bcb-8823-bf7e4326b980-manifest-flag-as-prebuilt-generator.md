---
id: "ADR-1ddf424a-db2c-4bcb-8823-bf7e4326b980"
type: adr
name: "--manifest を prebuilt 生成器に写す"
status: 採用
issues:
  - "#395"
origin: "Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0055 §5 を書き直した"
justifies:
  - "REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729"
supersedes:
  - "ADR-0055"
references:
  - "ADR-bac005fb-fe4d-4a3b-b9e1-517a62fe46c2"
  - "ADR-28605860-770e-41df-95dd-7dc05b20497a"
  - "ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9"
  - "ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1"
  - "ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246"
  - "ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa"
  - "ADR-6b7becf0-f029-41e3-895e-b672dce94d3d"
  - "ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3"
  - "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
---
# ADR-1ddf424a-db2c-4bcb-8823-bf7e4326b980: --manifest を prebuilt 生成器に写す

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-bac005fb-fe4d-4a3b-b9e1-517a62fe46c2, ADR-28605860-770e-41df-95dd-7dc05b20497a, ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9, ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1, ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246, ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa, ADR-6b7becf0-f029-41e3-895e-b672dce94d3d, ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3, ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0, ADR-fe035038-4415-4a1e-aa5c-509fdfce9955, GitHub Issue #395
- 起点: Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0055 §5 を書き直した

## 背景

`apply --manifest` は、build 済みの link-farm を直接適用する。HM などのモジュールの
activation は `apply --manifest` で engine を起動する。manifest の取得を生成器の 4 操作越しにする
（→ ADR-bac005fb-fe4d-4a3b-b9e1-517a62fe46c2）と、`--manifest` の処理だけが生成器の外に残る。

## 決定

`--manifest` で渡した link-farm は prebuilt 生成器に写す。Discover は与えられたパスを返し、Roots は
その link-farm の `manifest.json` から読み、Build と DryBuild はそのパスを返す。`--all` と合成する
`--manifest <name>=<link-farm>` の繰り返しでは、prebuilt 生成器が名前付き link-farm の集合を持ち、
Roots は name ごとの root を返し、Build と DryBuild は name で link-farm を選ぶ。

prebuilt 生成器に写しても、`--manifest` のフラグと挙動は変わらない。prebuilt 生成器の規範は
REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729 に置く。

## 根拠

`--manifest` の処理を生成器の外に残すと、engine へ manifest を渡す道筋が 2 本になり、root の種別の
読み方や `Targets` の導出がそれぞれ別の実装になる。prebuilt 生成器にすれば、nix 生成器・テスト
ダブルと同じ契約で扱え、`--manifest` のフラグと挙動を変えずに内部の処理だけを揃えられる。

## 棄却した代替案

### `--manifest` の処理を生成器の外に残す

root の種別の取得と `Targets` の導出を、生成器と `--manifest` の処理で二重に持つことになる。
