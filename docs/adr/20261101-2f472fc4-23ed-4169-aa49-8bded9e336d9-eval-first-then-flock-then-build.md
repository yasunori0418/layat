---
id: "ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9"
type: adr
name: "eval を先に行い、flock を取ってから build する"
status: 採用
origin: "実装前の残セマンティクス確定（2026-06-11〜06-14）。ADR-0006 §「実行フロー」・ADR-0023 §1・ADR-0024 §3 を書き直した"
justifies:
  - "REQ-60c6b7ea-e936-4ce8-bd75-ad35e9c693b9"
  - "REQ-9c111c32-8f6c-4eda-859c-dae02c0624fc"
  - "REQ-7a71a049-5876-4cfc-a65e-44e9a0349856"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:41+09:00"
---
# ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9: eval を先に行い、flock を取ってから build する

- ステータス: 採用
- 日付: 2026-11-01
- 起点: 実装前の残セマンティクス確定（2026-06-11〜06-14）。ADR-0006 §「実行フロー」・ADR-0023 §1・ADR-0024 §3 を書き直した

## 背景

ロックのキーは profileDir で、profileDir は root を解決しないと決まらない。root の解決には root の種別が
要り、それは build の成果物である manifest の中にある。build を先に行うと、ロックの外で build することに
なり、並行する apply が同じ pending out-link を奪い合う。

build する apply と、build しない reset / rollback / list-generations のどちらも、前回適用した manifest の
読み出しや、ロックのために profileDir を先に確定しなければならない。

## 決定

apply は、build の前に root の種別を読み、root を解決して profileDir を確定する。既定の nix 生成器は、
root の種別を build せずに eval で読む。次にその profileDir のロックを取り、build はロックの内側で行う。
build しないコマンドも同じ前段で profileDir を確定し、rollback と reset はそのあとでロックを取る。
`--dryrun` も同じ前段で root を解決するが、配置しないのでロックは取らない。

段ごとの手順は REQ-60c6b7ea-e936-4ce8-bd75-ad35e9c693b9、build しないコマンドの前段は
REQ-9c111c32-8f6c-4eda-859c-dae02c0624fc、`--dryrun` の扱いは REQ-7a71a049-5876-4cfc-a65e-44e9a0349856 に置く。

## 根拠

root の種別だけを build の前に安く取れば、profileDir ← root の解決 ← root の種別 ← build という循環が切れる。
build がロックの内側に入るので、pending out-link の奪い合いも同時に消え、別の対策が要らない。

build しないコマンドも同じ前段を持つので、`--root` を付けたときの profileDir の決め方が全コマンドで揃う。

## 棄却した代替案

### ロックを取ってから root を解決する

root を解決する前は profileDir が決まらないので、ロックのキーが無い。

### 一時パスに out-link を張って build し、後で profileDir へ gcroot を張り直す

eval を増やさずに済むが、ロックの外の build が残り、gcroot を張り直す段が増える。

### out-link を PID 付きの名前にする

並行する out-link の奪い合いは避けられるが、profileDir が root の解決前に決まらない循環は残る。
