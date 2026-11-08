---
id: "SDR-5ea31001-96f8-4324-a40b-647cdfa5cc9d"
type: sara_adr
name: "ドキュメントは sara のナレッジグラフとして item に分割する"
status: stable
sources:
  - id: "issue-203"
    resource: "https://github.com/yasunori0418/layat/issues/203"
    title: "#203"
  - id: "issue-206"
    resource: "https://github.com/yasunori0418/layat/issues/206"
    title: "#206"
  - id: "issue-207"
    resource: "https://github.com/yasunori0418/layat/issues/207"
    title: "#207"
origin: "epic #203 の grilling セッション（2026-07-31・2026-08-01）で確定。ADR-0048 §1 を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:32+09:00"
---
# SDR-5ea31001-96f8-4324-a40b-647cdfa5cc9d: ドキュメントは sara のナレッジグラフとして item に分割する

- ステータス: stable
- 日付: 2026-11-01
- 関連: GitHub Issue #203, #206, #207
- 起点: epic #203 の grilling セッション（2026-07-31・2026-08-01）で確定。ADR-0048 §1 を書き直した

## 背景

layat の仕様・設計・テストは、`docs/spec.md`・`docs/design.md`・`docs/concept.md` と ADR に散文で
書かれていた。この形では、要求とリスクの紐付けが一貫しているかを機械的に確かめられない。

PR #206 は grep と awk による ID の比較を試作したが、参照先 ID が実在するかを確かめていなかった。
同じ PR で sara を試し、破損した参照と階層違反を検出できることは確認していた。ただ、散文中心の構造とは
相性が悪いとも結論していた。

## 決定

ドキュメントを sara のナレッジグラフとして扱い、要求・設計・リスク・テストを 1 ファイル 1 主張の
item に分割する。#206 の「相性が悪い」は散文の構造を保つ前提での評価で、ここではその構造自体を
変える。分割のコストは受け入れる。

検証基盤の構成は [トレーサビリティ検証基盤（sara によるドキュメントグラフの機械検証）](../infrastructure/traceability-verification.md) に置く。

## 根拠

主目的は、要求とリスクの紐付けの一貫性を管理することにある。ファイル単位の item では「`spec.md` に
テストがある」までしか言えず、「この要求のテストが無い」を検出できない。検出したいなら、主張の
単位まで分割するしかない。

## 棄却した代替案

### grep / awk の比較に参照先の実在の確認を足して使う

sara は参照先の実在に加えて、ID の重複・循環・階層違反も検出する。同等以上のものがあるのに、
自前の実装を保守する理由が無い。

### 散文の構造を保ち、sara は ADR だけに使う

要求とリスクの紐付けという主目的が達成できない。
