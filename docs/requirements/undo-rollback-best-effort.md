---
id: "REQ-9fca28c9-d3b1-4ad7-8f24-13b2ec7aeab2"
type: requirement
name: "巻き戻し自体の失敗は best-effort で続行し、全件を stderr へ報告して停止する"
section: placement-rollback
status: stable
derives_from:
  - "UC-0b6f60cb-3e98-4ee7-8929-4d94a29f0af6"
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  A failure of an individual inverse operation SHALL NOT abort the rollback: the remaining
  entries of the undo journal SHALL still be rolled back. Once every rollback attempt has
  been made, the original apply error together with a list of the items that could not be
  rolled back SHALL be reported in full on stderr, and the command SHALL then stop, taking
  the same shape as conflict reporting — enumerate everything, then raise one aggregated
  error. This report SHALL always go to stderr and SHALL NOT be subject to the
  silent-on-success discipline, being a failure path.
specification_ja: |
  個々の逆操作の失敗によって巻き戻しを中断してはならず、undo ジャーナルの残りの巻き戻しは
  続行しなければならない。全ての巻き戻し試行が終わった時点で、元の apply エラーと巻き戻せ
  なかった項目の一覧を stderr へ全報告して停止しなければならない（conflict 報告と同じ
  「全件列挙してから 1 本の集約エラー」の形）。この報告は失敗時の処理のため常時 stderr とし、
  成功時沈黙の対象にしてはならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:31:01+09:00"
---
# 巻き戻し自体の失敗は best-effort で続行し、全件を stderr へ報告して停止する

## 仕様

- **巻き戻し自体の失敗は best-effort 続行**: 個々の逆操作が失敗してもジャーナルの残りの
  巻き戻しは続行する。全ての巻き戻し試行が終わった時点で、元の apply エラーと巻き戻せ
  なかった項目の一覧を stderr へ全報告して停止する（`reportConflicts` と同じ「全件列挙して
  から 1 本の集約エラー」の形）
- **報告は常時 stderr**（失敗経路のため沈黙対象外・既定 silent の対象にしない）

> **上は原文の写しで、規範は frontmatter が正**。undo ジャーナルそのものの規範は
> [途中失敗した apply / rollback はインメモリ undo ジャーナルで全 FS 変更を巻き戻す](undo-journal-rollback.md)、成功時沈黙の出力規律そのものは [成功時はデフォルト沈黙とし warning と error は常時 stderr へ出す](silent-on-success.md) の担当。conflict の全件報告
> （`reportConflicts`）は [conflict で停止するときは全件を対処ガイダンス付きで列挙してから 1 本の集約エラーを返す](conflict-full-report.md) の担当で、本 item はその形を巻き戻し失敗の報告にも
> 用いることだけを規定する。

## 出典

`docs/spec.md`「配置動作仕様」→「途中失敗時の巻き戻し」節の箇条書き第 1・2 項。

決定の実体は [巻き戻しの失敗は best-effort で続行する](../adr/undo-failure-best-effort.md)で、
失敗時の処理を沈黙の対象にしない出力規律は [指定が無いときは成功時に何も出力しない](../adr/silent-on-success.md) が定めている。
