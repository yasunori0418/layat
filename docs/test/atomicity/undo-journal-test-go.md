---
id: "CASE-154af597-df3e-49fb-a96b-b4f371dfcc63"
type: test_case
name: "undo_journal_test.go — 実 FS 故障注入による end-to-end 巻き戻しと commit 失敗の非対称性"
status: stable
resource: "internal/engine/undo_journal_test.go"
covers:
  - "TC-3b02ab58-5bfb-4ae7-9ac1-c69e2ece2722"
  - "TC-e048211f-0844-4446-9957-fb444c1da4e5"
  - "TC-83fe0d4a-9730-431a-b7f9-b731060b9484"
follows:
  - "TP-e7c25263-6d2d-4a37-8275-26906889d912"
  - "TP-deb05610-44bc-4962-8939-952392e5fbd0"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:31+09:00"
---
# undo_journal_test.go — 実 FS 故障注入による end-to-end 巻き戻しと commit 失敗の非対称性

## 対象

`internal/engine/undo_journal_test.go`

`Apply` を end-to-end で駆動する tmpdir 統合テスト。故障注入の共通ヘルパー `blockWrite`
が対象ディレクトリを mode 0o555 で作り、plan 時の `Lstat` / `ReadDir` は通したまま
execute 時の書き込みだけを EACCES で落とす（通常ファイルで置き換える手法だと plan 時に
ENOTDIR で落ち、「バッチ途中の失敗」にならない）。root では権限チェックが迂回されるため
`os.Geteuid() == 0` を検出して skip する（→ [途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](../../adr/in-memory-undo-journal.md)、[原子性は実 FS の条件で故障を誘発して不変条件ごとに検証する](../../test-plan/fault-injection-atomicity.md)）。

## 検証内容

**mid-batch 失敗の巻き戻し**（[実 FS の故障注入で FS 変更段の途中を失敗させ、apply 前の状態が完全に復元される](mid-batch-rollback-restores-preapply.md)）

- 3 本の新規 symlink 配置の 2 本目が失敗したとき、1 本目は巻き戻され、3 本目は配置されて
  いない。blocker ディレクトリ自体は無傷で、世代も commit されない
- 同一バッチで先に行われた張替え（PlaceReplace）が pre-apply の dest へ復元される
- 先に materialize 済みの copy が、後続段の removeStale 失敗で除去される
  （materializeCopies は place の後・removeStale の前に実行される）
- PreRemove の実 dir target migration（unlink + rmdir）と、それによって可能になった
  dir symlink の配置が、後続の無関係な配置失敗でまとめて巻き戻る。実ディレクトリと
  配下の leaf symlink が元どおりになり、half-migrated で残らない
- `--recopy` の rename 退避が rename で戻り、新しい copy 内容ではなく pre-apply の
  （ローカル編集済み）内容が残る。stray な `.layat-recopy-aside` を残さない

**commit 失敗の非対称性**（[commit 段の失敗では配置を巻き戻さず、退避ファイルも掃除しない](commit-failure-not-unwound.md)）

- commit 失敗では配置が巻き戻らず、そのまま残る（冪等な再実行が収束する → [新規配置と張替えを先に、世代のコミットを最後に行う](../../adr/placement-order-commit-last.md)、
  [symlink を unlink と symlink の 2 操作で張り替える](../../adr/relink-unlink-then-symlink.md)）
- commit 失敗後は `--recopy` の退避ファイルも残る（discardJournal は commit 成功後に
  しか実行されず、unwind も起きない）

**journal の記録順序**（[unwind が LIFO で巻き戻し、1 件の失敗でも継続して全件報告し、成功時は journal を破棄する](unwind-lifo-and-journal-lifecycle.md)）

- `preRemove` を直接駆動し、RemoveUnlink と RemoveRmdir を含むバッチ（children before
  parents）の journal が、LIFO 巻き戻しで親を先に復元する順で記録される
  （→ [配置前除去の除去も undo の対象にする](../../adr/undo-covers-preremove.md)、[配置前除去を planner が計算するデータとして表す](../../adr/preremove-as-planner-data.md)）
