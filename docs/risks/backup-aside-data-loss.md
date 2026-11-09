---
id: "RISK-3719d298-8175-4501-95d5-75f3a29568fe"
type: risk
name: "--backup の rename 退避がユーザーの実データを上書き・喪失させる"
status: stable
threatens:
  - "REQ-5dd5a4e9-6162-4fa5-b295-66844f5a4f3b"
  - "REQ-9b0046e0-8ddc-4c0b-940e-3fe6f36d0e98"
likelihood: medium
impact: high
level: high
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:36:11+09:00"
---
# --backup の rename 退避がユーザーの実データを上書き・喪失させる

## リスク

`apply --backup` は conflict の脱出ハッチとして、配置を塞ぐ記録外実体を
`<target>.<suffix>` へ rename 退避する。退避対象は定義上「layat が記録していない実データ」
であり、扱いを誤れば直接ユーザーのファイルを失わせる。

失われ方は 3 つある。

1. **退避先を利用者に知らせずに上書きする** — `<target>.<suffix>` が既存（前回の退避の残り）のとき、
   conflict で停止せず rename してしまうと、前回退避した実データが消える
2. **plan / execute 間の TOCTOU** — plan 時点で退避先が不在でも、execute 時点で他プロセスが
   そこにファイルを作っていることがある。rename 直前に再検証しなければ、その並行生成物を
   利用者に知らせずに上書きする（→ [退避先が既にあれば conflict で止める](../adr/backup-existing-destination-conflict.md)）
3. **退避が commit 後に消える** — `--recopy` の rename 退避は成功時に discardJournal が
   掃除するが、`--backup` の退避はユーザーのバックアップであり永久に残す契約である
   （→ [退避物を成功後も残して利用者の所有物にする](../adr/backup-aside-left-to-user.md)・`reset` も復元しない）。掃除の対象を取り違えると、ユーザーが唯一の
   コピーだと思っていたものが消える

逆方向の欠陥も本リスクに含む。`--backup` の指定が無いときは conflict のまま停止するはずの
ケースが退避へ流れると、ユーザーが opt-in していない破壊的操作が、指定が無いまま実行されることになる。
世代スキップ（drift 修復）の処理で Backup 段だけは到達し続けるという段の位置付け
（→ [backup 退避は配置前除去の後・配置の前に置き、drift 修復の処理でも同じく実施する](../requirements/backup-stage-position.md)）が崩れる場合も、退避されるべき実体が conflict のまま放置されるか、
逆に修復の処理で予期しない退避が起きるかのどちらかになる。

## 実現性

**likelihood: medium** — 退避は opt-in であり、指定が無いときは実行されない。ただし発火条件
（記録外実体が target を塞ぐ）は foreign tool による書き換えで日常的に成立する。

**impact: high** — 対象が定義上「layat の管理外にあるユーザーの実データ」なので、
喪失は復旧不能。

## 緩和

[--backup の退避が opt-in で発動し、commit 後は残り、途中失敗では戻る](../test/atomicity/backup-aside-lifecycle.md)（退避のライフサイクル: 発動・suffix・commit 後の存置・巻き戻し）と
[退避先が既存なら停止し、rename 直前に退避先を再検証する](../test/atomicity/backup-conflict-and-toctou.md)（退避先既存の conflict と rename 直前の再検証）が緩和する。
