---
id: "REQ-61856da1-8883-401e-ad57-9f326b96d400"
type: requirement
name: "既存 symlink の張替えは unlink + symlink の 2 操作で行い冪等な再実行で収束させる"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  Replacing an existing symlink SHALL be carried out as two operations, an unlink followed
  by a symlink, and an atomic swap based on rename SHALL NOT be adopted. A crash between
  the two operations may make the target vanish for a while, and this SHALL be accepted on
  the grounds that an idempotent re-run converges, consistently with the guarantee that
  the generations that get stacked are always fully applied. When the run itself detects
  an error at a later stage — that is, in any case other than a crash such as SIGKILL or
  power loss — the destination captured before the unlink SHALL be recorded in the undo
  journal and restored by re-creating the symlink.
specification_ja: |
  既存 symlink の張替えは unlink + symlink の 2 操作で行わなければならず、rename ベースの
  atomic swap を採ってはならない。2 操作の間でクラッシュすると target が一時消失しうるが、
  冪等な再実行で収束することを根拠にこれを受け入れなければならない（「積まれる世代は常に
  完全適用済み」と整合）。
  クラッシュ（SIGKILL・電源断）以外で、この run 自身が後続の段でエラーを検知した場合は、
  unlink 前に捕捉した張替え先を undo ジャーナルへ記録し、symlink の再作成で復元しなければ
  ならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# 既存 symlink の張替えは unlink + symlink の 2 操作で行い冪等な再実行で収束させる

## 仕様

既存 symlink の張替えは **unlink + symlink の 2 操作**で行う（rename ベースの atomic swap は
採らない）。間でクラッシュすると target が一時消失しうるが、**冪等な再実行で収束**する
（「積まれる世代は常に完全適用済み」と整合）。クラッシュ（SIGKILL・電源断）以外でこの run
自身が後続の段でエラーを検知した場合は、unlink 前の張替え先を undo ジャーナルが記録しており
re-symlink で復元する。

> **上は原文の写しで、規範は frontmatter が正**。undo ジャーナルそのものの規範は
> [途中失敗した apply / rollback はインメモリ undo ジャーナルで全 FS 変更を巻き戻す](undo-journal-rollback.md) の担当で、本 item は張替えを 2 操作で行うことと、その復元素材（unlink 前の
> 張替え先）を捕捉することだけを規定する。「積まれる世代は常に完全適用済み」（コミット
> 最後）は [世代は link farm derivation を layat 自前 profile へコミットして積み、前世代 manifest から古くなった配置を除去する](generation-mechanism-link-farm-manifest.md) の担当。

## 出典

`docs/spec.md`「配置動作仕様」→「symlink モード」の箇条書き第 1 項。

決定の実体は [symlink を unlink と symlink の 2 操作で張り替える](../adr/relink-unlink-then-symlink.md)で、冪等再実行による収束の前提は [新規配置と張替えを先に、世代のコミットを最後に行う](../adr/placement-order-commit-last.md) が、undo ジャーナルによる
復元は [途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](../adr/in-memory-undo-journal.md) が定めている。
