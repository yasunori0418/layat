---
id: "ADR-bac005fb-fe4d-4a3b-b9e1-517a62fe46c2"
type: adr
name: "生成器の契約を Discover / Roots / Build / DryBuild の 4 操作にする"
status: stable
issues:
  - "#395"
origin: "Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0055 §2 と ADR-0032 §6、ADR-0007 §3 の発見順を書き直した"
justifies:
  - "REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729"
  - "REQ-1cc080f6-ae91-4c1f-973e-b7054cfc0198"
references:
  - "ADR-87a2cf34-9592-4c81-8980-246bca579ca0"
  - "ADR-28605860-770e-41df-95dd-7dc05b20497a"
  - "ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9"
  - "ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1"
  - "ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246"
  - "ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa"
  - "ADR-6b7becf0-f029-41e3-895e-b672dce94d3d"
  - "ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3"
  - "ADR-e683ac05-c0e0-41c1-8b6a-38b605972fe1"
  - "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# ADR-bac005fb-fe4d-4a3b-b9e1-517a62fe46c2: 生成器の契約を Discover / Roots / Build / DryBuild の 4 操作にする

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-87a2cf34-9592-4c81-8980-246bca579ca0, ADR-28605860-770e-41df-95dd-7dc05b20497a, ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9, ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1, ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246, ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa, ADR-6b7becf0-f029-41e3-895e-b672dce94d3d, ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3, ADR-e683ac05-c0e0-41c1-8b6a-38b605972fe1, ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0, ADR-fe035038-4415-4a1e-aa5c-509fdfce9955, GitHub Issue #395
- 起点: Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0055 §2 と ADR-0032 §6、ADR-0007 §3 の発見順を書き直した

## 背景

manifest の取得は生成器越しに行う（→ ADR-87a2cf34-9592-4c81-8980-246bca579ca0）。CLI が生成器に頼む
仕事は 4 つに分かれる。entrypoint を見つける、build せずに root の種別を知る、link-farm を build する、
gcroot を張らずに link-farm の store path を得る、の 4 つである。

## 決定

生成器の契約は Discover / Roots / Build / DryBuild の 4 操作にし、引数は config 名 `name` だけにして、
`system` のような生成器固有の概念は生成器の内部に隠す。

entrypoint の発見規則は Discover を実装する生成器が持つ。nix 生成器は、`-f` が無ければ cwd で
`flake.nix` → `shell.nix` → `default.nix` の順に探し、`-f` がディレクトリならその中を同じ順で探す。

正規化後の target 一覧（`Targets`）を返す義務は、`--all` 用の一括の Roots と prebuilt 実装に限る。

各操作の入出力と事後条件は REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729、発見順と `-f` の扱いは
REQ-1cc080f6-ae91-4c1f-973e-b7054cfc0198 に置く。

## 根拠

既存の nix 呼び出しは全てこの 4 つのどれかに当たる。entrypoint の発見は Discover、rootKind の先読み eval は
Roots、out-link 付きの build は Build、gcroot を張らない build は DryBuild に当たる。

発見規則は entrypoint の形とともに生成器ごとに変わるので、CLI ではなく生成器が持つ。nix 生成器の
順序は flake を優先し、legacy の `shell.nix` / `default.nix` を後に置く。

`Targets` を使うのは、`apply --all` の cross-config の target の衝突の確認だけである。単一 config の
Roots で返すと、nix 生成器の eval が 1 本増える。

## 棄却した代替案

### 発見規則を CLI に残す

生成器を足すたびに CLI の発見処理が変わり、生成器ごとの entrypoint の形を CLI が知ることになる。

### 全ての Roots に `Targets` を課す

単一 config の apply で、使わない target 一覧のために eval が 1 本増える。

### `system` を契約の引数に出す

nix 以外の生成器にとって意味の無い引数が契約に残る。
