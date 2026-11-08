---
id: "ADR-3e283fe2-0177-4caf-8b60-577c5ee3b10a"
type: adr
name: "entry の配置種別のフィールドを method と名付ける"
status: stable
origin: "実装前の最終レビューで詰めた残りの細目。ADR-0015 §6 を書き直した"
justifies:
  - "REQ-3e446ad9-a6f4-4229-b5c5-184754b0ef51"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T02:02:56+09:00"
---
# ADR-3e283fe2-0177-4caf-8b60-577c5ee3b10a: entry の配置種別のフィールドを method と名付ける

- ステータス: stable
- 日付: 2026-11-01
- 関連: `docs/spec.md`, [method](../ubiquitous-language/method.md)
- 起点: 実装前の最終レビューで詰めた残りの細目。ADR-0015 §6 を書き直した

## 背景

entry は配置の種別（`"symlink"` か `"copy"`）を 1 つのフィールドで選ぶ。このフィールドは当初
`mode` という名前だったが、`mode` は unix のファイルモード（`0644` など）と読み違えやすい。

## 決定

配置種別のフィールドは `method` と名付ける。Nix の entry・`manifest.json`・engine のどこでも同じ
名前を使う。旧名の `mode` は受け付けず、評価時のエラーにする。

## 根拠

`mode` はファイルのパーミッションを強く連想させ、配置種別の enum と紛らわしい。読み違えの原因は、
実装が無く `schemaVersion` も初期のうちに取り除くのが最も安い。[配置元の中のパスを全層で subpath と呼ぶ](subpath-naming-across-layers.md) の改名と同じ理屈。

entry は未知のキーを拒否するので、旧名を書いた設定は評価時のエラーになり、その場で気付ける
（[entry submodule は strict とし未知キーと旧名を評価時エラーにする](../requirements/entry-submodule-strict.md)）。

## 棄却した代替案

### `mode` のままにする

読み違えの原因が残る。名前を変える費用は、実装が進むほど大きくなる。
