---
id: "TC-4b4709c9-1b7a-42f0-8430-62a26ead4eff"
type: test_condition
name: "engine 公開 API から見たロックの取得モードと保持区間が仕様どおりである"
status: stable
mitigates:
  - "RISK-2b17fefb-6e92-4513-9e7c-de21897c9cfe"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:31+09:00"
---
# engine 公開 API から見たロックの取得モードと保持区間が仕様どおりである

## テスト条件

`Apply` / `Rollback` / `Reset` を外側から駆動し、`profileDir` の flock がどう取られ、
いつ解放されるかを観測する。内部ヘルパーには結合せず、公開 API 越しのブラックボックスで
条件を組む。

**取得モード** — `NoWait`（`shellHook` から起動する場面）の `Apply` は、ロック保持中に実行すると
`ErrSkipped` でスキップし、何も配置しない。`Rollback` / `Reset` は既定の blocking で、
ロック保持中は先へ進まず、解放後に処理を完了する。

**保持区間** — ロックは操作の全区間で保持され、完了時に解放される。commit をブロックして
`Apply` をロック内で停留させ、その間に実行した `NoWait` の `Apply` がスキップすること
（操作中は保持）と、停留を解いて返ったあとの `NoWait` の `Apply` が成功すること
（完了で解放）の 2 点で観測する。

**sentinel の同一性** — `ErrSkipped` は `fmt.Errorf("%w")` のラップ（多重を含む）越しでも
`errors.Is` で到達できる。エラーへの文脈追加が sentinel の識別を切らないことの watchdog。
無関係なエラーをラップしたものが `ErrSkipped` に化けないことも併せて見る。

進行の同期は channel で決定的に取る。「保持中は進まない」だけは短い timeout 付きの
select で見るが、これは取りこぼしても偽陽性（block に気付かない）にしかならず、
flaky な失敗にはならない。

上位の規範は [原子性は実 FS の条件で故障を誘発して不変条件ごとに検証する](../../test-plan/fault-injection-atomicity.md) の「並行実行に対する原子性」で、競合を保持状態の模擬ではなく
実際にロックを保持して作ることもそこが定める。

## 対応する CASE

[engine/lock_test.go — 公開 API 越しのロック取得モード・保持区間・sentinel 同一性](engine-lock-test-go.md)（`internal/engine/lock_test.go`）。
