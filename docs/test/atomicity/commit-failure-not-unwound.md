---
id: "TC-e048211f-0844-4446-9957-fb444c1da4e5"
type: test_condition
name: "commit 段の失敗では配置を巻き戻さず、退避ファイルも掃除しない"
status: stable
mitigates:
  - "RISK-68e810c5-4e68-4b25-9bc0-6b2613022b49"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:31+09:00"
---
# TC-e048211f-0844-4446-9957-fb444c1da4e5: commit 段の失敗では配置を巻き戻さず、退避ファイルも掃除しない

## テスト条件

[実 FS の故障注入で FS 変更段の途中を失敗させ、apply 前の状態が完全に復元される](mid-batch-rollback-restores-preapply.md) の裏返し。commit（`nix-env --set`）でのみ失敗させ、**巻き戻しが起きない**
ことを検証する。commit へ到達した時点でこの run の FS 書き込みは全て成功しており、FS は
既に整合している。ここで unwind すればその整合状態を壊すため、巻き戻さないのが正しい
（→ [undo の発火を FS 反映の段に限り、世代コミットの後は対象にしない](../../adr/undo-scope-ends-at-commit.md)）。世代が進まないだけで、冪等な再実行が収束する。

条件は次の 3 点。

- 配置済み symlink / copy が commit 失敗後も無傷で残る
- `--recopy` の rename 退避ファイルが残る。discardJournal は commit 成功後にしか実行されず、
  unwind も起きないため、どちらの処理でも触られない。新しい copy が live のまま、
  pre-apply の内容は退避ファイルとして手元に残る
- 世代は進んでいない

commit 段は `Commit` 関数を差し替えて失敗させる。実 FS 条件で誘発すべき対象は FS 操作の
故障であり、commit の故障は nix 側の失敗を表す注入点として double で足りる。

## 対応する CASE

[undo_journal_test.go — 実 FS 故障注入による end-to-end 巻き戻しと commit 失敗の非対称性](undo-journal-test-go.md)（`internal/engine/undo_journal_test.go`）。commit 失敗時の部分 `Result`
の中身（`Unwound` が false であることを含む）は generations 側の
[result_extensions_test.go — 実行結果のインベントリ・世代観測・到達状態・構造化 warning](../generations/result-extensions-test-go.md) も隣接して検証する。
