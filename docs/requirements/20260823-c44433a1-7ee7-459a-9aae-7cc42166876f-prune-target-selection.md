---
id: "REQ-c44433a1-7ee7-459a-9aae-7cc42166876f"
type: requirement
name: "prune は backref の root が実在しない roothash 系列だけを系列ごと削除する"
derives_from:
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  `layat prune` SHALL scan the user state base `<state>/nix/profiles/layat/` and the system
  base `/nix/var/nix/profiles/layat/` (or the base named by `LAYAT_SYSTEM_PROFILE_BASE`, an
  isolation seam for tests), and under each SHALL consider only the directories
  that hold a backref file `.root`. It SHALL delete a series only when the absolute root
  path that `.root` records does not exist on the filesystem. A series whose root does
  exist SHALL be left alone, whatever other sign of disuse it carries. A series it does
  delete SHALL be deleted whole — the `<roothash>` directory with every `<name>` profile,
  generation link, `.pending` and `.root` under it. A directory without `.root` — the
  `<name>`-keyed series of home mode and system mode — SHALL be out of scope structurally,
  no path being derivable for it to be judged by. A root reached through a dangling
  symlink SHALL count as not existing. Where the root path cannot be decided — the backref
  being unreadable, empty or not absolute, or the stat of the root failing for a reason
  other than non-existence — the series SHALL be kept and the reason SHALL be reported as
  a warning; a series whose deletion cannot even begin for want of privilege SHALL be kept
  the same way. A base that does not exist SHALL be treated as holding no series and
  SHALL NOT be reported; a base that exists but cannot be listed SHALL be reported as a warning naming
  it, so that a run which could not look is distinguishable from one that found nothing.
  `layat prune` SHALL NOT touch any placed artifact, and SHALL NOT thin the generations of
  a series it keeps.
specification_ja: |
  `layat prune` はユーザー state 基底 `<state>/nix/profiles/layat/` と system 基底
  `/nix/var/nix/profiles/layat/`（または `LAYAT_SYSTEM_PROFILE_BASE` が名指しする基底。
  テスト用の隔離口）を走査しなければならず、各基底の直下では backref ファイル
  `.root` を持つディレクトリだけを対象としなければならない。`.root` が記録する root の
  絶対パスが FS 上に実在しないときにのみ、その系列を削除しなければならない。root が実在
  する系列は、他にどのような不使用の兆候があっても対象にしてはならない。削除する系列は
  `<roothash>` ディレクトリごと（配下の全 `<name>` profile・世代リンク・`.pending`・
  `.root`）削除しなければならない。`.root` を持たないディレクトリ（home mode / system
  mode の `<name>` 直キー系列）は、判定すべき root パスが導けないため構造的に対象外と
  しなければならない。dangling symlink 越しの root は不在として扱わなければならない。
  root パスを決められないとき（backref が読めない・空・絶対パスでない、あるいは root の
  stat が不在以外の理由で失敗した）は系列を残し、その理由を warning として報告しなければ
  ならない。権限が足りず削除に着手すらできない系列も同じく残さなければならない。基底が存在しない
  ときはその基底に系列が無いものとして扱わなければならず、報告してはならない。基底が存在
  するのに列挙できないときは、その基底を名指しした warning として報告しなければならない
  （見に行けなかった実行と、見た結果何も無かった実行を区別できるようにするため）。
  `layat prune` は配置物に一切触れてはならず、残す系列の世代を間引いてもならない。
---
# REQ-c44433a1-7ee7-459a-9aae-7cc42166876f: prune は backref の root が実在しない roothash 系列だけを系列ごと削除する

## 仕様

**走査対象** — 2 つの基底の直下で backref `.root` を持つディレクトリ、つまり project mode /
fixed root / `--root` 上書きで生じた `<roothash>` 系列だけ。基底が 2 つあるのは、system mode の
profile 状態が `/nix/var/nix/profiles/layat/` に住み（→ ADR-0036 §3）、ユーザー state 基底
`<state>/nix/profiles/layat/` とは別の場所になるため。

`.root` を持たない `<name>` 直キーの系列（home mode の root = `$HOME`・system mode の
root = `/`）は、判定すべき root パスが導けないため構造的に対象外になる。system mode でも
`--root` を明示した系列は `<roothash>` キー + backref になるので、通常の判定に乗る。

基底そのものが無いのは正常（system mode を使ったことがない環境では system 基底が無い）で、
その基底の系列 0 件として黙って続ける。基底があるのに列挙できないときだけ報告する — 見に
行けなかったのか、見て何も無かったのかが区別できないと、非 root 実行で system 基底が読めない
状況が「孤児なし」と同じ見た目になる。

system 基底の位置は環境変数 `LAYAT_SYSTEM_PROFILE_BASE` で差し替えられる。state 基底は
`XDG_STATE_HOME` を移せば動くのに対し system 基底は絶対パスで、口が無いと破壊的な prune を
実機で駆動するテストがランナーの実共有状態を走査・削除しうるため（→ DSG-096dc893-21f4-45e3-9347-986e9275b4d1）。
隔離・テスト用であって「別の基底を prune する」機能ではない（prune は基底を引数に取らない）。

**判定条件** — 「`.root` が記録する root 絶対パスが FS 上に実在しない」の 1 条件のみ。root が
実在する系列は、entrypoint の消失など「もう使っていない」兆候があっても対象にしない
（機械判定できず誤削除リスクが高い）。dangling symlink 越しの root は不在として扱う。root
だったパスがファイル等に置き換わっていれば「実在する」側に落ちる（種別は見ない）。

**判定段で安全側へ倒す場合** — backref が読めない・空・絶対パスでない、root の stat が不在
以外の理由（権限等）で失敗する、のいずれも「削除しない + warning」に倒す。判定材料が欠けた
ときに削除へ倒す経路を持たない。

**削除段で安全側へ倒す場合** — 削除すると決めた系列でも、権限が無くて**削除に着手すらできない**
ときは残して warning を出す（→ ADR-0036 §3。system 基底の系列を非 root で実行したときに起きる）。
判定段の条件が「削除するか決められない」なのに対し、こちらは「削除すると決めたが 1 つも消せ
なかった」で発生する段が違う。どちらも残す点は同じで、1 系列を残したことが他の系列の処理を
止めることもない。

着手した後で失敗した系列は「残した」に含めない。その系列は無傷ではなく半端に壊れているので、
残した系列と同じ扱いにすると実態と食い違う報告になる。この場合の扱いは
REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed が持つ。

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

走査対象に system 基底 `/nix/var/nix/profiles/layat/` を加えることと、非 root 実行で削除権限が
無い系列を warning 付きで skip することは ADR-0036 §3 が定めている（ADR-0034 は system mode の
実装決定より前の 2026-07-04 で、当時は root = `/` が常に実在することだけを理由に対象外として
いた）。
