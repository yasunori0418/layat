---
id: "CASE-ed4b32fd-4f2e-497c-baa4-cc91f8a34e4a"
type: test_case
name: "backup_test.go — --backup の退避ライフサイクルと退避先の保護"
status: stable
resource: "internal/engine/backup_test.go"
covers:
  - "TC-ed4992c0-8513-4383-be0a-e45acbbc229f"
  - "TC-cd03ff56-6ed0-4100-9ec4-eb681ab66407"
follows:
  - "TP-e7c25263-6d2d-4a37-8275-26906889d912"
  - "TP-deb05610-44bc-4962-8939-952392e5fbd0"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:31+09:00"
---
# backup_test.go — --backup の退避ライフサイクルと退避先の保護

## 対象

`internal/engine/backup_test.go`

`apply --backup` の rename 退避（→ [--backup は配置を塞ぐ記録外の実体だけを退避する](../../adr/backup-scope-unrecorded-entities.md)、issue #169）を tmpdir 上の実 FS で検証する
統合テスト。大半は `Apply` を通した end-to-end だが、plan / execute 間の窓を作る 1 本だけ
`applier.backup` を直接駆動する。

## 検証内容

**発動・命名・非発動**（[--backup の退避が opt-in で発動し、commit 後は残り、途中失敗では戻る](backup-aside-lifecycle.md)）

- 配置を塞ぐ通常ファイルが `<target>.layat-backup`（既定 suffix）へ退避され、entry が
  新規配置され、target が `Result.BackedUp` に載る
- `--backup=<suffix>` が既定に代えて指定 suffix を使う
- `--backup` 無し（既定）では同じ配置が conflict で停止する

**run 終了後の扱い**（[--backup の退避が opt-in で発動し、commit 後は残り、途中失敗では戻る](backup-aside-lifecycle.md)）

- commit 成功後も退避物が残る（`--recopy` の退避と異なり、ユーザーのバックアップとして
  無期限に保持する → [退避物を成功後も残して利用者の所有物にする](../../adr/backup-aside-left-to-user.md)）
- 同一バッチの後続配置が失敗すると退避が rename で戻り、target が pre-apply の内容へ
  復元され、退避先パスは消える

**段の位置と巻き添え警告**（[--backup の退避が opt-in で発動し、commit 後は残り、途中失敗では戻る](backup-aside-lifecycle.md)）

- 世代スキップ（drift 修復）の処理でも Backup 段が発火し、shell 再入間に現れた foreign な
  通常ファイルを退避して配置し直す（PreRemove と違い「derivation 不変なら発火しない」
  不変条件を持たない → [backup 退避は配置前除去の後・配置の前に置き、drift 修復の処理でも同じく実施する](../../requirements/backup-stage-position.md)）
- `--dryrun --backup` が塞いでいる target を conflict ではなく退避予定として報告し、
  FS には一切触れない
- 実 dir target を丸ごと退避したとき、前世代が自分の古くなった symlink として記録している
  兄弟 leaf が removeStale から drift 警告されない（親ごと 1 回の rename で去ったのは
  想定内 → planner.markDirEntriesPreRemoved の回帰テスト）

**退避先の保護**（[退避先が既存なら停止し、rename 直前に退避先を再検証する](backup-conflict-and-toctou.md)）

- plan 時点で `<target>.<suffix>` が既存なら conflict で停止し、利用者に知らせずに上書きしない
- plan 時に不在だった退避先が execute 時点で存在する（TOCTOU 窓）ケースを
  `applier.backup` の直接駆動で作り、rename 直前の再検証が並行生成物を検出して
  loud に中断する（→ [退避先が既にあれば conflict で止める](../../adr/backup-existing-destination-conflict.md)）
