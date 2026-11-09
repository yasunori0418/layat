---
id: "TC-5454b99c-f274-438a-9ba9-5be09aa71c98"
type: test_condition
name: "lock プリミティブが try で競合を弾き、blocking で待ち、解放後に再取得できる"
status: stable
mitigates:
  - "RISK-2b17fefb-6e92-4513-9e7c-de21897c9cfe"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:31+09:00"
---
# lock プリミティブが try で競合を弾き、blocking で待ち、解放後に再取得できる

## テスト条件

engine から切り離した `internal/lock` パッケージ単体で、flock ラッパーの 3 つの基本
振る舞いを検証する。engine 側の観測（[engine 公開 API から見たロックの取得モードと保持区間が仕様どおりである](lock-mode-semantics.md)）が仕様レベルの契約を見るのに対し、
こちらは下層のプリミティブが正しいことを固定する。

- **try-lock の競合** — 保持中に try-lock すると `ErrLocked` を返す
- **解放後の再取得** — 解放したロックは同じキーで再取得できる（fd / ファイルの後始末が
  次の取得を妨げない）
- **blocking の待機** — blocking 取得は保持中に取得へ進まず、解放されてから取得する

上位の規範は [engine 公開 API から見たロックの取得モードと保持区間が仕様どおりである](lock-mode-semantics.md) と同じく [原子性は実 FS の条件で故障を誘発して不変条件ごとに検証する](../../test-plan/fault-injection-atomicity.md) の「並行実行に対する原子性」。

## 対応する CASE

[internal/lock/lock_test.go — flock ラッパーの try / blocking / 再取得](internal-lock-test-go.md)（`internal/lock/lock_test.go`）。
