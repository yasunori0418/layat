---
id: "ADR-ef5eac75-c65f-45b3-a8b4-748e0e767989"
type: adr
name: "flake-parts module は transposition だけを行う"
status: 採用
issues:
  - "#26"
  - "#530"
origin: "Issue #26 と Issue #530 の grilling（2026-10-25）。ADR-0029 §5 と ADR-0068 §5b を書き直した"
justifies:
  - "REQ-c50df875-2cb0-4e72-8a21-858359a11cae"
references:
  - "ADR-e6820da0-fce3-4aef-8b79-d53abbca6051"
  - "ADR-db293e0c-4789-45b7-93a9-adedc021344f"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T16:58:03+09:00"
---
# ADR-ef5eac75-c65f-45b3-a8b4-748e0e767989: flake-parts module は transposition だけを行う

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-e6820da0-fce3-4aef-8b79-d53abbca6051, ADR-db293e0c-4789-45b7-93a9-adedc021344f, GitHub Issue #26, #530
- 起点: Issue #26 と Issue #530 の grilling（2026-10-25）。ADR-0029 §5 と ADR-0068 §5b を書き直した

## 背景

flake-parts module（ADR-e6820da0-fce3-4aef-8b79-d53abbca6051）を import する利用者は、同じ
`inputs.layat` から `lib` も参照できる。module に便宜の機能を足すなら、候補は 2 つある。`mkManifest` や
root マーカーを perSystem の引数として注入することと、事前展開した apply コマンドを devShell へ
注入する option を持たせることである。

## 決定

flake-parts module は `perSystem.layat` の転置だけを行う。`lib` の関数やマーカーを perSystem の引数に
注入せず、devShell へ apply コマンドを注入する option も持たない。利用者は `inputs.layat.lib` を
直接参照する。

option が `mkManifest` の derivation を入れるだけであることの規範は
REQ-c50df875-2cb0-4e72-8a21-858359a11cae に置く。

## 根拠

利用者は module を import するために `inputs.layat` を既に参照しているので、`lib` を直接使えば
足りる。注入すると、どこから来たか読み取れない引数が perSystem に増え、module も複雑になる。

devShell への注入は、module が devShell の形を決めることになる。apply コマンドを組む関数を `lib` に
公開しておけば（ADR-db293e0c-4789-45b7-93a9-adedc021344f）、利用者は任意の shell に自分で埋め込める。

## 棄却した代替案

### `mkManifest` やマーカーを perSystem の引数に注入する

`inputs.layat.lib.` の前置きを省けるが、暗黙の引数が増え、module の責務が転置の外へ広がる。

### devShell 用の option を足して apply コマンドを注入する

利用者の記述は減るが、devShell の構成を module が決めることになる。利用者が自分の devShell の
組み方を選べなくなる。
