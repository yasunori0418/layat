---
id: "CASE-6f6fabaa-ae1e-4af5-962e-087a9064f85d"
type: test_case
name: "engine/lock_test.go — 公開 API 越しのロック取得モード・保持区間・sentinel 同一性"
status: stable
resource: "internal/engine/lock_test.go"
covers:
  - "TC-4b4709c9-1b7a-42f0-8430-62a26ead4eff"
follows:
  - "TP-e7c25263-6d2d-4a37-8275-26906889d912"
  - "TP-deb05610-44bc-4962-8939-952392e5fbd0"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-14T01:44:52+09:00"
---
# engine/lock_test.go — 公開 API 越しのロック取得モード・保持区間・sentinel 同一性

## 対象

`internal/engine/lock_test.go`

`Apply` / `Rollback` / `Reset` をブラックボックスで駆動し、`profileDir` の flock 意味論を
観測する統合テスト（→ [flock は明示実行で待ち、hook 実行でスキップする](../../adr/flock-mode-explicit-wait-hook-skip.md)、[実行フローの順序は eval 先行 → flock → build とし build をロック内に閉じる](../../requirements/exec-flow-eval-lock-build.md)）。内部ヘルパーには結合しない。
進行の同期は channel で決定的に取り、`time.Sleep` によるタイミング依存を持たない。
「保持中は進まない」の確認だけが短い timeout 付き select（取りこぼしても偽陽性にしか
ならず flaky にならない）。

## 検証内容

- **NoWait の try-lock** — `profileDir` の flock を別 fd から保持した状態で `NoWait` の
  `Apply` を実行すると `ErrSkipped` でスキップし、何も配置しない。判定は `==` ではなく
  `errors.Is`（将来の文脈ラップ #90 に耐える形）
- **`ErrSkipped` の透過性** — `fmt.Errorf("%w")` の単段・多段ラップ越しでも `errors.Is`
  で到達する。無関係なエラーのラップは `ErrSkipped` に化けない
- **保持区間** — commit をブロックして `Apply` をロック内に停留させ、その間の並行 `NoWait`
  `Apply` がスキップすること（操作中は保持）、停留解除後の `NoWait` `Apply` が成功する
  こと（完了で解放）を順序固定で観測する
- **`Reset` の blocking** — ロック保持中は進まず、解放後に完了する
- **`Rollback` の blocking** — ロック保持中は待ち、解放後に前世代（gen1 = {a}）へ
  再収束して profile ポインタを移す
