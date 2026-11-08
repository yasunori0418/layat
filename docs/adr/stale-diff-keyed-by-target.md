---
id: "ADR-95b04ff7-8f9d-4ed1-a6e3-c22e4901c895"
type: adr
name: "古くなった配置の除去で使う比較のキーを解決後の target にする"
status: stable
origin: "ADR-0014 §3a を書き直した"
justifies:
  - "REQ-0b0cd1e3-bfeb-45c1-978d-e2e11c568336"
  - "REQ-1be4d678-959c-44d7-a346-44bfd95af56e"
references:
  - "ADR-c494e86e-ceb9-4380-b983-4e09553b01d6"
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
  - "ADR-ee31b8d0-f7c5-442f-ae90-dcd1dd16b7a0"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-95b04ff7-8f9d-4ed1-a6e3-c22e4901c895: 古くなった配置の除去で使う比較のキーを解決後の target にする

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-c494e86e-ceb9-4380-b983-4e09553b01d6, ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd, ADR-ee31b8d0-f7c5-442f-ae90-dcd1dd16b7a0, `docs/spec.md`
- 起点: ADR-0014 §3a を書き直した

## 背景

古くなった配置の除去は、前世代の manifest と今回の manifest を entry ごとに比べ、今回に無いものを
除く。比べるキーには、以前は手で付けた `name` を使っていた。entries を target をキーにした
attrset にすると、候補は属性キーと解決後の target になる。属性キーは `target` の明示で上書き
でき、両者は一致しないことがある。

## 決定

古くなった配置の除去で前世代と今回の entry を比べるキーは、解決後の target とする。属性キーは
使わない。manifest に記録する項目は REQ-0b0cd1e3-bfeb-45c1-978d-e2e11c568336、世代と古くなった配置の
除去の機構は REQ-1be4d678-959c-44d7-a346-44bfd95af56e に置く。

## 根拠

ディスクに実際に置かれるのは target なので、target をキーにすれば「何を置いたか」と比べるキーが一致
する。属性キーを付け替えても target が同じなら、同じ配置物として扱われ、無駄な除去と再配置が
起きない。

## 棄却した代替案

### 属性キーで比較する

キーを論理ラベルにして `target` を上書きした entry では、キーを変えただけで同じ配置物が古くなったと
みなされ、除去されてから置き直される。逆に target を変えてもキーが同じなら、古い配置物が残る。
