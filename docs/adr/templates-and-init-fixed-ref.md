---
id: "ADR-d15a216f-b5d6-4df9-9825-b23a341bec1d"
type: adr
name: "templates と layat init は固定 ref で導入する"
section: ci-release-docs
status: stable
sources:
  - id: "issue-3"
    resource: "https://github.com/yasunori0418/layat/issues/3"
    title: "#3"
origin: "Issue #3（CLI 化）と、/grill-me による実装前の横断検査（実装前残セマンティクスの第 7 巡）。ADR-0006「layat の露出と環境セットアップ」節・ADR-0007 §6・ADR-0025 §3 を書き直した"
justifies:
  - "REQ-6be1cbf1-6c6e-498b-8acb-7f4b80037169"
  - "REQ-cbd61281-64b0-4487-a4b7-ce76e70dc4f9"
  - "DSG-16373ec2-3496-4b12-b3b1-ef74e0435b58"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# templates と layat init は固定 ref で導入する

- ステータス: stable
- 日付: 2026-11-01
- 関連: GitHub Issue #3
- 起点: Issue #3（CLI 化）と、/grill-me による実装前の横断検査（実装前残セマンティクスの第 7 巡）。ADR-0006「layat の露出と環境セットアップ」節・ADR-0007 §6・ADR-0025 §3 を書き直した

## 背景

新しいプロジェクトや dotfiles で layat を使い始めるには、input の追加・manifest の公開・devShell の
配線を備えた starter の flake が要る。layat は「設定を生成しない」ことを方針にしており、layat 自身が
ファイルを書き出すとこの方針を損なう。

`layat init` でテンプレートを取得するには、layat のリポジトリを指す flake ref をどう決めるかも
定める必要があった。

## 決定

starter の flake は nix 標準の `templates` output（`standalone` / `project`）で提供する。`layat init
<template>` は `nix flake init -t github:yasunori0418/layat#<template>` をそのまま呼ぶラッパーとし、参照する
flake ref はバイナリにハードコードした固定の文字列にする。

ラッパーの規範は [layat init は nix flake init -t への透明なラッパーとしファイルを生成しない](../requirements/init-template-wrapper.md) に、固定 ref は
[init のテンプレート参照はバイナリにハードコードした固定 flake ref とする](../requirements/init-fixed-flake-ref.md) に、`templates` output の公開形は
[layat 自身の flake outputs は packages / templates / 各モジュール / flakeModules / lib の 5 系統で構成する](../design/flake-outputs-attrs.md) に置く。

## 根拠

ファイルを作るのは nix の templates の機構で、layat は 1 文字も生成しない。`nix flake init` の
「既存ファイルを上書きしない」保守的な振る舞いもそのまま引き継げる。`init` は `nix flake init -t` を呼ぶだけなので
実装も小さい。

固定 ref は最も小さな実装で、利用者の環境の登録に左右されない。`init` は新規の立ち上げに使うもので、
CLI の版とテンプレートの版のずれ（常に最新の main を参照する）は許容できる。適用時の `schemaVersion` の
一致は、project mode の devShell に同梱した pin 版の CLI が担う。

## 棄却した代替案

### layat がファイルを直接スキャフォルドする

テンプレートの機構に頼らずに済むが、layat がファイルを生成することになり「設定を生成しない」方針を
損なう。

### nix registry の `layat` エントリを参照する

ref が最も短いが、registry に登録していない環境では動かない。

### ビルド時に自身の ref を ldflags で埋め込む

`init` が CLI と同じ版のテンプレートを取得できるが、`buildGoModule` への ref の注入と、flake / legacy の
両方への対応が重い。版の一致は devShell の pin が担うので、そこまでは要らない。
