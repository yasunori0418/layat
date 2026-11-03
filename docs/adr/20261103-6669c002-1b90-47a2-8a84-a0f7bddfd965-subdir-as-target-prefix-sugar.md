---
id: "ADR-6669c002-1b90-47a2-8a84-a0f7bddfd965"
type: adr
name: "subdir は全 target へ前置する評価時の糖衣にし、root を動かさない"
status: 採用
origin: "次期マイルストーン追加計画の grilling（2026-07-04）。ADR-0040 §2a を書き直した"
justifies:
  - "REQ-08c8f639-790c-4970-a954-c9567538cf9b"
  - "REQ-3f541d39-da41-4ef8-858b-707f54cf6a29"
references:
  - "ADR-5b347bfa-39d6-47f0-be59-1cf31dc25509"
---
# ADR-6669c002-1b90-47a2-8a84-a0f7bddfd965: subdir は全 target へ前置する評価時の糖衣にし、root を動かさない

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-5b347bfa-39d6-47f0-be59-1cf31dc25509, `docs/spec.md`
- 起点: 次期マイルストーン追加計画の grilling（2026-07-04）。ADR-0040 §2a を書き直した

## 背景

root マーカーの適用形が持つ `subdir` を、どこで実現するかを決める必要がある。layat の意味論では、
基点を `root/sub` にすることと、全 target に `sub/` を前置することは等価である。古くなった配置の除去で使う
比較のキー・foreign symlink の判定・祖先 symlink の確認・`gitignore` の anchor 形は、どれも target の
文字列で動くからだ。

root を実際に動かすと、`manifest.json` に新しいフィールドが要る。古い engine がエラーにせず無視すると
誤配置になるので、`schemaVersion` も上げなければならない。home と system の mode では profileDir を
config の名前だけで鍵にするので、同じ名前の config が `subdir` 違いで世代の系列を共有してしまう。

## 決定

`subdir` は `normalizeManifest` が全 entry の正規化後の target に前置する、評価時の糖衣とする。
root の解決値・`rootKind`・`manifest.json` のスキーマ・`schemaVersion` は変えず、engine は
`subdir` を知らない。前置の規則は REQ-08c8f639-790c-4970-a954-c9567538cf9b、root マーカーが
種別だけを運び解決を engine が行うことは REQ-3f541d39-da41-4ef8-858b-707f54cf6a29 に置く。

## 根拠

等価なので、root を動かす実装は複雑さだけを増やし、機能では何も得ない。糖衣にすれば実装は
`normalizeManifest` の 1 か所に閉じ、engine の変更も `manifest.json` の版上げも要らない。
root が動かないので profileDir の鍵も変わらず、既存の確認と意味論は前置後の target でそのまま
正しく動く。

## 棄却した代替案

### root を実際に動かす（manifest に subdir のフィールドを足す）

`schemaVersion` の版上げが要り、同じ名前の config が `subdir` 違いで世代の系列を共有し、engine の
分岐も増える。糖衣との機能の差は無い。
