---
id: "TC-f0958a4c-b585-49c5-96a9-d7beff571e71"
type: test_condition
name: "記録は版 2 で出所を持ち、読みは版 1 と 2 を受け付け、出所の食い違いを警告し、書き直しは出所を引き継ぐ"
status: stable
mitigates:
  - "RISK-94be367e-9e2b-43ed-a351-6f16d41113b8"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:53:33+09:00"
---
# TC-f0958a4c-b585-49c5-96a9-d7beff571e71: 記録は版 2 で出所を持ち、読みは版 1 と 2 を受け付け、出所の食い違いを警告し、書き直しは出所を引き継ぐ

## テスト条件

last-applied 記録の出所（`origin`）の書き・読み・比較・引き継ぎを検証する。

| テーマ | 期待 |
|---|---|
| 版 2 の書き | apply は通常経路と generation skip 経路のどちらでも `recordVersion: 2` と今回の `origin` で書く。nix 生成器は `generator: "nix"` と entrypoint の絶対パス、prebuilt 生成器は `generator: "prebuilt"` だけを持つ。manifest が `identity` を持てば `origin.identity` にその値を書き、持たなければ省く |
| 版 1 と 2 の読み | 版 1 と版 2 の記録はどちらも読める。版 1 は `origin` 無しとして読む。それ以外の版と未知のフィールドは error。`origin` がオブジェクトでも `null` でもなければ error で、`null` は `origin` 無しとして読む。`origin` の中の `generator` の欠落・各フィールドの型違い・未知のフィールドも error。それ以外の値は読んだまま比較に使う |
| 出所の引き継ぎ | reset / rollback の書き直しは記録済みの `origin` を残す。記録に `origin` が無いとき（記録が無く profile から比較の基準を取ったときを含む）は、`origin` を省いて版 2 で書く |
| 食い違い時の警告 | 記録と今回の両方に `identity` があれば `identity` だけで比べ、一致なら `generator` / `entrypoint` が違っても警告せず、不一致なら同じでも警告する。どちらかに無ければ、`generator` で違うとき、`entrypoint` だけが違うときのどちらでも警告する。警告する apply（`--dryrun` と generation skip 経路を含む）は stderr に 1 行と `W_LAYAT_ORIGIN_MISMATCH`（`detail = {recorded, current}`。非空の `identity` は文面と各 origin に含む）を出し、配置は続行して後勝ちになる。一致すれば警告しない |
| `origin` の無い記録の比較スキップ | 記録に `origin` が無ければ（版 1、`origin` を省いた版 2、`origin: null` の版 2）比べず、警告を出さない |

## 対応する CASE

[`record_test.go` — last-applied 記録の読み書き・比較の基準の優先順位・各操作での更新](record-test-go.md)（`internal/engine/record_test.go`）、[e2e 02-home — 実 nix での home mode 世代コミットと rollback / reset の往復](e2e-02-home.md)（`tests/e2e/scenarios/02-home.sh`）。
