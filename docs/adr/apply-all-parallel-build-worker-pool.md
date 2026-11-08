---
id: "ADR-c109ac0c-d5df-4a15-acf0-d6d85ce5402e"
type: adr
name: "apply --all の build を worker pool で並列に実行する"
status: stable
sources:
  - id: "issue-584"
    resource: "https://github.com/yasunori0418/layat/issues/584"
    title: "#584"
origin: "ADR-0039 §1（build 段階を worker pool で並列化する）を書き直した"
justifies:
  - "REQ-91f0a9a7-c2b2-4cda-8cce-cbf5d8c5d04d"
references:
  - "ADR-75df92ae-0694-452b-8fb4-350994e31b25"
  - "ADR-af37a636-a798-4c03-8fa4-e73ed6375a92"
  - "ADR-b1e1dd5e-61a4-4d04-8389-137ef2347c56"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-c109ac0c-d5df-4a15-acf0-d6d85ce5402e: apply --all の build を worker pool で並列に実行する

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-75df92ae-0694-452b-8fb4-350994e31b25, ADR-af37a636-a798-4c03-8fa4-e73ed6375a92, ADR-b1e1dd5e-61a4-4d04-8389-137ef2347c56
- 起点: ADR-0039 §1（build 段階を worker pool で並列化する）を書き直した

## 背景

`apply --all` は一括 eval の後、config ごとに `nix build` と配置を順に処理していた。所要時間の大半は
config ごとの `nix build` で、config の数に比例して直列に伸びる。

build は link-farm の derivation を store に実現するだけで、配置先には触れない。store の整合は
nix daemon が守る。target の衝突を調べるのは eval の段と配置の段で、build の段には無い。

## 決定

`apply --all` は、選択した全 config の build を配置より先に、同時実行数に上限を設けた worker pool で
並列に実行する。上限は `--jobs` で指定できる。build に失敗した config は
配置へ進めず、部分失敗として集約する。

段の分け方・`--jobs` の解決・失敗の扱いは REQ-91f0a9a7-c2b2-4cda-8cce-cbf5d8c5d04d に置く。

## 根拠

build は所要時間の大半を占め、配置先に触れないので、並列に実行しても配置の意味は変わらない。
衝突の確認は build の前に済んでいる（ADR-75df92ae-0694-452b-8fb4-350994e31b25）。

上限は layat が同時に起動する build プロセスの数を抑えるもので、nix daemon 側のビルド並列度とは
別の軸である。二重に制御することにはならない。

## 棄却した代替案

### 並列度に上限を設けない

config の数だけ `nix build` のプロセスが立ち上がり、daemon のキューとメモリを圧迫する。

### config ごとに build と配置を 1 本の流れで通す

実装はできるが、build の前に衝突の確認という関門が 1 つあるので、段を分けても重なり方は実質変わらない。
段を分けたほうが、集約と進捗の表示が単純になる。
