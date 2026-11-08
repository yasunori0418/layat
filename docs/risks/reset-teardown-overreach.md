---
id: "RISK-bb54245e-b284-4b4d-9896-8fec2b4e521c"
type: risk
name: "reset の teardown が管理外の実体まで消し、preview と同意の条件も機能しない"
status: stable
threatens:
  - "REQ-31f2882e-d2e3-4e3b-b783-feb627d73ac6"
  - "REQ-6a950d6d-c191-4235-a1b4-73ffc7c2bb38"
  - "REQ-31dae599-f3a3-4bbe-b367-c955535265da"
likelihood: medium
impact: high
level: high
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:53:31+09:00"
---
# RISK-bb54245e-b284-4b4d-9896-8fec2b4e521c: reset の teardown が管理外の実体まで消し、preview と同意の条件も機能しない

## リスク

`reset` は配置物を無い状態へ戻す FS-only teardown で、copy target を消す唯一の明示手段
でもある（→ [reset は profile を触らない FS-only teardown で配置物を無い状態へ戻す](../requirements/reset-fs-only-teardown.md)）。削除だけを行うコマンドであるぶん、誤りの影響は片方向にしか出ない。
消し過ぎは不可逆で、消し足りないぶんは次の apply が再配置するので実害が無い。

**除去範囲の逸脱** — symlink の除去は古くなった配置の除去と同じ保守的不変条件（layat 管理・前世代の
記録どおりを指すもののみ・foreign は warning で残す）に従う必要がある。ここを緩めると、
ユーザーが別途張った symlink や他ツールの配置物が消える。target フィルタの解釈違い
（存在しない target 名を利用者に知らせずに 0 件処理として成功させる）も、ユーザーが「消えたつもり」の
まま実体を残す方向の誤りになる。

**preview の破れ** — `--dryrun` が副作用を持てば、削除対象を確認するつもりの操作が削除
そのものになる。teardown の preview は、この操作を安全に試す唯一の手段である。

**同意の条件の破れ** — 確認プロンプトの中断（no を選ぶ・非 TTY で `-y` が無い）が機能せず
処理が進めば、ユーザーの同意なくデータが消える。

このリスクは rollback の誤収束（[rollback が FS を前世代へ収束させ損ね、世代表示と実配置が食い違う](rollback-misconverges-fs.md)）とは脅威が別で、あちらは「意図した状態へ
持っていけない」、こちらは「消してはいけないものを消す」。

## 実現性

**likelihood: medium** — 保守的不変条件の実装は古くなった配置の除去と共有されており、そちらの
変更が reset 側へ波及しうる。

**impact: high** — 削除は不可逆で、対象が layat 管理外のユーザーの実データになりうる。

## 緩和

[reset が保守的不変条件の範囲だけを消し、dryrun は副作用ゼロ、非同意では中断する](../test/generations/reset-teardown-safety.md)（reset の teardown 安全性）が緩和する。ただし覆えるのは engine 層まで
（`Confirm` コールバックが false を返したときの中断）で、TTY 判定と `-y` / `--yes` の
解釈は CLI 側の責務のため、同意の条件の残り半分は cli-json 区分の TC が担う。
