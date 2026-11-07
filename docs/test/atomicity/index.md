# atomicity

## テスト条件（`test_condition`）

* [実 FS の故障注入で FS 変更段の途中を失敗させ、apply 前の状態が完全に復元される](./20260809-3b02ab58-5bfb-4ae7-9ac1-c69e2ece2722-mid-batch-rollback-restores-preapply.md) - `Apply` を end-to-end で実行し、FS を変更する各段の途中で失敗させたときに、その run が 加えた変更が全て巻き戻り、apply 前の状態へ戻ることを検証する。故障は実 FS の条件で 誘発する（→ TP-deb05610-44bc-4962-8939-952392e5fbd0）。
* [engine 公開 API から見たロックの取得モードと保持区間が仕様どおりである](./20260809-4b4709c9-1b7a-42f0-8430-62a26ead4eff-lock-mode-semantics.md) - `Apply` / `Rollback` / `Reset` を外側から駆動し、`profileDir` の flock がどう取られ、 いつ解放されるかを観測する。内部ヘルパーには結合せず、公開 API 越しのブラックボックスで 条件を組む。
* [lock プリミティブが try で競合を弾き、blocking で待ち、解放後に再取得できる](./20260809-5454b99c-f274-438a-9ba9-5be09aa71c98-lock-primitive-behavior.md) - engine から切り離した `internal/lock` パッケージ単体で、flock ラッパーの 3 つの基本 振る舞いを検証する。engine 側の観測（TC-4b4709c9-1b7a-42f0-8430-62a26ead4eff）が仕様レベルの契約を見るのに対し、 こちらは下層のプリミティブが正しいことを固定する。
* [unwind が LIFO で巻き戻し、1 件の失敗でも継続して全件報告し、成功時は journal を破棄する](./20260809-83fe0d4a-9730-431a-b7f9-b731060b9484-unwind-lifo-and-journal-lifecycle.md) - 個々の逆操作の正しさ（TC-9504e908-a4bc-4d2d-80d3-af07264284f8）ではなく、journal 全体を扱う層の振る舞いを検証する。
* [undo の逆操作が、対応する前方操作の直前状態を単体で復元する](./20260809-9504e908-a4bc-4d2d-80d3-af07264284f8-undo-inverse-op-correctness.md) - undo ジャーナルの各エントリ種別に対応する逆操作を、`Apply` を通さず個別に駆動し、 前方操作の直前状態が復元されることを検証する。
* [退避先が既存なら停止し、rename 直前に退避先を再検証する](./20260809-cd03ff56-6ed0-4100-9ec4-eb681ab66407-backup-conflict-and-toctou.md) - `--backup` が退避先 `<target>.<suffix>` を上書きしないことを、plan 時点と execute 時点の 両方で検証する。
* [commit 段の失敗では配置を巻き戻さず、退避ファイルも掃除しない](./20260809-e048211f-0844-4446-9957-fb444c1da4e5-commit-failure-not-unwound.md) - TC-3b02ab58-5bfb-4ae7-9ac1-c69e2ece2722 の裏返し。commit（`nix-env --set`）でのみ失敗させ、**巻き戻しが起きない** ことを検証する。commit へ到達した時点でこの run の FS 書き込みは全て成功しており、FS は 既に整合している。ここで unwind すればその整合状態を壊すため、巻き戻さないのが正しい （→ ADR-f29d1256-5e14-4b83-ac77-4717b2210193）。世代が進まないだけで、冪等な再実行が収束する。
* [--backup の退避が opt-in で発動し、commit 後は残り、途中失敗では戻る](./20260809-ed4992c0-8513-4383-be0a-e45acbbc229f-backup-aside-lifecycle.md) - `apply --backup` の rename 退避が、発動条件・退避先の命名・run 終了後の扱いのそれぞれで 契約どおりに振る舞うことを検証する。

## テストケース（`test_case`）

* [undo_test.go — 逆操作単体と unwind の順序・best-effort・journal 破棄](./20260809-02475ac2-5555-4fa1-a3e5-cda5015919c5-undo-test-go.md) - `internal/engine/undo_test.go`
* [undo_journal_test.go — 実 FS 故障注入による end-to-end 巻き戻しと commit 失敗の非対称性](./20260809-154af597-df3e-49fb-a96b-b4f371dfcc63-undo-journal-test-go.md) - `internal/engine/undo_journal_test.go`
* [internal/lock/lock_test.go — flock ラッパーの try / blocking / 再取得](./20260809-2e675eea-8e0e-41f6-bb94-5f2ce944255b-internal-lock-test-go.md) - `internal/lock/lock_test.go`
* [engine/lock_test.go — 公開 API 越しのロック取得モード・保持区間・sentinel 同一性](./20260809-6f6fabaa-ae1e-4af5-962e-087a9064f85d-engine-lock-test-go.md) - `internal/engine/lock_test.go`
* [backup_test.go — --backup の退避ライフサイクルと退避先の保護](./20260809-ed4b32fd-4f2e-497c-baa4-cc91f8a34e4a-backup-test-go.md) - `internal/engine/backup_test.go`
