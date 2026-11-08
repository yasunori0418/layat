---
id: "TC-cd03ff56-6ed0-4100-9ec4-eb681ab66407"
type: test_condition
name: "退避先が既存なら停止し、rename 直前に退避先を再検証する"
status: stable
mitigates:
  - "RISK-3719d298-8175-4501-95d5-75f3a29568fe"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:31+09:00"
---
# TC-cd03ff56-6ed0-4100-9ec4-eb681ab66407: 退避先が既存なら停止し、rename 直前に退避先を再検証する

## テスト条件

`--backup` が退避先 `<target>.<suffix>` を上書きしないことを、plan 時点と execute 時点の
両方で検証する。

**plan 時点で既存** — 前回の退避の残りなど、退避先が既に存在する状態で `--backup` を
付けた apply は conflict で停止する。利用者に知らせずに上書きしない（→ [apply --backup は配置を塞ぐ記録外実体を rename 退避してから配置する](../../requirements/apply-backup-rename-aside.md)）。

**plan と execute の間に出現（TOCTOU）** — plan 計算時には退避先が存在せず、`backup()`
が実行される時点までに他プロセスがそこを作った場合、rename 直前の再検証がその並行生成物を
検出して停止する。plan 時の判定を信じて rename すれば、並行して作られた実データを利用者に知らせずに
上書きする（→ [退避先が既にあれば conflict で止める](../../adr/backup-existing-destination-conflict.md)）。

この条件は full `Apply` では窓を作れないため、`BackupAction` を組み立てて `applier.backup`
を直接駆動し、plan 済みの状態から execute までの間に退避先を作る、という形で検証する。

上位の規範は [--backup の退避が opt-in で発動し、commit 後は残り、途中失敗では戻る](backup-aside-lifecycle.md) と同じく [エンジンとコマンド層の Go テストは nix を介さない実 FS 統合テストを主戦力とする](../../test-plan/go-test-layering.md)。

## 対応する CASE

[backup_test.go — --backup の退避ライフサイクルと退避先の保護](backup-test-go.md)（`internal/engine/backup_test.go`）。
