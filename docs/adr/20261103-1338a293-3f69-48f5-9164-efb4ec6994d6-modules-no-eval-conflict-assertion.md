---
id: "ADR-1338a293-3f69-48f5-9164-efb4ec6994d6"
type: adr
name: "モジュールは configs 間の target 衝突を評価時に検査しない"
status: 採用
issues:
  - "#585"
origin: "ADR-0060 §3 と ADR-0066 §3c を書き直した"
justifies:
  - "REQ-5923ac79-4a2d-43cd-b56c-2f1000c01b44"
  - "REQ-5c6b07da-3d06-414d-8770-4f438234b322"
  - "REQ-fc1118b1-b0e8-4ddf-80f6-c70956651693"
---
# ADR-1338a293-3f69-48f5-9164-efb4ec6994d6: モジュールは configs 間の target 衝突を評価時に検査しない

- ステータス: 採用
- 日付: 2026-11-03
- 関連: GitHub Issue #585, `docs/spec.md`
- 起点: ADR-0060 §3 と ADR-0066 §3c を書き直した

## 背景

1 つの構成の中で、複数の config が正規化後に同じ target を狙うことがある。放っておくと、activation の
たびに config 同士が同じ target を奪い合う。全 config が 1 回のモジュール評価に載るので、nix の
`assertions` で評価時に止めることもできる。

一方で activation は全 config を 1 回の `apply --all` に渡し、`apply --all` は build と配置の前に、
選択した集合の target の重複を検査する。

## 決定

モジュールは configs 間の target 衝突を、評価時の `assertions` で検査しない。検出は activation の
`apply --all` の前段検査に委ね、衝突があれば build と配置の前に止めて switch を失敗させる。

検出経路の分け方は REQ-5c6b07da-3d06-414d-8770-4f438234b322、HM の configs 間衝突での停止は
REQ-5923ac79-4a2d-43cd-b56c-2f1000c01b44、前段検査に掛からない config 間の衝突の扱いは
REQ-fc1118b1-b0e8-4ddf-80f6-c70956651693 に置く。

## 根拠

- 同じ規則の検出器を 2 つ持つと、正規化の規則（属性キーの既定値・`target` の上書き・`subdir` の前置）を
  2 か所で揃え続けることになり、片方だけの変更が見逃しや誤検出を生む
- standalone の `apply --all` とモジュールの activation が同じ検査を受ける
- 代償は、止まる時機が評価時から activation に移ることだけである。どの config も配置されずに
  switch が失敗するので、誤った設定を通さない点は評価時に止めるのと変わらない

## 棄却した代替案

### モジュールの評価時に `assertions` で止める

`apply --all` の前段検査と同じ規則を nix で二重に持つことになる。

### 評価時の `assertions` と前段検査を両方持つ

検出器が 2 つになり、正規化の規則のずれが見逃しや誤検出を生む。

### 実行時の後勝ちのまま放置する

activation のたびに config 同士が target を奪い合い、上書きと警告が恒常化する。
