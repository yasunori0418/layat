---
id: "TC-83fe0d4a-9730-431a-b7f9-b731060b9484"
type: test_condition
name: "unwind が LIFO で巻き戻し、1 件の失敗でも継続して全件報告し、成功時は journal を破棄する"
status: stable
mitigates:
  - "RISK-68e810c5-4e68-4b25-9bc0-6b2613022b49"
  - "RISK-a1622cdd-c7e2-4178-a50d-85bc2a35b29f"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:31+09:00"
---
# TC-83fe0d4a-9730-431a-b7f9-b731060b9484: unwind が LIFO で巻き戻し、1 件の失敗でも継続して全件報告し、成功時は journal を破棄する

## テスト条件

個々の逆操作の正しさ（[undo の逆操作が、対応する前方操作の直前状態を単体で復元する](undo-inverse-op-correctness.md)）ではなく、journal 全体を扱う層の振る舞いを検証する。

**LIFO 順序** — 複数エントリを積んだ journal は last-in-first-out で巻き戻される。順序が
結果を左右するのは前方操作に依存関係があるときで、代表形は PreRemove の「子 symlink を unlink →
空になった親を rmdir」という bottom-up のバッチ。この LIFO は親の mkdir を先に、子の
再作成を後にする top-down の undo になり、これ以外の順序では子を作る先が存在しない。
`preRemove` を直接駆動して、journal がこの順序で記録されること自体も条件に含む。

**best-effort 継続と全件報告** — 1 件の逆操作が失敗しても残りの巻き戻しは続行し、元の
エラーと復旧できなかった項目の双方が報告される。journal が空（FS 書き込み前の失敗）の
ときは元のエラーだけを報告し、「N 件」の行を出さない。

**成功時の journal 破棄** — commit 成功後の discardJournal が `--recopy` の rename 退避
ファイルを掃除する。undo は起きなかったのだから、新しい copy がそのまま残る。
`--backup` の退避を掃除しない側は [--backup の退避が opt-in で発動し、commit 後は残り、途中失敗では戻る](backup-aside-lifecycle.md) が担当する。

## 対応する CASE

[undo_test.go — 逆操作単体と unwind の順序・best-effort・journal 破棄](undo-test-go.md)（`internal/engine/undo_test.go`）・[undo_journal_test.go — 実 FS 故障注入による end-to-end 巻き戻しと commit 失敗の非対称性](undo-journal-test-go.md)
（`internal/engine/undo_journal_test.go` の `preRemove` 直接駆動）。
