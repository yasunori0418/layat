---
id: "REQ-2b48620a-abaa-43df-a106-954bbba3de56"
type: requirement
name: "method 変更は symlink→copy のみ配置前除去で移行し、copy→symlink は移行しない"
status: stable
derives_from:
  - "UC-403fbe32-b146-401b-8b53-fe67c1e169c5"
specification: |
  When the method for the same target changes from symlink to copy, and the previous
  generation recorded a symlink whose on-disk destination matches that record, the engine
  SHALL remove it before placement (PreRemove) and then newly place the copy; this removal
  SHALL NOT be reported as a warning, being an intended migration, and SHALL instead fall
  under the ordinary output discipline of the placement report.
  When the on-disk link has drifted from the record, the engine SHALL NOT migrate it and
  SHALL fall back to the ordinary foreign classification. A method change from copy to
  symlink SHALL NOT be migrated automatically and SHALL remain a conflict, giving priority
  to protecting copied data the user may have edited.
specification_ja: |
  同一 target で method が symlink→copy に変わり、前世代が記録した symlink の on-disk が
  記録通りであるとき、engine は配置前に除去（PreRemove）してから copy を新規配置しなければ
  ならない。この除去は意図された移行であり warning にしてはならず、配置レポートの通常の出力
  規律に従わせなければならない。on-disk が記録と不一致（readlink drift）のときは移行しては
  ならず、通常の foreign 判定へフォールバックしなければならない。method 変更 copy→symlink は
  自動移行してはならず、conflict のままとしなければならない（ユーザーが編集し得る copy
  データの保護を優先するため）。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# method 変更は symlink→copy のみ配置前除去で移行し、copy→symlink は移行しない

## 仕様

```
target と同一 target で method が symlink→copy に変わるとき、前世代が記録した symlink で on-disk が
記録通りなら → 配置前に除去（PreRemove・silent）してから copy を新規配置
readlink drift（on-disk が記録と不一致）は移行せず通常の foreign 判定へフォールバックする
```

method 変更 copy→symlink は**自動移行しない**（ユーザー編集済み copy データの保護を優先し、
従来通り conflict のまま）。

> **上は原文の写しで、規範は frontmatter が正**。copy→symlink の conflict は `--backup`
> を付ければ退避 + 配置される（[apply --backup は配置を塞ぐ記録外実体を rename 退避してから配置する](apply-backup-rename-aside.md)）。退避が配置手順のどの段に入るかは
> [backup 退避は配置前除去の後・配置の前に置き、drift 修復の処理でも同じく実施する](backup-stage-position.md) の担当。

祖先 symlink の migration は [祖先 symlink は自己記録の古くなったものだけを配置前除去し、それ以外はエラーで停止する](ancestor-symlink-nest-migration.md)、実 dir target の migration は [実 dir の target は全 leaf が除去可能なときだけ全体を配置前除去して symlink 化する](real-dir-target-migration.md) の
担当。除去を既定 silent とし `-v` で可視化するという出力規律そのもの（→ [指定が無いときは成功時に何も出力しない](../adr/silent-on-success.md)・[-v で配置レポートを opt-in で出す](../adr/verbose-opts-into-placement-report.md)）は
[成功時はデフォルト沈黙とし warning と error は常時 stderr へ出す](silent-on-success.md) / [冗長度は -v、デバッグは --debug に分離し --json と直交させる](verbose-and-debug-separation.md) の担当で、本 item は「この除去を warning にせず配置レポート側で
扱う」ことだけを規範とする。

## 出典

`docs/spec.md`「配置動作仕様」→「symlink モード」の手順 0.6 と、同節の箇条書き
「method 変更 copy→symlink は自動移行しない」。

決定の実体は [method の symlink から copy への変更を自動で移行する](../adr/symlink-to-copy-migration.md)・[method の copy から symlink への変更は移行しない](../adr/copy-to-symlink-stays-conflict.md)。
