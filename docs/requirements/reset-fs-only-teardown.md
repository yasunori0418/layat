---
id: "REQ-31f2882e-d2e3-4e3b-b783-feb627d73ac6"
type: requirement
name: "reset は profile を触らない FS-only teardown で配置物を無い状態へ戻す"
status: stable
derives_from:
  - "UC-403fbe32-b146-401b-8b53-fe67c1e169c5"
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  `layat reset <name> [target...]` SHALL be a teardown that returns placed objects to the
  state of not being present; omitting `target` SHALL select every entry, and specifying
  it SHALL select only those entries. Symlinks SHALL be removed under the same
  conservative invariants as stale removal (only those managed by layat and matching the
  record; foreign ones SHALL be left with a warning), the selected entries and the
  destinations recorded for them being taken from the last-applied record when one exists
  and otherwise from the current generation of the profile
  (REQ-9d622539-a1e5-4719-af51-e0cd85b66f35), and copy targets SHALL also be
  deleted, `reset` being the only explicit means of deleting a copy. After removing either
  a symlink or a copy, empty-parent-directory pruning SHALL be applied. Because of the
  risk of data loss, the command SHALL request consent, either through a confirmation
  prompt or via `-y` / `--yes` (the latter being for scripts and CI), and SHALL report the
  deleted targets. `reset` SHALL be an FS-only
  teardown that leaves the profile and generations untouched, while rewriting the
  last-applied record (and, when only a record exists, creating the profileDir that is the
  flock key together with the backref `.root` for the roothash series) as stated by
  REQ-9d622539-a1e5-4719-af51-e0cd85b66f35; as long as the config still holds the entry,
  the next apply SHALL place it again. It SHALL be usable in both home and project mode.
specification_ja: |
  `layat reset <name> [target...]` は配置物を無い状態へ戻す teardown でなければならない。
  target 省略で全 entry、指定でその entry のみを対象としなければならない。symlink は古くなった配置の
  除去と同じ保守的不変条件（layat 管理・記録通りのみ・foreign は warning で残す）で除去し
  （対象の entry と manifest が記録する配置先は、last-applied 記録があれば記録から、無ければ profile の current
  世代から取る。REQ-9d622539-a1e5-4719-af51-e0cd85b66f35）、
  copy target も削除しなければならない（copy を消す唯一の明示手段）。symlink・copy いずれの
  除去後も空親ディレクトリ剪定を適用しなければならない。データ損失リスクのため確認プロンプトを
  出すか `-y` / `--yes`（スクリプト / CI 用）で同意を要求し、削除 target をレポート表示
  しなければならない。`reset` は profile / 世代を触らない FS-only teardown でなければならず
  （REQ-9d622539-a1e5-4719-af51-e0cd85b66f35 が定める last-applied 記録の書き直しと、記録だけが
  あるときの flock の鍵である profileDir（roothash 系列は backref `.root` も）の作成は行う）、
  config が entry を残す限り次の apply がこれを再配置しなければならない。home / project の
  両モードで使用できなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:29:58+09:00"
---
# REQ-31f2882e-d2e3-4e3b-b783-feb627d73ac6: reset は profile を触らない FS-only teardown で配置物を無い状態へ戻す

## 仕様

```bash
layat reset <name> [target...]  # 配置物を無い状態へ戻す。target 省略で全 entry、指定でその entry のみ
```

`reset <name> [target...]` は配置物を**無い状態へ戻す** teardown。symlink は stale 除去と
同じ**保守的不変条件**（layat 管理・記録通りのみ・foreign は warning で残す）で除去し、
**copy target も削除**する（copy を消す唯一の明示手段）。symlink・copy いずれの除去後も
空親ディレクトリ剪定を適用する。データ損失リスクのため**確認プロンプト**を出すか
`--yes` で同意を要求し、削除 target をレポート表示する。**profile / 世代は触らない
FS-only teardown** で、config が entry を残す限り次の apply で再配置される（transient・
project mode は lstat 検査で復帰）。恒久除去は config から entry を消して apply、
profile 完全除去は `nix-env --profile <profileDir>/profile --delete-generations`。
home / project 両モード可。

> **上は原文の写しで、規範は frontmatter が正**。原文が参照する次の規範は本 item の
> 担当ではない。
>
> - 保守的不変条件の中身 → [古くなった配置の除去は前世代の記録通りを指す symlink のみに限り、copy は消さず孤児を警告する](stale-removal-invariants.md)、空親ディレクトリ剪定の規則 → [target 除去後は空の親ディレクトリチェーンを root 境界まで保守的に剪定する](empty-parent-dir-pruning.md)。
>   本 item はそれらを reset にも適用することだけを規定する
> - project mode の lstat での確認による復帰（→ [project mode では同じ link farm なら世代を積まず drift だけを直す](../adr/project-mode-generation-skip-and-drift-repair.md)）→ [project mode は世代を非公開にし、project / system mode は derivation 同一なら世代を積まず lstat ドリフト修復だけ行う](project-mode-generation-skip.md)
> - 確認プロンプトを stdin が TTY のときだけ出し、非 TTY で同意が無ければ即エラー停止
>   すること（→ [reset は非対話の実行で --yes を必須にする](../adr/reset-confirmation-and-yes.md)）→ [reset の確認プロンプトは stdin が TTY のときだけ出し、非 TTY で同意が無ければ即エラー停止する](reset-confirm-non-tty.md)
>
> 逆に、上の写しには現れないが規範文が持つものが 1 つある。同意を取る手段の短縮形
> `-y` で、原文はこの箇条書きでは `--yes` としか書かない。短縮形はグローバルフラグ表
> （`-y, --yes`）と [reset は非対話の実行で --yes を必須にする](../adr/reset-confirmation-and-yes.md)（`--yes` / `-y`）が定めており、
> `-f` / `--file`・`-v` / `--verbose` が短縮形込みで規範化されているのに揃えて、
> 規範文では `-y` / `--yes` と書いた。
>
> **2026-10-11 追記（[reset は撤去できなかった entry を残して記録を書き直す](../adr/reset-rewrites-record-with-survivors.md)）**: 撤去対象の entry と manifest が記録する配置先は、last-applied 記録があれば記録から、無ければ
> profile の current 世代から取る。reset は profile / 世代を触らないが、撤去で target が消えた entry を除いた
> manifest で記録を書き直す（全撤去なら記録を削除する）。profile が無く記録だけがあるときは flock の鍵である
> profileDir（roothash 系列は backref `.root` も）を作る。記録の規範は [apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](last-applied-record.md) の担当。
>
> **2026-10-11 追記（[世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う](../adr/generation-backend-contract-paths-and-ints.md)）**: 上の写しの「profile 完全除去は `nix-env --profile <profileDir>/profile
> --delete-generations`」は、既定の世代バックエンド nixprofile の profile に対して利用者が叩く nix コマンドで、
> layat のサブコマンドではない（世代の間引きを nix コマンドへ委譲することは
> [世代操作は世代バックエンドの契約越しに行い、GC root の間引きと store 解放は nix コマンドへ分けて委譲する](generation-gc.md) の担当）。

名指し必須と flock は [reset は名指し必須で profileDir 単位の blocking flock を取る](reset-named-only-and-flock.md)、`--dryrun` は [reset --dryrun は副作用ゼロで削除対象を表示して終了する](reset-dryrun-no-side-effect.md) の担当。

## 出典

`docs/spec.md`「CLI 仕様」→「サブコマンド体系」の `reset <name> [target...]` の箇条書きと、
同節グローバルフラグ表の `-y, --yes`。

決定の実体は [reset で profile を触らずに配置物を撤去する](../adr/reset-fs-only-teardown.md)で、`-y` / `--yes` に
よる同意は [reset は非対話の実行で --yes を必須にする](../adr/reset-confirmation-and-yes.md) が決めている。撤去対象を last-applied 記録から優先して取り、記録を書き直すのは
[古くなった配置の除去は、比較の基準を記録から優先して読む](../adr/baseline-prefers-record.md)・[reset は撤去できなかった entry を残して記録を書き直す](../adr/reset-rewrites-record-with-survivors.md)。
