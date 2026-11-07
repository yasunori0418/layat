---
id: "ADR-ee31b8d0-f7c5-442f-ae90-dcd1dd16b7a0"
type: adr
name: "manifest 内の target の衝突を評価時に検出する"
status: 採用
origin: "垂直トレーサー弾の着手前に行った実装前のレビュー第 6 巡。ADR-0024 §5 を書き直した"
justifies:
  - "REQ-5c6b07da-3d06-414d-8770-4f438234b322"
references:
  - "ADR-5e46ebd6-e44b-48dc-a93f-a8a3a3903967"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# ADR-ee31b8d0-f7c5-442f-ae90-dcd1dd16b7a0: manifest 内の target の衝突を評価時に検出する

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-5e46ebd6-e44b-48dc-a93f-a8a3a3903967, `docs/spec.md`
- 起点: 垂直トレーサー弾の着手前に行った実装前のレビュー第 6 巡。ADR-0024 §5 を書き直した

## 背景

`entries` は target をキーにした attrset で、キーの一意性は Nix が保証する（ADR-c494e86e-ceb9-4380-b983-4e09553b01d6）。ただし entry は
`target` フィールドでキーと違う target を明示できるので、別々のキーの entry が同じ target を指す
ことがある。この衝突は、正規化した後の target の文字列を比べれば、ファイルシステムを見なくても
判定できる。

## 決定

1 つの manifest の中で、正規化した後の target が重なったら、`normalizeManifest` が評価時に
`lib.throwIf` で止める。engine の実行時まで持ち越さない。別の manifest（別の config）どうしの
target の衝突はこの確認の対象にしない。

検出する処理の分け方は REQ-5c6b07da-3d06-414d-8770-4f438234b322 に置く。

## 根拠

設定だけで判定できる誤りは、評価時に止めるのが lib の確認の方針である。評価時に止めれば、CLI と
モジュールのどちらの使い方でも、engine を起動する前に分かる。

`normalizeManifest` は 1 つの manifest しか受け取らないので、別の manifest どうしの衝突はこの確認では
扱えない。単一の起動に載る複数の config の衝突は、`apply --all` の前段で検出する
（REQ-38506419-c23d-4fce-b167-30715ce6a69d・REQ-5923ac79-4a2d-43cd-b56c-2f1000c01b44）。

## 棄却した代替案

### target の衝突を engine の実行時に検出する

確認の処理は 1 つで済むが、評価時に止められる誤りを実行時まで遅らせることになる。
