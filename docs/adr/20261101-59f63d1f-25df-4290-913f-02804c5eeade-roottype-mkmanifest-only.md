---
id: "ADR-59f63d1f-25df-4290-913f-02804c5eeade"
type: adr
name: "root の型を mkManifest だけに持たせる"
status: 採用
origin: "manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §2（型定義）の rootType の項を書き直した"
justifies:
  - "REQ-d1b5b3f5-10a0-400d-9f03-ba00c63d1c34"
supersedes:
  - "ADR-0010"
references:
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
  - "ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75"
  - "ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c"
  - "ADR-2119433e-de69-4447-85a5-575d58246007"
---
# ADR-59f63d1f-25df-4290-913f-02804c5eeade: root の型を mkManifest だけに持たせる

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-fe035038-4415-4a1e-aa5c-509fdfce9955, ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75, ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c, ADR-2119433e-de69-4447-85a5-575d58246007, `docs/spec.md`
- 起点: manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §2（型定義）の rootType の項を書き直した

## 背景

entry の型は `lib/types.nix` に置き、`mkManifest` とモジュールの両方で共有する。`root` の型
（固定の絶対パスの文字列か、root のマーカー）も同じように共有するかを決める必要があった。

モジュールは root を自分で決めて固定する（home-manager は home、NixOS は system）。利用者は
モジュール経由では root を選ばない。

## 決定

`root` の型は `mkManifest` の検査だけに持たせ、モジュールとは共有しない。

`mkManifest` の検査と型の共有範囲の規範は REQ-d1b5b3f5-10a0-400d-9f03-ba00c63d1c34 に置く。

## 根拠

モジュールは root を固定するので、利用者に root を書かせるオプションも、その型も要らない。
共有しても使われない型が増えるだけになる。

## 棄却した代替案

### root の型も entry の型と同じくモジュールと共有する

モジュールは root を固定するので、共有した型を使う場面が無い。
