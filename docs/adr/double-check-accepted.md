---
id: "ADR-f66cddc5-6b4f-40e4-a85a-ce1a85cd9a0c"
type: adr
name: "モジュールの使い方での二重の確認を許容する"
status: stable
origin: "manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §3（クロスエントリ制約と必須性）の二重検査の項を書き直した"
justifies:
  - "REQ-d1b5b3f5-10a0-400d-9f03-ba00c63d1c34"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:30:50+09:00"
---
# ADR-f66cddc5-6b4f-40e4-a85a-ce1a85cd9a0c: モジュールの使い方での二重の確認を許容する

- ステータス: stable
- 日付: 2026-11-01
- 関連: `docs/spec.md`
- 起点: manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §3（クロスエントリ制約と必須性）の二重検査の項を書き直した

## 背景

モジュールの使い方では、entries オプション（`layat.configs.<name>.entries`）がまずホスト
（home-manager / NixOS）の `evalModules` で確認され、その後 `mkManifest` の `evalModules` でもう一度
確認される。同じ entry を 2 回確認することになる。

## 決定

モジュールの使い方での二重の確認は許容する。モジュールの使い方のときだけ `mkManifest` の確認を省く
分岐は作らない。

entry の確認を `mkManifest` に集める規範は [mkManifest 自身が evalModules で入力を確認する単一の箇所になる](../requirements/mkmanifest-input-validation.md) に置く。

## 根拠

entry を確認・正規化する場所を `mkManifest` の 1 か所に保てる。使い方によって確認の有無が
変わらない。確認は純粋で冪等なので、2 回目の評価の費用は無視できる。

## 棄却した代替案

### モジュールの使い方では `mkManifest` の確認を省く

使い方ごとに確認の有無が分かれ、確認の場所が `mkManifest` の 1 か所でなくなる。省いて得られる評価の費用は
わずかである。
