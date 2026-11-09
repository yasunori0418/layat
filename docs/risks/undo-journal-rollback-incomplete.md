---
id: "RISK-68e810c5-4e68-4b25-9bc0-6b2613022b49"
type: risk
name: "途中失敗した run の巻き戻しが不完全で、FS に部分適用の痕跡が残る"
status: stable
threatens:
  - "REQ-5e75aabc-0e8f-4a6c-92bd-a712dc68a940"
  - "REQ-61856da1-8883-401e-ad57-9f326b96d400"
likelihood: high
impact: high
level: high
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-05T00:25:33+09:00"
---
# 途中失敗した run の巻き戻しが不完全で、FS に部分適用の痕跡が残る

## リスク

apply / rollback が FS 変更バッチの途中で失敗したとき、その run が既に加えた変更が
巻き戻されずに残る。逆操作そのものが誤っている（新規 symlink を消し忘れる・張替えの
prevDest を復元しない・PreRemove が消した実ディレクトリを作り直さない・`--recopy` /
`--backup` の rename 退避を戻さない）場合と、逆操作は正しいが適用順序が誤っている
（LIFO でないため親ディレクトリを作り直す前に子 symlink を復元しようとする）場合の
どちらでも起こる。

被害は「失敗した apply が痕跡を残さない」という保証の喪失に留まらない。失敗 run が
新規に置いた symlink は前世代 manifest に記録が無いため、次の apply の古くなった配置の除去でも
対象にならず、FS 上に恒久的なゴミとして残る（→ [途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](../adr/in-memory-undo-journal.md) の背景）。PreRemove の migration が
半分だけ進んだ状態は更に悪く、実ディレクトリが消えたまま新しい dir symlink も無い、
という配置前より劣化した状態を作る。

逆に、巻き戻すべきでない段で巻き戻すのも同じ深刻度の欠陥である。commit（`nix-env
--set`）の失敗時点では FS の全書き込みが成功しており整合しているため、ここで unwind
すると整合状態を壊す（→ [undo の発火を FS 反映の段に限り、世代コミットの後は対象にしない](../adr/undo-scope-ends-at-commit.md)）。この非対称性が崩れる方向の欠陥も本リスクに含む。

`design` ではなく `requirement` へ張るのは、undo ジャーナルという機構そのものが
[途中失敗した apply / rollback はインメモリ undo ジャーナルで全 FS 変更を巻き戻す](../requirements/undo-journal-rollback.md) の `specification` に「インメモリ undo ジャーナルへ逆操作を 1 件記録する」と
書き下ろされて要求へ昇格しており、対応する design item が `docs/design/` に存在しない
ため（[途中失敗した apply / rollback はインメモリ undo ジャーナルで全 FS 変更を巻き戻す](../requirements/undo-journal-rollback.md) を satisfies するのは [engine は nix を使わない tmpdir 統合テストで検証し、古くなった配置の除去の安全不変条件を table-driven に重点配分する](../design/engine-test-go-unit-tmpdir.md) というテスト戦略の design だけ）。

## 実現性

**likelihood: high** — 逆操作は 6 種（unlinkNew / relinkOld / removeCopy / restoreRename /
mkdir / backup 戻し）あり、それぞれの前方操作が置かれる段（PreRemove / Backup / place /
materializeCopies / removeStale）と組み合わせで機能する。段の追加・順序変更のたびに
journal への記録し忘れが入りうる構造で、実際に PreRemove の追加時に journal 側の対応が必要になっている
（現在の決定は → [配置前除去を planner が計算するデータとして表す](../adr/preremove-as-planner-data.md) / [祖先 symlink は自己記録の古くなったものだけを移行する](../adr/ancestor-symlink-recorded-stale-only.md) / [実ディレクトリの target は全 leaf を除去できるときだけ移行する](../adr/real-dir-target-migration.md)）。

**impact: high** — 復旧はユーザーの手作業になり、消えた実ディレクトリは復旧手段が無い。

## 緩和

[undo の逆操作が、対応する前方操作の直前状態を単体で復元する](../test/atomicity/undo-inverse-op-correctness.md)（逆操作単体の正しさ）・[unwind が LIFO で巻き戻し、1 件の失敗でも継続して全件報告し、成功時は journal を破棄する](../test/atomicity/unwind-lifo-and-journal-lifecycle.md)（LIFO 順序と journal のライフサイクル）・
[実 FS の故障注入で FS 変更段の途中を失敗させ、apply 前の状態が完全に復元される](../test/atomicity/mid-batch-rollback-restores-preapply.md)（実 FS 故障注入による end-to-end 巻き戻し）・[commit 段の失敗では配置を巻き戻さず、退避ファイルも掃除しない](../test/atomicity/commit-failure-not-unwound.md)（commit 失敗は
unwind しない非対称性）が緩和する。検証方針そのものは [原子性は実 FS の条件で故障を誘発して不変条件ごとに検証する](../test-plan/fault-injection-atomicity.md) が定める。
