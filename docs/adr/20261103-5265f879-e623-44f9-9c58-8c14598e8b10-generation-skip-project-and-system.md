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
references:
  - "ADR-0022"
  - "ADR-0058"
  - "ADR-0066"
  - "ADR-0067"
  - "ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d"
---
# ADR-5265f879-e623-44f9-9c58-8c14598e8b10: 世代スキップを project mode と system mode で行う

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-0022, ADR-0058, ADR-0066, ADR-0067, ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d, GitHub Issue #519, #493
- 起点: Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）と実装前の残セマンティクス確定の第 4 巡（2026-06-14）。ADR-0067 §1 と ADR-0022 §2 を書き直した

## 背景

project mode では、新しい link farm が前回適用したものと同一なら新しい世代を積まず、記録どおりでない
entry だけを置き直す（世代スキップ）。devShell の `shellHook` は、シェルに入るたびに apply を実行するからである。

NixOS の activation は switch のときだけでなく、boot のたびにも実行される。NixOS モジュールはこの activation
から engine を起動するので、構成を変えていなくても boot のたびに system mode の apply が走る。世代スキップが
project mode に限られていると、boot を重ねるだけで system mode の profile の世代が増え続ける。

世代スキップのときに置き直す対象をどこまでにするかも決める必要があった。copy で置いた entry は、
利用者が内容を編集していることがある。

## 決定

世代スキップは project mode と system mode の両方で行い、home mode では行わない。判定と置き直しの
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

home mode を据え置くのは、この決定が扱うのが system mode の boot による世代の増加だからである。
同じ内容の apply でも世代を残す home mode の振る舞いは、この変更のついでに変えない。

## 棄却した代替案

### NixOS モジュールが boot のときの起動を止める

boot のときに配置のずれを直す機会を失う。モジュールに起動条件の分岐を持ち込むことにもなる。

### 全モードで世代スキップを行う

home mode の世代の振る舞いを、この決定の範囲を超えて変えてしまう。

### 世代スキップを opt-in の CLI フラグにする

モジュールが必ず付けるフラグになり、system mode の既定にしない理由が無い。

### 世代スキップのときは symlink だけを置き直す

単純だが、home mode では消えた copy が戻るのに、project / system mode では戻らなくなる。

### copy の内容まで比べて src へ戻す

place-once の方針に反して利用者の編集を壊す。`shellHook` のように高い頻度で実行する経路では、
内容の比較のコストも大きい。
