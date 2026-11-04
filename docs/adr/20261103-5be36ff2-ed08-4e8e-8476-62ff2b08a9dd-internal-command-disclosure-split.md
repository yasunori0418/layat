---
id: "ADR-5be36ff2-ed08-4e8e-8476-62ff2b08a9dd"
type: adr
name: "内部コマンドの開示を CLI と生成器で分担する"
status: 採用
issues:
  - "#395"
origin: "Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0055 §8 を書き直した"
justifies:
  - "REQ-4ffda99a-7062-4c00-915f-70b525cb215b"
references:
  - "ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246"
  - "ADR-cd139de9-06a4-49a2-8bea-4801c8acc518"
  - "ADR-28605860-770e-41df-95dd-7dc05b20497a"
  - "ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9"
  - "ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1"
  - "ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa"
  - "ADR-6b7becf0-f029-41e3-895e-b672dce94d3d"
  - "ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3"
  - "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
---
# ADR-5be36ff2-ed08-4e8e-8476-62ff2b08a9dd: 内部コマンドの開示を CLI と生成器で分担する

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246, ADR-cd139de9-06a4-49a2-8bea-4801c8acc518, ADR-28605860-770e-41df-95dd-7dc05b20497a, ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9, ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1, ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa, ADR-6b7becf0-f029-41e3-895e-b672dce94d3d, ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3, ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0, ADR-fe035038-4415-4a1e-aa5c-509fdfce9955, GitHub Issue #395
- 起点: Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0055 §8 を書き直した

## 背景

layat は内部で実行するコマンドを利用者に開示して透明性を保つ（→ ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246）。
開示の場面は、`--help` の説明文、`--debug` での実行中の逐次表示、失敗時の表示の 3 つある。manifest の
取得が生成器越しになると、実際に何を実行するかを知っているのは生成器だけになる。

## 決定

内部コマンドの開示は 3 つの場面で分担する。`--help` の静的な説明文は CLI が持つ。`--debug` のときの
実行中の逐次開示は、生成器が実行するコマンドを生成器名付きで渡された writer へ書く。失敗時の表示は
CLI が組み立て、`--debug` のときは人間向けの要約に生成器名を添える。

開示の規範は REQ-4ffda99a-7062-4c00-915f-70b525cb215b に置く。

## 根拠

実行中に何を実行するかは生成器だけが知っているので、逐次開示は生成器が書く。`--help` は生成器を
起動する前に出す既定の nix 生成器の説明で、失敗時の案内ではないため、nix 固有の文面を生成器に閉じる
規律（→ ADR-cd139de9-06a4-49a2-8bea-4801c8acc518）の対象外として CLI に置ける。失敗時の表示は表示形式の
一部なので、ほかの表示と同じく CLI が組み立てる。

## 棄却した代替案

### 開示を全て CLI が持つ

CLI が生成器ごとの内部コマンドを知る必要があり、生成器の実装事情が CLI へ漏れる。

### 開示を全て生成器が持つ

`--help` を出すために生成器を選んで起動する必要が生じ、失敗時の表示も生成器ごとに形が揃わなくなる。
