---
id: "ADR-b08479ba-0eae-4668-a4a2-910cd40bd316"
type: adr
name: "apply の修飾フラグを --all と合成できるようにする"
status: stable
sources:
  - id: "issue-584"
    resource: "https://github.com/yasunori0418/layat/issues/584"
    title: "#584"
origin: "ADR-0021 §3b（apply --all --recopy を合成できるようにする）を書き直した"
justifies:
  - "REQ-687e225f-5046-46db-88fb-f9e527a1e97a"
references:
  - "ADR-fb01d942-2875-4dcb-a7df-879e18f63883"
  - "ADR-a51897c5-f709-4b7a-8db2-61e6ebb0040c"
  - "ADR-45380022-d959-4b22-aa05-5993ac1f3c87"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-b08479ba-0eae-4668-a4a2-910cd40bd316: apply の修飾フラグを --all と合成できるようにする

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-fb01d942-2875-4dcb-a7df-879e18f63883, ADR-a51897c5-f709-4b7a-8db2-61e6ebb0040c, ADR-45380022-d959-4b22-aa05-5993ac1f3c87
- 起点: ADR-0021 §3b（apply --all --recopy を合成できるようにする）を書き直した

## 背景

`--recopy` は apply の修飾フラグとして設けた（ADR-fb01d942-2875-4dcb-a7df-879e18f63883）。apply は
`--all` で選んだ複数の config を一度に扱えるが、`--all` と `--recopy` を同時に指定したときの扱いが
決まっていなかった。

layat は、独立した操作をサブコマンドにし、操作の修飾をフラグにしている
（ADR-a51897c5-f709-4b7a-8db2-61e6ebb0040c）。

## 決定

apply の修飾フラグは `--all` と合成できる。`--all`（と root モードのフィルタ）で選んだ各 config に、
修飾フラグをそれぞれ適用する。

規範は REQ-687e225f-5046-46db-88fb-f9e527a1e97a に置く。

## 根拠

`--all` は対象の config を選ぶだけで、修飾フラグは各 config の apply のしかたを変えるだけである。
両者は別の軸なので、組み合わせを特別に扱う理由が無い。

## 棄却した代替案

### `--all` と修飾フラグの同時指定を拒否する

全 config を `src` に追従させたいとき、名指しで 1 つずつ実行することになる。拒否で守れるものも無い。
