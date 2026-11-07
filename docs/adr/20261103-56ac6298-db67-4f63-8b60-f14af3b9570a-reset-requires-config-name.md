---
id: "ADR-56ac6298-db67-4f63-8b60-f14af3b9570a"
type: adr
name: "reset は config の名指しを必須にする"
status: 採用
issues:
  - "#584"
origin: "ADR-0021 §1（reset は名指し必須で --all に対応しない）を書き直した"
justifies:
  - "REQ-a8edc58f-4adc-4637-b888-ab8ccc7e73e4"
references:
  - "ADR-ddeafc2b-f5a9-4332-a0b9-b18bea6ce6f7"
  - "ADR-45380022-d959-4b22-aa05-5993ac1f3c87"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-56ac6298-db67-4f63-8b60-f14af3b9570a: reset は config の名指しを必須にする

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-ddeafc2b-f5a9-4332-a0b9-b18bea6ce6f7, ADR-45380022-d959-4b22-aa05-5993ac1f3c87
- 起点: ADR-0021 §1（reset は名指し必須で --all に対応しない）を書き直した

## 背景

reset は symlink に加えて copy target も消す（ADR-ddeafc2b-f5a9-4332-a0b9-b18bea6ce6f7）。copy には
利用者の編集が入っていることがあり、消すと戻せない。

apply は `--all` で全 config をまとめて扱える。reset にも同じ一括指定を用意するかが決まっていなかった。
rollback は、全 config を一斉に巻き戻す `--all` を破壊的すぎるとして持たない。

## 決定

reset には `--all` も root モードのフィルタも設けず、config の名指しを必須にする。複数の config を
撤去したいときは、名指しで複数回実行する。config 内の entry 単位の指定はこれまでどおり使える。

規範は REQ-a8edc58f-4adc-4637-b888-ab8ccc7e73e4 に置く。

## 根拠

全 config の配置物を copy ごと一斉に消す操作は、打ち間違い 1 つで被害が大きくなる。名指しを
必須にすれば、この一括操作そのものを構造的に起こせなくなる。

rollback が `--all` を持たないのと同じ理由で、copy まで消す reset はなおさら一括を避けるべきである。
配置を利用者が明示的に決めるという考え方にも合う。

## 棄却した代替案

### `--all` を設け、`--yes` か確認で守る

一括で撤去できるのは便利だが、確認を 1 回通すだけで全 config が消える。rollback の判断とも
食い違う。
