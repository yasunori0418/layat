---
id: "ADR-d9cad6f6-d694-4c1d-afee-b57b129955dc"
type: adr
name: "marker を判別タグで derivation と見分ける"
status: stable
origin: "manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §2（型定義）の marker の項を書き直した"
justifies:
  - "REQ-1dcc9a33-b0f2-43e0-8310-fc4b19e68fe7"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# ADR-d9cad6f6-d694-4c1d-afee-b57b129955dc: marker を判別タグで derivation と見分ける

- ステータス: stable
- 日付: 2026-11-01
- 関連: `docs/spec.md`
- 起点: manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §2（型定義）の marker の項を書き直した

## 背景

`src` と `root` には、layat の関数が返す marker（out-of-store のマーカーや root のマーカー）を
渡せる。`src` には derivation（`fetchFromGitHub` の戻り値など）も渡せる。marker も derivation も
attrset なので、そのままでは型の確認で区別できない。

## 決定

marker に判別タグ `_layatMarker` を持たせ、独自の optionType の `check` でタグを見て marker と
derivation を見分ける。

規範は [marker は判別タグで識別し manifest.json には漏らさない](../requirements/marker-tag-discrimination.md) に置く。

## 根拠

タグがあれば、attrset どうしでも型の確認の段で確実に判別できる。タグの種類は、そのまま
`manifest.json` の `srcKind` / `rootKind` を決める材料にもなる。

## 棄却した代替案

### marker をタグの無い素の attrset にする

derivation と marker が区別できず、`src` の型の判別が成り立たない。
