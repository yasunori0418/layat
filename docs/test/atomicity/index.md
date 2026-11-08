# atomicity

## テスト条件（`test_condition`）

* [--backup の退避が opt-in で発動し、commit 後は残り、途中失敗では戻る](./backup-aside-lifecycle.md) - `apply --backup` の rename 退避が、発動条件・退避先の命名・run 終了後の扱いのそれぞれで契約どおりに振る舞うことを検証する。
* [退避先が既存なら停止し、rename 直前に退避先を再検証する](./backup-conflict-and-toctou.md) - `--backup` が退避先 `<target>.<suffix>` を上書きしないことを、plan 時点と execute 時点の両方で検証する。
* [commit 段の失敗では配置を巻き戻さず、退避ファイルも掃除しない](./commit-failure-not-unwound.md) - [実 FS の故障注入で FS 変更段の途中を失敗させ、apply 前の状態が完全に復元される](mid-batch-rollback-restores-preapply.md) の裏返し。commit（`nix-env --set`）でのみ失敗させ、**巻き戻しが起きない**ことを検証する。commit へ到達した時点でこの run の FS 書き込みは全て成功しており、FS は既に整合している。ここで unwind すればその整合状態を壊すため、巻き戻さないのが正しい（→ [undo の発火を FS 反映の段に限り、世代コミットの後は対象にしない](../../adr/undo-scope-ends-at-commit.md)）。世代が進まないだけで、冪等な再実行が収束する。
* [engine 公開 API から見たロックの取得モードと保持区間が仕様どおりである](./lock-mode-semantics.md) - `Apply` / `Rollback` / `Reset` を外側から駆動し、`profileDir` の flock がどう取られ、いつ解放されるかを観測する。内部ヘルパーには結合せず、公開 API 越しのブラックボックスで条件を組む。
* [lock プリミティブが try で競合を弾き、blocking で待ち、解放後に再取得できる](./lock-primitive-behavior.md) - engine から切り離した `internal/lock` パッケージ単体で、flock ラッパーの 3 つの基本振る舞いを検証する。engine 側の観測（[engine 公開 API から見たロックの取得モードと保持区間が仕様どおりである](lock-mode-semantics.md)）が仕様レベルの契約を見るのに対し、こちらは下層のプリミティブが正しいことを固定する。
* [実 FS の故障注入で FS 変更段の途中を失敗させ、apply 前の状態が完全に復元される](./mid-batch-rollback-restores-preapply.md) - `Apply` を end-to-end で実行し、FS を変更する各段の途中で失敗させたときに、その run が加えた変更が全て巻き戻り、apply 前の状態へ戻ることを検証する。故障は実 FS の条件で誘発する（→ [原子性は実 FS の条件で故障を誘発して不変条件ごとに検証する](../../test-plan/fault-injection-atomicity.md)）。
* [undo の逆操作が、対応する前方操作の直前状態を単体で復元する](./undo-inverse-op-correctness.md) - undo ジャーナルの各エントリ種別に対応する逆操作を、`Apply` を通さず個別に駆動し、前方操作の直前状態が復元されることを検証する。
* [unwind が LIFO で巻き戻し、1 件の失敗でも継続して全件報告し、成功時は journal を破棄する](./unwind-lifo-and-journal-lifecycle.md) - 個々の逆操作の正しさ（[undo の逆操作が、対応する前方操作の直前状態を単体で復元する](undo-inverse-op-correctness.md)）ではなく、journal 全体を扱う層の振る舞いを検証する。

## テストケース（`test_case`）

* [backup_test.go — --backup の退避ライフサイクルと退避先の保護](./backup-test-go.md) - `internal/engine/backup_test.go`
* [engine/lock_test.go — 公開 API 越しのロック取得モード・保持区間・sentinel 同一性](./engine-lock-test-go.md) - `internal/engine/lock_test.go`
* [internal/lock/lock_test.go — flock ラッパーの try / blocking / 再取得](./internal-lock-test-go.md) - `internal/lock/lock_test.go`
* [undo_journal_test.go — 実 FS 故障注入による end-to-end 巻き戻しと commit 失敗の非対称性](./undo-journal-test-go.md) - `internal/engine/undo_journal_test.go`
* [undo_test.go — 逆操作単体と unwind の順序・best-effort・journal 破棄](./undo-test-go.md) - `internal/engine/undo_test.go`
