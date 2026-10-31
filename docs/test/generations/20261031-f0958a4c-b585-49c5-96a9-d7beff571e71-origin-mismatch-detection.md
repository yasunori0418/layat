---
id: "TC-f0958a4c-b585-49c5-96a9-d7beff571e71"
type: test_condition
name: "記録は版 2 で出所を持ち、読みは版 1 と 2 を受け付け、出所の食い違いを警告し、書き直しは出所を引き継ぐ"
mitigates:
  - "RISK-94be367e-9e2b-43ed-a351-6f16d41113b8"
---
# TC-f0958a4c-b585-49c5-96a9-d7beff571e71: 記録は版 2 で出所を持ち、読みは版 1 と 2 を受け付け、出所の食い違いを警告し、書き直しは出所を引き継ぐ

## テスト条件

last-applied 記録の出所（`origin`）の書き・読み・比較・引き継ぎを検証する。

| テーマ | 期待 |
|---|---|
| 版 2 の書き | apply は `recordVersion: 2` と今回の `origin` で書く。nix 生成器は `generator: "nix"` と entrypoint の絶対パス、prebuilt 生成器は `generator: "prebuilt"` だけを持つ |
| 版 1 と 2 の読み | 版 1 と版 2 の記録はどちらも読める。版 1 は `origin` 無しとして読む。それ以外の版と未知のフィールド（`origin` の中を含む）は error |
| 出所の引き継ぎ | reset / rollback の書き直しは記録済みの `origin` を残す。記録に `origin` が無ければ省いたまま版 2 で書く |
| 食い違い時の警告 | 記録と今回の `origin` が `generator` で違うとき、`entrypoint` だけが違うときのどちらでも、apply（`--dryrun` と generation skip 経路を含む）は stderr に 1 行と `W_LAYAT_ORIGIN_MISMATCH`（`detail = {recorded, current}`）を出し、配置は続行して後勝ちになる。一致すれば警告しない |
| `origin` の無い記録の比較スキップ | 記録に `origin` が無ければ（版 1、または `origin` を省いた版 2）比べず、警告を出さない |

## 対応する CASE

未起票。
