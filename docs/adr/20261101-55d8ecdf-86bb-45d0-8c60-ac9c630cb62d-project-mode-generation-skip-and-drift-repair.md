---
id: "ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d"
type: adr
name: "project mode では同じ link farm なら世代を積まず drift だけを直す"
status: 採用
origin: "ADR-0005 の決定「世代スキップ短絡を project mode の必須要件にする」と ADR-0017 §3 を書き直した"
justifies:
  - "REQ-46fccb80-4bae-4d37-bc19-dded88e9a9c0"
supersedes:
  - "ADR-0005"
  - "ADR-0017"
references:
  - "ADR-0058"
  - "ADR-0067"
---
# ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d: project mode では同じ link farm なら世代を積まず drift だけを直す

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-0058, ADR-0067, `docs/spec.md`
- 起点: ADR-0005 の決定「世代スキップ短絡を project mode の必須要件にする」と ADR-0017 §3 を書き直した

## 背景

project mode では devShell の `shellHook` がシェルに入るたびに apply を実行する。apply のたびに
新しい世代を積むと、世代が際限なく増える。一方で、derivation が変わっていなくても、別のツールが
target を書き換えたり消したりしていることはある。

## 決定

project mode では、新しい link farm が前回適用したものと同一なら新しい世代を積まない。その場合も
各 entry の target を `lstat` で検査し、記録どおりでない entry（書き換えられたもの・消えたもの）
だけを置き直す。同一かどうかの判定元は ADR-0058、system mode への適用は ADR-0067 が定める。
規範の細目は REQ-46fccb80-4bae-4d37-bc19-dded88e9a9c0 に置く。

## 根拠

世代を積まないことで、高い頻度で起動しても世代は増えない。`lstat` による比較は安価なので、
シェルに入るたびに実行しても負担にならない。derivation が同じなら配置も同じだという前提は
別のツールの書き換えで崩れるので、検査を残すことで、世代を積まずに配置の状態を収束させられる。

## 棄却した代替案

### derivation が同一なら何もしない

最も速いが、別のツールによる書き換え（drift）が直らない。

### 世代スキップをやめ、毎回すべてを適用し直す

drift は確実に直るが、世代が際限なく増える。
