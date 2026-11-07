---
id: "ADR-97edd9af-befb-4a9f-a8ef-de038c76336a"
type: adr
name: "layat は buildGoModule と vendorHash でビルドする"
status: stable
issues:
  - "#552"
origin: "engine / CLI の技術スタックの確定（2026-06-13）。ADR-0011 §5 を書き直した"
justifies:
  - "REQ-637599dc-a1ec-4af5-9e97-e882c7df56d0"
  - "DSG-16373ec2-3496-4b12-b3b1-ef74e0435b58"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-01T23:01:38+09:00"
---
# ADR-97edd9af-befb-4a9f-a8ef-de038c76336a: layat は buildGoModule と vendorHash でビルドする

- ステータス: stable
- 日付: 2026-11-01
- 関連: GitHub Issue #552
- 起点: engine / CLI の技術スタックの確定（2026-06-13）。ADR-0011 §5 を書き直した

## 背景

layat の CLI と engine は 1 つの Go バイナリで、cobra などの第三者依存を持つ。Nix のサンドボックスは
ネットワークを遮断するので、依存を固定して取り込む方法が要る。

## 決定

`packages.<system>.layat` は nixpkgs の `buildGoModule` でビルドし、依存を `vendorHash` の文字列で
固定する。依存を変えたときは hash を更新する。

依存の固定の規範は REQ-637599dc-a1ec-4af5-9e97-e882c7df56d0 に、output の形は
DSG-16373ec2-3496-4b12-b3b1-ef74e0435b58 に置く。

## 根拠

flake input を増やさずに、nixpkgs の標準のビルド方法をそのまま使える。今の依存の規模なら、hash を更新する負荷も軽い。

## 棄却した代替案

### gomod2nix を使う

flake input が増え、`gomod2nix.toml` を再生成する手間も残る。細かい単位でキャッシュできる利点は、
今の依存の規模では小さい。

### `vendor/` をコミットし `vendorHash = null` にする

hash の更新は不要になるが、vendor の差分がリポジトリに載る。
