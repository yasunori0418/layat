---
id: "ADR-2be5080a-1cba-4fea-bed8-66e0d0d7b8d0"
type: adr
name: "Go toolchain は nixpkgs の go に pin する"
status: stable
sources:
  - id: "issue-552"
    resource: "https://github.com/yasunori0418/layat/issues/552"
    title: "#552"
origin: "engine / CLI の技術スタックの確定（2026-06-13）。ADR-0011 §8 を書き直した"
justifies:
  - "REQ-637599dc-a1ec-4af5-9e97-e882c7df56d0"
references:
  - "ADR-97edd9af-befb-4a9f-a8ef-de038c76336a"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-01T23:01:38+09:00"
---
# ADR-2be5080a-1cba-4fea-bed8-66e0d0d7b8d0: Go toolchain は nixpkgs の go に pin する

- ステータス: stable
- 日付: 2026-11-01
- 関連: [layat は buildGoModule と vendorHash でビルドする](buildgomodule-vendorhash.md), GitHub Issue #552
- 起点: engine / CLI の技術スタックの確定（2026-06-13）。ADR-0011 §8 を書き直した

## 背景

Go のビルドで使う toolchain は、nixpkgs が提供する go と、`go.mod` の `toolchain` ディレクティブの
どちらでも指定できる。Nix のサンドボックスはネットワークを遮断するので、toolchain の自動ダウンロードは
失敗する。

## 決定

Go toolchain は、`buildGoModule` が使う nixpkgs の go に pin する。`go.mod` には `go 1.xx` の
言語バージョンだけを書き、`toolchain` ディレクティブは置かない。nixpkgs を更新すれば go も追従する。

toolchain の規範は [CLI が持ち込む依存は許可した第三者ライブラリと pin した Go に限り、いずれも固定する](../requirements/cli-dependency-policy.md) に置く。

## 根拠

版の正を nixpkgs の go の 1 箇所に置けば、二重管理が起きず、ビルドが独自に toolchain を取りに行くことも
無い。

## 棄却した代替案

### `go.mod` の `toolchain` を正にする

nixpkgs の go と食い違うと、サンドボックス内でダウンロードを試みてビルドが失敗する。版を二重に管理する
ことにもなる。
