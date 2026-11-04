---
id: "ADR-d504b89e-9a2b-440f-83c1-be372b12241b"
type: adr
name: "gitignore --all は project mode の全 config の target を重複なく整列して出す"
status: 採用
origin: "ADR-0018 §4a を書き直した"
justifies:
  - "REQ-1f128917-4424-4e37-8a88-e0bb23a09da7"
references:
  - "ADR-baf6346d-783b-4607-ab1c-c2c93c09e064"
---
# ADR-d504b89e-9a2b-440f-83c1-be372b12241b: gitignore --all は project mode の全 config の target を重複なく整列して出す

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-baf6346d-783b-4607-ab1c-c2c93c09e064
- 起点: ADR-0018 §4a を書き直した

## 背景

`--all` は当初 `apply` のためのフラグで、`gitignore` でも受けるかは決まっていなかった。`gitignore`
は project mode の配置物を `.gitignore` に載せるための行を出す。project に config が複数あると、
config ごとに実行して出力を足し合わせる必要があった。

## 決定

`gitignore` は `--all` を受け、projectRoot の全 config の target をソートし、重複を除いて出す。

出力の形は REQ-1f128917-4424-4e37-8a88-e0bb23a09da7 に置く。

## 根拠

リポジトリの `.gitignore` は 1 つなので、全 config の target を一度に列挙できるほうが自然である。
`gitignore` は出力するだけで配置を変えないので、一括にしても危険がない。

ソートと重複除去で、出力は config の並びや定義の重なりに左右されなくなる。再生成した `.gitignore`
の差分が順序だけで揺れない。

## 棄却した代替案

### 名指しだけを受ける

実装は単純だが、config が複数ある project では実行と連結を利用者に任せることになる。

### config の順に連結して出す

config を跨いで同じ target が重なると同じ行が二度出て、`.gitignore` に重複が残る。並びも config の
定義順に引きずられる。
