---
id: "ADR-12ad7852-e689-44d7-b9fe-0667effb3d13"
type: adr
name: "entry の未知のキーを評価時に拒否する"
status: 採用
origin: "manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §2（型定義）の strict の項を書き直した"
justifies:
  - "REQ-3e446ad9-a6f4-4229-b5c5-184754b0ef51"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:41+09:00"
---
# ADR-12ad7852-e689-44d7-b9fe-0667effb3d13: entry の未知のキーを評価時に拒否する

- ステータス: 採用
- 日付: 2026-11-01
- 関連: `docs/spec.md`
- 起点: manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §2（型定義）の strict の項を書き直した

## 背景

entry は `lib.types.submodule` で定義する。submodule は、指定が無いときは宣言していないキーをエラーにする。
`freeformType` を付ければ、宣言していないキーも受け付けるように緩められる。

entry のフィールドは改名を重ねており、旧名（`source` や `dir` など）を書いた設定が残り得る。

## 決定

entry は submodule の既定どおり strict にし、宣言していないキーを評価時のエラーにする。
`freeformType` で緩めない。

規範は REQ-3e446ad9-a6f4-4229-b5c5-184754b0ef51 に置く。

## 根拠

タイポや旧名のキーがエラーにならずに無視されると、利用者は設定が反映されていない理由に気付けない。評価時に
止めれば、誤りを書いたその場で分かる。「利用者が配置を明示的に握る」という layat の考え方とも合う。

## 棄却した代替案

### `freeformType` で未知のキーを受け付ける

タイポや旧名がエラーにならずに無視され、設定が反映されていないことに気付けない。
