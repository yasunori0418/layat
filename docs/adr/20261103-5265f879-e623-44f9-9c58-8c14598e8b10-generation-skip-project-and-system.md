---
id: "ADR-5265f879-e623-44f9-9c58-8c14598e8b10"
type: adr
name: "世代スキップを project mode と system mode で行う"
status: 採用
issues:
  - "#519"
  - "#493"
origin: "Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）と実装前の残セマンティクス確定の第 4 巡（2026-06-14）。ADR-0067 §1 と ADR-0022 §2 を書き直した"
justifies:
  - "REQ-46fccb80-4bae-4d37-bc19-dded88e9a9c0"
supersedes:
  - "ADR-0022"
  - "ADR-0067"
references:
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
  - "ADR-11507e12-5b0f-4f8f-bae9-fb7fd41a9203"
  - "ADR-419914ba-9896-4dd1-853b-fb514fcb9f38"
  - "ADR-41d506b4-2c5c-4a17-9954-b0de37e97de5"
  - "ADR-df8ab6ff-7636-42ff-8b47-6ee1209dfd15"
  - "ADR-52423842-bacf-4773-bbca-bcb34ae71013"
  - "ADR-10c6dc33-d188-406b-beb4-d10ccbe3c645"
  - "ADR-a91585ad-971a-4558-9f25-b7fcd6e57671"
  - "ADR-a6a1ef1b-676c-484a-9c8a-3e6d49c1b444"
  - "ADR-be5865a8-000a-491f-8438-6308076ea727"
  - "ADR-facd3173-a5ce-4d81-90a9-370fb26428ff"
  - "ADR-f3d06307-ae5b-4f17-94ad-db92e80e3415"
  - "ADR-fd3353cf-00af-4343-bd3f-3369c21d7194"
  - "ADR-8e253f2e-c255-482a-a0f8-0837e122e80f"
  - "ADR-c2ea1648-5e7a-47a1-a2c6-30169f93e22c"
  - "ADR-68f453b3-e6c7-4e0d-9978-32b5a00f627e"
  - "ADR-eee039fc-40af-42ac-bc65-979d6eaa0d06"
  - "ADR-f58e9981-bacc-4937-98b8-bddce0a057db"
  - "ADR-a0d0cf15-8234-42be-b83e-9b8271850475"
  - "ADR-1338a293-3f69-48f5-9164-efb4ec6994d6"
  - "ADR-1743c957-a732-4a44-a89a-4c03ab277f0c"
  - "ADR-0da9d817-c870-4f48-8a6d-8fb9b55e2a70"
  - "ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75"
  - "ADR-d01223d3-47a2-4bdd-ac5d-26370d553f7c"
  - "ADR-abfe1b12-1a7d-42e1-8520-c1e9fcf5846c"
  - "ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d"
  - "ADR-45380022-d959-4b22-aa05-5993ac1f3c87"
---
# ADR-5265f879-e623-44f9-9c58-8c14598e8b10: 世代スキップを project mode と system mode で行う

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd, ADR-11507e12-5b0f-4f8f-bae9-fb7fd41a9203, ADR-419914ba-9896-4dd1-853b-fb514fcb9f38, ADR-41d506b4-2c5c-4a17-9954-b0de37e97de5, ADR-df8ab6ff-7636-42ff-8b47-6ee1209dfd15, ADR-52423842-bacf-4773-bbca-bcb34ae71013, ADR-10c6dc33-d188-406b-beb4-d10ccbe3c645, ADR-a91585ad-971a-4558-9f25-b7fcd6e57671, ADR-a6a1ef1b-676c-484a-9c8a-3e6d49c1b444, ADR-be5865a8-000a-491f-8438-6308076ea727, ADR-facd3173-a5ce-4d81-90a9-370fb26428ff, ADR-f3d06307-ae5b-4f17-94ad-db92e80e3415, ADR-fd3353cf-00af-4343-bd3f-3369c21d7194, ADR-8e253f2e-c255-482a-a0f8-0837e122e80f, ADR-c2ea1648-5e7a-47a1-a2c6-30169f93e22c, ADR-68f453b3-e6c7-4e0d-9978-32b5a00f627e, ADR-eee039fc-40af-42ac-bc65-979d6eaa0d06, ADR-f58e9981-bacc-4937-98b8-bddce0a057db, ADR-a0d0cf15-8234-42be-b83e-9b8271850475, ADR-1338a293-3f69-48f5-9164-efb4ec6994d6, ADR-1743c957-a732-4a44-a89a-4c03ab277f0c, ADR-0da9d817-c870-4f48-8a6d-8fb9b55e2a70, ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75, ADR-d01223d3-47a2-4bdd-ac5d-26370d553f7c, ADR-abfe1b12-1a7d-42e1-8520-c1e9fcf5846c, ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d, ADR-45380022-d959-4b22-aa05-5993ac1f3c87, GitHub Issue #519, #493
- 起点: Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）と実装前の残セマンティクス確定の第 4 巡（2026-06-14）。ADR-0067 §1 と ADR-0022 §2 を書き直した

## 背景

project mode では、新しい link farm が前回適用したものと同一なら新しい世代を積まず、記録どおりでない
entry だけを置き直す（世代スキップ）。devShell の `shellHook` は、シェルに入るたびに apply を実行するからである。

NixOS の activation は switch のときだけでなく、boot のたびにも実行される。NixOS モジュールはこの activation
から engine を起動するので、構成を変えていなくても boot のたびに system mode の apply が実行される。世代スキップが
project mode に限られていると、boot を重ねるだけで system mode の profile の世代が増え続ける。

世代スキップのときに置き直す対象をどこまでにするかも決める必要があった。copy で置いた entry は、
利用者が内容を編集していることがある。

## 決定

世代スキップは project mode と system mode の両方で行い、home mode と fixed mode では行わない。判定と置き直しの
機構は両モードで同じものを使い、新しい CLI フラグは足さない。世代スキップのときの置き直しは symlink と
copy の両方を対象にし、copy は target が無いときだけ置き直して、内容が異なるときは触らない。

判定の条件と置き直しの規範は REQ-46fccb80-4bae-4d37-bc19-dded88e9a9c0 に置く。

## 根拠

boot のたびの再 activation は、内容の同じ apply を高い頻度で繰り返す点で `shellHook` と同じ性質を持つ。
project mode で世代の増え続けるのを止めた理由が、そのまま当てはまる。判定と置き直しは project mode で
実装と検証が済んでおり、モードごとに条件を変える理由が無い。boot の前に他のツールが target を書き換えて
いても、置き直しがそれを直す。

世代スキップは新しい世代を積まないだけで、既存の世代を消さない。system mode の rollback 先は失われない。
host が旧構成へ戻したときは、旧 link farm が profile の current 世代と異なるので、スキップされずに
新しい世代が積まれる。

copy を target が無いときだけ置き直すのは、home mode で apply のたびに place-once を実行し直すと消えた
copy が戻るのと振る舞いを揃えるためである。内容まで比べて戻すと、place-once の方針に反して利用者の
編集を壊す。src への追従は `apply --recopy` に限る。

home mode と fixed mode を据え置くのは、この決定が扱うのが system mode の boot による世代の増加だからである。
同じ内容の apply でも世代を残す両モードの振る舞いは、この変更のついでに変えない。

## 棄却した代替案

### NixOS モジュールが boot のときの起動を止める

boot のときに配置のずれを直す機会を失う。モジュールに起動条件の分岐を持ち込むことにもなる。

### 全モードで世代スキップを行う

home mode と fixed mode の世代の振る舞いを、この決定の範囲を超えて変えてしまう。

### 世代スキップを opt-in の CLI フラグにする

モジュールが必ず付けるフラグになり、system mode の既定にしない理由が無い。

### 世代スキップのときは symlink だけを置き直す

単純だが、home mode では消えた copy が戻るのに、project / system mode では戻らなくなる。

### copy の内容まで比べて src へ戻す

place-once の方針に反して利用者の編集を壊す。`shellHook` のように高い頻度で実行する場面では、
内容の比較のコストも大きい。
