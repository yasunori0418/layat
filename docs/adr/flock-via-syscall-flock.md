---
id: "ADR-a0bac33b-d195-4dfd-9e74-61673c655873"
type: adr
name: "flock は syscall.Flock で実装する"
status: stable
origin: "engine / CLI の技術スタックの確定（2026-06-13）。ADR-0011 §6 を書き直した"
justifies:
  - "REQ-b74a118a-1272-44eb-944c-7725163211c6"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-01T23:01:53+09:00"
---
# flock は syscall.Flock で実装する

- ステータス: stable
- 日付: 2026-11-01
- 起点: engine / CLI の技術スタックの確定（2026-06-13）。ADR-0011 §6 を書き直した

## 背景

engine は profileDir ごとにロックを取る。取得できるまで待つ取り方と、取れなければすぐに諦める取り方の
両方が要る。engine は第三者の依存を持たず、標準ライブラリだけで書く。

ロックを持ったプロセスがクラッシュしても、ロックが残って以降の適用を止め続けてはならない。

## 決定

profileDir のロックは、Go 標準ライブラリの `syscall.Flock` による advisory lock で実装する。待つ取り方は
排他ロック（`LOCK_EX`）、待たない取り方は排他ロックに非ブロックのフラグ（`LOCK_NB`）を加えて表す。

engine が標準ライブラリだけに依存することは [engine は第三者依存ゼロの stdlib-only とし内部層に閉じる](../requirements/engine-stdlib-only.md) に置く。

## 根拠

flock のロックはプロセスの終了時に OS が解放するので、クラッシュしても古いロックが残らない。Linux と
macOS の両方で使え、標準ライブラリだけで書けるので、engine の依存を増やさない。待つ取り方と待たない
取り方を、同じ呼び出しのフラグの違いで表せる。

## 棄却した代替案

### `O_EXCL` で作るロックファイル

クラッシュするとロックファイルが残る。PID の生存確認などの回収処理を自前で持たなければならない。
