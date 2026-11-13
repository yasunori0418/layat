---
id: "REQ-053cfed2-265a-4997-a27b-97b0bed10d8a"
type: requirement
name: "target に通常ファイル・ディレクトリが在れば上書きせずエラーで停止する"
section: placement-symlink
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  When a regular file or a directory exists at the target of a symlink entry, the engine
  SHALL stop with an error and SHALL NOT overwrite it. A real directory SHALL be the sole
  exception, being removed before placement and then placed when it satisfies the
  conditions for real-directory migration.
specification_ja: |
  symlink entry の target に通常ファイルまたはディレクトリが存在する場合、engine は
  エラーで停止しなければならず、上書きしてはならない。実 dir のみが例外でなければならず、
  実 dir migration の条件を満たす場合に限り配置前に除去して配置しなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-05T00:35:05+09:00"
---
# target に通常ファイル・ディレクトリが在れば上書きせずエラーで停止する

## 仕様

target に通常ファイルまたはディレクトリが存在する場合はエラーで停止（**上書きしない**）。
ただし実 dir は条件を満たせば例外的に PreRemove で除去して配置する。

> **上は原文の写しで、規範は frontmatter が正**。実 dir が例外になる条件そのもの
> （全 leaf 判定と部分除去の禁止）は [実 dir の target は全 leaf が除去可能なときだけ全体を配置前除去して symlink 化する](real-dir-target-migration.md) の担当。`--backup` 有効時にこの
> エラー停止が退避 + 配置へ変わることは [apply --backup は配置を塞ぐ記録外実体を rename 退避してから配置する](apply-backup-rename-aside.md) / [backup 退避は配置前除去の後・配置の前に置き、drift 修復の処理でも同じく実施する](backup-stage-position.md) の担当。

## 出典

`docs/spec.md`「配置動作仕様」→「symlink モード」の箇条書き第 2 項。

決定の実体は [祖先 symlink は自己記録の古くなったものだけを移行する](../adr/ancestor-symlink-recorded-stale-only.md)（祖先 symlink /
実体衝突の扱い）で、実 dir の例外化は [実ディレクトリの target は全 leaf を除去できるときだけ移行する](../adr/real-dir-target-migration.md) が追加した。
