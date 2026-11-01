---
id: "ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3"
type: adr
name: "root の省略を evalModules の標準のエラーで止める"
status: 採用
origin: "manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §3（クロスエントリ制約と必須性）の root 必須の項を書き直した"
justifies:
  - "REQ-d1b5b3f5-10a0-400d-9f03-ba00c63d1c34"
references:
  - "ADR-0007"
---
# ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3: root の省略を evalModules の標準のエラーで止める

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-0007, `docs/spec.md`
- 起点: manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §3（クロスエントリ制約と必須性）の root 必須の項を書き直した

## 背景

`root` は暗黙の既定値を持たず、利用者が明示しなければならない（ADR-0007）。`mkManifest` は
`evalModules` で入力を検査するので、この必須性をどの仕組みで検査するかを決める必要があった。

`evalModules` は、既定値を持たないオプションが設定されないまま使われると、"option used but not
defined" で評価を止める。

## 決定

`root` のオプションには既定値を持たせず、省略されたときは `evalModules` の標準のエラーで止める。
必須性のための検査を別に書かない。

`mkManifest` の検査の規範は REQ-d1b5b3f5-10a0-400d-9f03-ba00c63d1c34 に置く。

## 根拠

モジュールシステムの標準の挙動だけで、省略を評価時に止められる。同じことを手書きの検査で
重ねる理由が無い。

## 棄却した代替案

### `root` の省略を `lib.throwIf` で別に検査する

`evalModules` が既に止める誤りを二重に検査することになる。
