---
id: "ADR-df8ab6ff-7636-42ff-8b47-6ee1209dfd15"
type: adr
name: "last-applied 記録を profileDir と同じ鍵で state 基底に置く"
status: stable
issues:
  - "#580"
  - "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0058 §1 を書き直した"
justifies:
  - "REQ-9d622539-a1e5-4719-af51-e0cd85b66f35"
  - "DSG-3c156bbe-e72b-43dd-8cb0-2cc105314ad5"
references:
  - "ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988"
  - "ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d"
  - "ADR-d97154cd-a2e5-4853-a682-800ff9435c89"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
  - "ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# ADR-df8ab6ff-7636-42ff-8b47-6ee1209dfd15: last-applied 記録を profileDir と同じ鍵で state 基底に置く

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988, ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d, ADR-d97154cd-a2e5-4853-a682-800ff9435c89, ADR-fe035038-4415-4a1e-aa5c-509fdfce9955, ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93, GitHub Issue #580, #406
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0058 §1 を書き直した

## 背景

古くなった配置の除去・drift 修復・reset・rollback は、「前回 layat がこの root
に何を置いたか」を比較の基準として使う。この比較の基準は nix profile の current 世代の manifest
から読んでいた。

世代管理が layat の外へ出る終着点では、profile は layat の所有物でなくなる。比較の基準を profile
から読み続けると、layat は配置のために世代管理 CLI の内部レイアウトへ依存し続ける。「この root
に最後に置いたもの」は世代の概念ではなく FS の事実なので、layat
が自分で記録すれば世代の側から読まずに済む。

## 決定

apply は、最後に配置した manifest を layat 自前の last-applied 記録に書く。記録は state
基底の下に置き、鍵は profileDir と同じ規則で決める（ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988）。
パスとレイアウトの細目は DSG-3c156bbe-e72b-43dd-8cb0-2cc105314ad5 に置く。

## 根拠

鍵を profileDir と揃えれば、config と記録の対応が profile と同じ規則で決まり、鍵の解決を 1
か所にまとめられる。roothash 系列では backref から root を逆引きできる点も profile と変わらない。

記録を state 基底に置けば、profile が layat の外へ出ても layat は記録を読める。

## 棄却した代替案

### パイプラインが `--previous <link-farm>` で前回の状態を渡す

終着点の構成では成り立つが、前回の状態の識別子を保持する責務を呼び出し側が負う。layat
が自分で記録すれば、呼び出し側は状態を持たずに済む。

### layat が世代管理 CLI に訊く

配置のために世代側の API と内部レイアウトへ依存し続ける。世代管理を切り出す目的と矛盾する。

### 記録を link-farm の中に置き、世代に捕捉させる

profile が layat の外へ出ると読めなくなる。

### 記録を flock の鍵にする

flock・prune・`--json` の payload は profileDir に由来する責務として残す。記録の導入で、
これらの鍵まで動かさない。
