---
id: "RISK-2b17fefb-6e92-4513-9e7c-de21897c9cfe"
type: risk
name: "flock による直列化が破れ、同一 config への同時実行が互いの配置を壊す"
section: atomicity
status: stable
threatens:
  - "REQ-1c1526b1-59e3-4264-bb7c-65a10a4aa461"
  - "REQ-60c6b7ea-e936-4ce8-bd75-ad35e9c693b9"
  - "REQ-a8edc58f-4adc-4637-b888-ab8ccc7e73e4"
likelihood: medium
impact: high
level: high
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:36:11+09:00"
---
# flock による直列化が破れ、同一 config への同時実行が互いの配置を壊す

## リスク

同一 `profileDir` に対する apply / rollback / reset は flock で直列化される
（→ [実行フローの順序は eval 先行 → flock → build とし build をロック内に閉じる](../requirements/exec-flow-eval-lock-build.md) (2)(a)）。この直列化が破れると、2 つの run が同じ target 集合へ同時に
書き込む。片方の古くなった配置の除去がもう片方の配置直後の symlink を消す、片方の undo journal が
もう片方の書き込みを「自分が置いたもの」として巻き戻す、といった形で、どちらの run の
巻き戻し保証も成立しなくなる。原子性の他のリスクが「1 run の中の失敗」を扱うのに対し、
本リスクはその前提である「1 config を触るのは同時に 1 run だけ」の側が崩れる。

破れ方は取得モードの取り違えで起きる。既定は blocking（`LOCK_EX`）で待ち、`--no-wait`
のときだけ try-lock（`LOCK_NB`）でスキップする（→ [flock の取得は既定 blocking とし --no-wait のときだけ try-lock でスキップする](../requirements/flock-acquisition-mode.md)）。blocking のはずの処理が
try-lock になっていれば競合時に利用者に知らせずに何もせず「成功した」と見え、try-lock のはずの
`shellHook` の処理が blocking になっていればシェル入室がロック保持者の完了までブロックされる。
後者は破壊ではないが、devShell / direnv 運用を実質的に壊す。

保持区間の誤りも同じ処理の中で起きる。ロックは操作の全区間で保持され完了時に解放される必要が
あり、早すぎる解放は上の同時実行を許し、解放し忘れは以後の全 run を永久にブロックする。
スキップを表す `ErrSkipped` が `fmt.Errorf("%w")` のラップ越しに `errors.Is` で届かなく
なるのも、呼び出し側がスキップを失敗と誤認する形で本リスクに含む。

`reset` も解決した profileDir を鍵に blocking flock を取る（→ [reset は名指し必須で profileDir 単位の blocking flock を取る](../requirements/reset-named-only-and-flock.md)）。
reset が lock を取らないか別の鍵で取ると、同時に実行される apply と互いの配置を壊す。

## 実現性

**likelihood: medium** — 競合は devShell の `shellHook` が明示 apply と重なる、複数の
シェルを同時に開く、といった日常操作で成立する。

**impact: high** — 同時書き込みの結果は配置の破壊であり、巻き戻しでも戻せない。

## 緩和

[engine 公開 API から見たロックの取得モードと保持区間が仕様どおりである](../test/atomicity/lock-mode-semantics.md)（engine 公開 API から見たロック取得モードと保持区間の観測）と
[lock プリミティブが try で競合を弾き、blocking で待ち、解放後に再取得できる](../test/atomicity/lock-primitive-behavior.md)（`internal/lock` プリミティブ単体の try / blocking / 再取得）が緩和する。
