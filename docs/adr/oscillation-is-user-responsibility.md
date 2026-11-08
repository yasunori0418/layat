---
id: "ADR-b3f3b68e-5096-46d1-bafd-53a9ab4711f4"
type: adr
name: "配置の振動はユーザー責任とし、文書で注意するに留める"
status: stable
sources:
  - id: "issue-550"
    resource: "https://github.com/yasunori0418/layat/issues/550"
    title: "#550"
origin: "実装前残セマンティクス確定の第5巡（2026-06-14）で確定。ADR-0023 §5（5b）と ADR-0024 §10 を書き直した"
justifies:
  - "REQ-fc1118b1-b0e8-4ddf-80f6-c70956651693"
references:
  - "ADR-04cc2a03-78dd-4bf3-9cc7-3b8ed7faf429"
  - "ADR-27f99ee6-5fb4-4dbe-b0fd-8267eecaa3a1"
  - "ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d"
  - "ADR-3aebefee-405d-4342-b754-eef8eac29b1b"
  - "ADR-095220c9-5a3d-4fd4-94bd-8fa1a02dd36d"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# ADR-b3f3b68e-5096-46d1-bafd-53a9ab4711f4: 配置の振動はユーザー責任とし、文書で注意するに留める

- ステータス: stable
- 日付: 2026-11-01
- 関連: ADR-04cc2a03-78dd-4bf3-9cc7-3b8ed7faf429, ADR-27f99ee6-5fb4-4dbe-b0fd-8267eecaa3a1, ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d, ADR-3aebefee-405d-4342-b754-eef8eac29b1b, ADR-095220c9-5a3d-4fd4-94bd-8fa1a02dd36d, GitHub Issue #550
- 起点: 実装前残セマンティクス確定の第5巡（2026-06-14）で確定。ADR-0023 §5（5b）と ADR-0024 §10 を書き直した

## 背景

別の config A と B が同じ target を狙うと、ドリフト修復（ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d）と cross-config の後勝ち
（ADR-095220c9-5a3d-4fd4-94bd-8fa1a02dd36d）が組み合わさって、配置が往復し続ける。A が置き、
B が foreign として奪い、次の shellHook で A が取り戻す、という動きである。1 回きりの後勝ちと違い、シェルに入るたびに繰り返す。そのたびに
foreign symlink の warning も出る。

## 決定

この振動は「同じ target を複数の config で狙わない」というユーザーの責任とし、engine の実行時に
振動を検知して止める機構を持たない。warning は振動のあいだ出続けてよく、MVP ではそれを抑えたり
まとめたりする機構も持たない。config 間の target の重複を解消するという直し方は、文書で注意するに留める。

1 回の `apply --all` に載る config どうしの衝突のうち、eval で検出できるものは配置の前に止める
（ADR-75df92ae-0694-452b-8fb4-350994e31b25）。この決定が扱うのは、それ以外の衝突である。

振動を可視化する warning の規範は REQ-fc1118b1-b0e8-4ddf-80f6-c70956651693 に置く。

## 根拠

振動は設定ミスの症状で、warning が出続けることがそのシグナルになる。抑えたりまとめたりすると、
直すべき設定ミスが見えにくくなる。warning は成功時に何も出さない出力規律（ADR-36bfbfbd-f1dd-4581-a75a-943d80f6ccf5）でも常に stderr に
出るので、ユーザーの目に届く。

振動を止めるには、どちらの config が正しいかを engine が判断する必要がある。その判断材料は engine に
無い。

## 棄却した代替案

### engine の実行時に振動を検知して止める

実装が重く、同じ target を複数の config で狙わないという前提を破る使い方に対して、過剰な対策に
なる。どちらの config を勝たせるかも engine には決められない。

### 同じ warning を抑える、またはまとめる

設定ミスのシグナルを弱めるだけで、振動は止まらない。
