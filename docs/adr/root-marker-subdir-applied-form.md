---
id: "ADR-5b347bfa-39d6-47f0-be59-1cf31dc25509"
type: adr
name: "root マーカーに subdir の適用形を持たせる"
status: stable
origin: "次期マイルストーン追加計画の grilling（2026-07-04）。ADR-0040 §1a を書き直した"
justifies:
  - "REQ-37b56673-6e40-4a1b-a2a7-5d3c084e3e66"
  - "REQ-08c8f639-790c-4970-a954-c9567538cf9b"
references:
  - "ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed"
  - "ADR-eeeb9dc8-9089-43ec-946a-c3ed76e98137"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:26:59+09:00"
---
# root マーカーに subdir の適用形を持たせる

- ステータス: stable
- 日付: 2026-11-03
- 関連: [root は 3 つの marker で表す](three-root-markers.md), [layat.lib.mkEnv で buildEnv 合成環境を提供する](mkenv-buildenv.md), `docs/spec.md`, [root](../ubiquitous-language/root.md)
- 起点: 次期マイルストーン追加計画の grilling（2026-07-04）。ADR-0040 §1a を書き直した

## 背景

`~/.config` 専用の manifest のように、config 全体が root の下の同じサブディレクトリを基点に
したい場合がある。これまでは全 entry の `target` に共通の接頭辞を書くしかなく、宣言が重複していた。
`--root` は実行時に一律に上書きするもので、絶対パス文字列は評価時に固定される root なので、
どちらも「マーカーを実行時に解決し、その下へ宣言的にずらす」ことは表せない。grilling で、この
要望は root より下の方向へずらす意図だと確かめた。[root は 3 つの marker で表す](three-root-markers.md)。絶対パス文字列でも指定できる。

## 決定

3 つの root マーカーは、単体で使えば従来どおりのマーカーで、`homeRoot { subdir = ".config"; }`
のように引数を適用すると `subdir` を持つマーカーを返す。単体での使い方は変えない。適用形の
引数の規則は [root マーカーの subdir 適用形は全 target への前置と等価な評価時の糖衣で、root を動かさない](../requirements/root-marker-subdir-sugar.md)、root が取る形は
[root は 3 マーカーと絶対パス文字列の union を取る](../requirements/root-marker-kinds.md) に置く。

## 根拠

「root をずらす」という利用者の頭の中の形と書き味が一致する（grilling で確定）。`__functor` を
使えば単体での使い方との互換を壊さず、root の型の変更も任意の `subdir` を 1 つ許すだけで済む。
`mkOutOfStoreSymlink` と同じ「関数でマーカーを作る」既存の形の延長でもある。

## 棄却した代替案

### `mkManifest` の引数（`targetPrefix`）にする

実装は同じくらい単純で fixed root にも同じように使えるが、「root をずらす」という形との一致で
マーカーの適用形を選んだ。意味は同じなので、必要になれば別名の API として足せる。

### 何もせず、`target` に接頭辞を書く

機能としては成り立つが、config 全体の共通の基点という宣言の意図が entries に散らばる。
