---
id: "REQ-c44433a1-7ee7-459a-9aae-7cc42166876f"
type: requirement
name: "prune は backref の root が実在しない roothash 系列だけを系列ごと削除する"
specification: |
  `nput prune` SHALL scan only the directories directly under
  `<state>/nix/profiles/nput/` that hold a backref file `.root`, and SHALL delete a series
  only when the absolute root path that `.root` records does not exist on the filesystem.
  A series whose root does exist SHALL be left alone, whatever other sign of disuse it
  carries. A series it does delete SHALL be deleted whole — the `<roothash>` directory with
  every `<name>` profile, generation link, `.pending` and `.root` under it. A directory
  without `.root` — the `<name>`-keyed series of home mode and system mode — SHALL be out
  of scope structurally, no path being derivable for it to be judged by. A root reached
  through a dangling symlink SHALL count as not existing. Where the root path cannot be
  decided — the backref being unreadable, empty or not absolute, or the stat of the root
  failing for a reason other than non-existence — the series SHALL be kept and the reason
  SHALL be reported as a warning. `nput prune` SHALL NOT touch any placed artifact, and
  SHALL NOT thin the generations of a series it keeps.
specification_ja: |
  `nput prune` は `<state>/nix/profiles/nput/` 直下のうち backref ファイル `.root` を持つ
  ディレクトリだけを走査しなければならず、`.root` が記録する root の絶対パスが FS 上に
  実在しないときにのみ、その系列を削除しなければならない。root が実在する系列は、他に
  どのような不使用の兆候があっても対象にしてはならない。削除する系列は `<roothash>`
  ディレクトリごと（配下の全 `<name>` profile・世代リンク・`.pending`・`.root`）削除
  しなければならない。`.root` を持たないディレクトリ（home mode / system mode の
  `<name>` 直キー系列）は、判定すべき root パスが導けないため構造的に対象外としなければ
  ならない。dangling symlink 越しの root は不在として扱わなければならない。root パスを
  決められないとき（backref が読めない・空・絶対パスでない、あるいは root の stat が
  不在以外の理由で失敗した）は系列を残し、その理由を warning として報告しなければ
  ならない。`nput prune` は配置物に一切触れてはならず、残す系列の世代を間引いても
  ならない。
derives_from:
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
---
# REQ-c44433a1-7ee7-459a-9aae-7cc42166876f: prune は backref の root が実在しない roothash 系列だけを系列ごと削除する

## 仕様

**走査対象** — `<state>/nix/profiles/nput/` 直下で backref `.root` を持つディレクトリ、つまり
project mode / fixed root / `--root` 上書きで生じた `<roothash>` 系列だけ。`.root` を持たない
`<name>` 直キーの系列（home mode の root = `$HOME`・system mode の root = `/`）は、判定すべき
root パスが導けないため構造的に対象外になる。

**判定条件** — 「`.root` が記録する root 絶対パスが FS 上に実在しない」の 1 条件のみ。root が
実在する系列は、entrypoint の消失など「もう使っていない」兆候があっても対象にしない
（機械判定できず誤削除リスクが高い）。dangling symlink 越しの root は不在として扱う。

**安全側へ倒す場合** — backref が読めない・空・絶対パスでない、root の stat が不在以外の理由
（権限等）で失敗する、のいずれも「削除しない + warning」に倒す。判定材料が欠けたときに削除へ
倒す経路を持たない。

**削除の単位** — 対象になった系列は `<roothash>` ディレクトリごと消す。配下に世代が 1 つも無く
`.pending` と `.root` だけが残っている系列も同じく対象になる。

**触らないもの** — 配置物には一切触れない（root が消えている以上、配置先も消えている）。残す
系列の世代間引きも行わない（`nix-env --profile <dir> --delete-generations` の二重化になる）。
store の回収は `nix-collect-garbage` に委ねる。

> 孤児 profile が backref で逆引き可能なまま残ること自体は REQ-d41b1d0a-c6d5-41cc-93f9-e5cc7f152da4、
> backref `.root` を roothash 階層へ置くことは REQ-2aa3abbc-90b2-486e-92de-d785554bdeb3、
> profile を解決済み root でキーすること（孤児が生じる前提）は REQ-46fccb80-4bae-4d37-bc19-dded88e9a9c0 の担当。
> 削除を守る dryrun・確認・lock・`--json` の規律は REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed が持つ。

## 出典

決定の実体は ADR-0034 §1「削除対象 = 「root パスが実在しない」roothash 系列のみ」。
out-of-store な root がアンマウント中に「一時的に実在しない」と誤判定されうる caveat は
同 §2 にあり、機械側で区別せず確認プロンプトの root パス一覧で防ぐ（→ REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed）。
