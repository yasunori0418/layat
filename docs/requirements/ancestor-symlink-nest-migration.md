---
id: "REQ-c9ab91c1-f778-4f87-a2ea-c66d6b3c2575"
type: requirement
name: "祖先 symlink は自己記録の古くなったものだけを配置前除去し、それ以外はエラーで停止する"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  Before placing a symlink, the engine SHALL walk every ancestor component of the target
  with `lstat` and SHALL treat an ancestor symlink asymmetrically. An ancestor that is
  self-recorded stale — recorded by the entry's own previous-generation manifest, whose
  on-disk destination matches that record, and which is absent from the next generation —
  SHALL be removed before placement (PreRemove) so that the entries below it are newly
  placed, thereby migrating them into the nest; this removal SHALL NOT be reported as a
  warning, being an intended migration, and SHALL instead fall under the ordinary output
  discipline of the placement report. An ancestor that is foreign — not recorded, not
  matching the recorded destination, or having no previous generation — and a
  self-contradictory ancestor that also remains in the next generation SHALL cause the
  command to stop with an error, because nesting entries below such an ancestor would
  pollute the store or create dangling links.
specification_ja: |
  symlink の配置に先立ち、engine は target の各祖先 component を `lstat` で walk し、
  祖先の symlink を非対称に扱わなければならない。自己記録の古くなった祖先（当該 entry 自身の
  前世代 manifest が記録し、on-disk が記録 dest と一致し、次世代には無い）は配置前に除去
  （PreRemove）して配下の子を新規配置し、ネストへ移行させなければならない。この除去は
  意図された移行であり warning にしてはならず、配置レポートの通常の出力規律に従わせなければ
  ならない。foreign な祖先（記録なし / 記録 dest と不一致 / 前世代なし）と、次世代にも祖先
  が残る自己矛盾はエラーで停止しなければならない（配下にネストすると store 汚染 / dangling
  を招くため）。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T13:39:07+09:00"
---
# 祖先 symlink は自己記録の古くなったものだけを配置前除去し、それ以外はエラーで停止する

## 仕様

```
target の各祖先 component を lstat で walk。symlink の祖先があれば非対称に扱う:
  - 自己記録 stale（自身の前世代 manifest が記録・on-disk が記録 dest と一致・次世代に無い）
    → その祖先を配置前に除去（PreRemove）し、配下子を新規配置してネスト移行する（silent・`-v` で可視）
  - foreign（記録なし / 記録 dest と不一致 / 前世代なし）または次世代にも祖先が残る自己矛盾
    → エラーで停止（配下にネストできない。store 汚染 / dangling を防ぐ）
```

> **上は原文の写しで、規範は frontmatter が正**。原文が参照する次の規範は本 item の
> 担当ではない。
>
> - 既定 silent と `-v` による可視化の出力規律そのもの（→ [指定が無いときは成功時に何も出力しない](../adr/silent-on-success.md)・[-v で配置レポートを opt-in で出す](../adr/verbose-opts-into-placement-report.md)）→ [成功時はデフォルト沈黙とし warning と error は常時 stderr へ出す](silent-on-success.md) /
>   [冗長度は -v、デバッグは --debug に分離し --json と直交させる](verbose-and-debug-separation.md)。本 item は「この除去を warning にせず配置レポート側で扱う」ことだけを
>   規範とし、そのレポートを指定が無いときに出すか `-v` のときに出すかは委譲する
> - `--backup` 有効時にこのエラー停止がどう変わるか → 祖先 symlink conflict は
>   `--backup` の対象外（[apply --backup は配置を塞ぐ記録外実体を rename 退避してから配置する](apply-backup-rename-aside.md)・構造問題であり退避では解消しないため）

実 dir target の migration は [実 dir の target は全 leaf が除去可能なときだけ全体を配置前除去して symlink 化する](real-dir-target-migration.md)、method 変更 symlink→copy の migration は
[method 変更は symlink→copy のみ配置前除去で移行し、copy→symlink は移行しない](method-change-symlink-to-copy-migration.md) の担当。

## 出典

`docs/spec.md`「配置動作仕様」→「symlink モード」の手順 0。

決定の実体は [祖先 symlink は自己記録の古くなったものだけを移行する](../adr/ancestor-symlink-recorded-stale-only.md)・[移行する祖先の配下の entry は lstat を経ず不在として配置する](../adr/nested-children-placed-as-absent.md)・[配置前除去による移行を warning にせず -v のときだけ報告する](../adr/migration-reported-silently.md) で、foreign 祖先のエラー停止は [祖先 symlink は自己記録の古くなったものだけを移行する](../adr/ancestor-symlink-recorded-stale-only.md) が
定めている。
