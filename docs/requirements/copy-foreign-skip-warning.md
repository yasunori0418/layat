---
id: "REQ-07c3b735-3744-4778-a640-8c6fb66f4aa7"
type: requirement
name: "copy が foreign 実ファイルを skip したときは warning で可視化する"
section: placement-copy
status: stable
derives_from:
  - "UC-403fbe32-b146-401b-8b53-fe67c1e169c5"
specification: |
  When the target of a copy entry exists and the previous-generation manifest does not
  hold that copy entry — that is, the file was not placed by layat and is foreign — the
  engine SHALL skip the copy under place-once but SHALL emit a warning stating that an
  existing file was present at the target and the copy was skipped, symmetrically with the
  foreign warning for symlinks, so as to prevent the user from mistaking the content for
  something layat put there. The engine SHALL NOT overwrite it and SHALL NOT stop the apply
  as a whole. Whether layat placed it SHALL be decided by the presence of the entry in the
  previous-generation manifest and SHALL NOT be decided from the content.
specification_ja: |
  copy entry の target が存在し、かつ前世代 manifest にこの copy entry が無いとき（= layat が
  置いていない foreign なファイル）、engine は place-once により copy を skip しつつ、
  「target に既存ファイルがあり copy をスキップした」という warning を出さなければならない
  （symlink の foreign 警告と対称化し、「layat が中身を置いた」という誤認を防ぐため）。上書きは
  してはならず、apply 全体を止めてもならない。「自分が置いたか」は前世代 manifest の entry
  有無で判別しなければならず、内容で判別してはならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# copy が foreign 実ファイルを skip したときは warning で可視化する

## 仕様

**foreign 実ファイルの skip は warning で可視化する**: target が存在し**かつ前世代 manifest に
この copy entry が無い**（= layat が置いていない foreign ファイル）のとき、place-once により
skip するが **warning を出す**（「target に既存ファイルがあり copy をスキップした」）。
symlink の foreign 警告と対称化し、「layat が中身を置いた」と誤認する masking を防ぐ。
上書きはせず apply 全体も止めない。「自分が置いたか」は前世代 manifest の entry 有無で判別し、
内容は判別しない。

> **上は原文の写しで、規範は frontmatter が正**。place-once そのものは [copy は target 不在のときだけマテリアライズする place-once で世代管理の対象外とする](copy-place-once.md)、
> symlink 側の foreign 警告は [symlink 配置は親 dir を作り配置元/subpath を指すリンクを張り、foreign symlink は警告して後勝ちする](symlink-placement-procedure.md) の担当。`apply --backup` 有効時にこの skip +
> warning が退避 + copy 新規配置へ変わることは [apply --backup は配置を塞ぐ記録外実体を rename 退避してから配置する](apply-backup-rename-aside.md) / [backup 退避は配置前除去の後・配置の前に置き、drift 修復の処理でも同じく実施する](backup-stage-position.md) の担当。

## 出典

`docs/spec.md`「配置動作仕様」→「copy モード（place-once・ユーザー管理）」節の箇条書き第 1 項。

決定の実体は [copy は foreign の実ファイルを上書きせず warning を出して見送る](../adr/copy-skips-foreign-file-with-warning.md)で、対称化の元に
なる symlink の foreign 警告は [cross-config の同一 target は後勝ちにし、foreign symlink の上書きで警告する](../adr/cross-config-last-wins-with-foreign-warning.md) が定めている。
