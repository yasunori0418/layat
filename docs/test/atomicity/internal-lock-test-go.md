---
id: "CASE-2e675eea-8e0e-41f6-bb94-5f2ce944255b"
type: test_case
name: "internal/lock/lock_test.go — flock ラッパーの try / blocking / 再取得"
status: stable
resource: "internal/lock/lock_test.go"
covers:
  - "TC-5454b99c-f274-438a-9ba9-5be09aa71c98"
follows:
  - "TP-e7c25263-6d2d-4a37-8275-26906889d912"
  - "TP-deb05610-44bc-4962-8939-952392e5fbd0"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-08-18T22:23:49+09:00"
---
# internal/lock/lock_test.go — flock ラッパーの try / blocking / 再取得

## 対象

`internal/lock/lock_test.go`

engine から切り離した flock ラッパー単体のユニットテスト。

## 検証内容

- **try-lock の競合** — ロック保持中の try-lock が `ErrLocked` を返す
- **解放後の再取得** — 解放したロックを同じキーで再取得できる
- **blocking の待機** — blocking 取得は保持中に取得へ進まず、解放後に取得する
