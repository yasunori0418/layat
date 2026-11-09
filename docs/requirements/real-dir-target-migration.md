---
id: "REQ-7cee95dd-bc5a-4e86-bebc-6080ef78fe26"
type: requirement
name: "実 dir の target は全 leaf が除去可能なときだけ全体を配置前除去して symlink 化する"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  When the target itself is a real directory, the engine SHALL classify every leaf below
  it at any depth. Only when every leaf is either a recorded and stale symlink — recorded
  by the entry's own previous generation, matching that record on disk, and absent from
  the next generation — or an empty sub-directory of any origin, SHALL the whole target be
  removed before placement (PreRemove: leaves by unlink, directories by rmdir from child
  to parent) and the symlink newly placed; this removal SHALL NOT be reported as a warning,
  being an intended migration, and SHALL instead fall under the ordinary output discipline
  of the placement report. If even one other leaf exists — a real file or directory with
  content, a foreign symlink, or a self-contradictory entry that also remains in the next
  generation — the command SHALL stop with an error for the target as a whole and SHALL
  NOT remove it partially.
specification_ja: |
  target 自身が実 dir のとき、engine は配下（任意深さ）の全 leaf を判定しなければならない。
  全 leaf が「recorded かつ古くなった symlink（当該 entry 自身の前世代が記録・on-disk 一致・
  次世代に無い）」または「空の sub dir（由来を問わない）」であるときに限り、target 全体を
  配置前に除去（PreRemove: leaf は unlink・dir は子から親へ rmdir）してから symlink を新規
  配置しなければならない。この除去は意図された移行であり warning にしてはならず、配置レポート
  の通常の出力規律に従わせなければならない。それ以外の leaf が 1 つでもある（中身のある実
  file / dir・foreign symlink・
  次世代にも残る自己矛盾）ときは target 全体をエラーで停止しなければならず、部分除去を
  してはならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# REQ-7cee95dd-bc5a-4e86-bebc-6080ef78fe26: 実 dir の target は全 leaf が除去可能なときだけ全体を配置前除去して symlink 化する

## 仕様

```
target 自身が実 dir のとき、配下（任意深さ）の全 leaf を判定する:
  - 全 leaf が「recorded ∧ stale な symlink（自身の前世代が記録・on-disk 一致・次世代に無い）」
    または「空の sub dir（由来を問わない）」なら → target 全体を配置前に除去（PreRemove: leaf は Unlink・
    dir は子から親へ Rmdir・silent・`-v` で可視）してから symlink を新規配置
  - 上記以外の leaf が 1 つでもある（中身のある実 file/dir・foreign symlink・次世代にも残る自己矛盾）
    → target 全体をエラーで停止（部分除去はしない）
```

> **上は原文の写しで、規範は frontmatter が正**。この migration が失敗して「エラーで停止」
> と判定された実 dir も `--backup` の退避対象に入る。退避対象を判定各段の結論
> （エラー停止 / copy foreign スキップ）で括る規範と、退避が配置手順のどの段に入るかは
> [backup 退避は配置前除去の後・配置の前に置き、drift 修復の処理でも同じく実施する](backup-stage-position.md) の担当（退避そのものの契約は [apply --backup は配置を塞ぐ記録外実体を rename 退避してから配置する](apply-backup-rename-aside.md)）。

祖先 symlink の migration は [祖先 symlink は自己記録の古くなったものだけを配置前除去し、それ以外はエラーで停止する](ancestor-symlink-nest-migration.md)、method 変更 symlink→copy の migration は
[method 変更は symlink→copy のみ配置前除去で移行し、copy→symlink は移行しない](method-change-symlink-to-copy-migration.md) の担当。除去を既定 silent とし `-v` で可視化するという出力規律そのもの
（→ [指定が無いときは成功時に何も出力しない](../adr/silent-on-success.md)・[-v で配置レポートを opt-in で出す](../adr/verbose-opts-into-placement-report.md)）は [成功時はデフォルト沈黙とし warning と error は常時 stderr へ出す](silent-on-success.md) / [冗長度は -v、デバッグは --debug に分離し --json と直交させる](verbose-and-debug-separation.md) の担当で、本 item は「この除去を warning に
せず配置レポート側で扱う」ことだけを規範とする。

## 出典

`docs/spec.md`「配置動作仕様」→「symlink モード」の手順 0.5 と、同節の箇条書き
「target に通常ファイルまたはディレクトリが存在する場合はエラーで停止（上書きしない）。
ただし実 dir は §0.5 の条件を満たせば例外的に PreRemove で除去して配置する」。

決定の実体は [実ディレクトリの target は全 leaf を除去できるときだけ移行する](../adr/real-dir-target-migration.md)。
