---
id: "ADR-83a1275b-85bb-4918-8d21-8db9ab18a58c"
type: adr
name: "manifest.json へ marker を clean enum で写す"
status: 採用
origin: "manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §5（manifest.json ワイヤ形式）を書き直した"
justifies:
  - "REQ-1dcc9a33-b0f2-43e0-8310-fc4b19e68fe7"
---
# ADR-83a1275b-85bb-4918-8d21-8db9ab18a58c: manifest.json へ marker を clean enum で写す

- ステータス: 採用
- 日付: 2026-11-01
- 関連: `docs/spec.md`
- 起点: manifest の型検査を NixOS モジュールシステムで行えるかという確認。ADR-0010 §5（manifest.json ワイヤ形式）を書き直した

## 背景

lib の型の確認は、`src` / `root` に渡された値が marker か derivation かを、marker に持たせた判別タグ
`_layatMarker` で見分ける。一方 engine が読む `manifest.json` にも、配置元の種別と root の種別を
記録する必要がある。判別タグをそのまま契約に載せるかどうかを決める必要があった。

## 決定

`_layatMarker` は Nix の評価の中で使い終え、`manifest.json` には出さない。engine へは `srcKind` /
`rootKind` の clean enum として写す。写す変換は lib の正規化（`normalizeManifest`）が受け持ち、
store 由来の `src` も Nix 側で store パスの文字列に解決してから書く。

marker と enum の対応は REQ-1dcc9a33-b0f2-43e0-8310-fc4b19e68fe7 に置く。

## 根拠

判別タグの名前や構造は Nix 側の型実装の都合で決まる。契約に漏らさなければ、型実装を変えても
engine との契約は変わらない。

engine は文字列を受け取るだけで済み、store パスを得るために `nix eval` を呼び直さなくてよい。

## 棄却した代替案

### `_layatMarker` をそのまま `manifest.json` に出す

engine の契約が Nix 側の型実装の詳細に結び付き、タグの名前や構造を変えるたびに契約も変えなければならない。
