---
id: "REQ-d2277c7a-7992-49af-a9dc-4cc73843a6f9"
type: requirement
name: "copy は target 不在のときだけマテリアライズする place-once で世代管理の対象外とする"
section: placement-copy
status: stable
derives_from:
  - "UC-403fbe32-b146-401b-8b53-fe67c1e169c5"
specification: |
  A copy entry SHALL be materialized only while the target is absent. When the subpath
  denotes a directory and the target is absent, `<root>/<target>` SHALL be created and the
  contents of `<src>/<subpath>/` copied into it recursively through native filesystem
  operations; when the subpath denotes a file and the target is absent, the parent
  directory SHALL be created and `<src>/<subpath>` copied to `<root>/<target>`. Where the
  target already exists, nothing SHALL be done, the object being left to the user's own
  management; this is place-once, meaning that after the first materialization a present
  target SHALL NOT be touched. Reflecting an update of the store SHALL therefore be done
  either by `apply --recopy` or by removing the object with `reset` and applying again. A
  structural mismatch — a regular file existing at the target while the subpath is a
  directory, or the target being a directory while the subpath is a file — SHALL stop with
  an error. Copies SHALL be outside generation management: a rollback SHALL NOT restore the
  contents of a copy destination to those of the generation being returned to, place-once
  applying to a rollback exactly as it does to an apply, so that an existing copy
  destination SHALL be left untouched. Where the destination is absent at that point — a
  copy child beneath an ancestor symlink removed by the pre-placement removal, or a target
  unlinked by a symlink-to-copy method change — the rollback SHALL materialize it
  place-once, as an apply would. Reflecting the contents of another generation SHALL be
  left to the user's own `apply --recopy`.
specification_ja: |
  copy entry は target が不在のときにだけマテリアライズしなければならない。subpath が
  ディレクトリで target が不在のときは `<root>/<target>` を作成して `<src>/<subpath>/` の内容を
  ネイティブ FS 操作で再帰コピーしなければならず、subpath がファイルで target が不在のときは
  親ディレクトリを作成して `<src>/<subpath>` を `<root>/<target>` へコピーしなければならない。
  target が既に存在するときは何もせずユーザー管理に委ねなければならない（place-once。初回
  マテリアライズ後、target が在れば触ってはならない）。したがってストア更新の反映は
  `apply --recopy`、または `reset` で撤去後に再 apply で行わなければならない。
  構造不一致（subpath がディレクトリのとき target に通常ファイルが存在する / subpath がファイルの
  とき target がディレクトリである）はエラーで停止しなければならない。copy は世代管理の対象外で
  なければならない。rollback は copy 先の内容を戻り先世代の内容へ復元してはならず、place-once は
  apply と同様に rollback にも適用されるため、既存の copy 先に触れてはならない。その時点で配置先が
  不在のとき（配置前除去が祖先 symlink を除去した配下の copy child・symlink → copy の method
  change で unlink された target）は、apply と同じく place-once でマテリアライズしなければ
  ならない。他世代の内容の反映はユーザー自身の `apply --recopy` に委ねなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# copy は target 不在のときだけマテリアライズする place-once で世代管理の対象外とする

## 仕様

```
subpath がディレクトリの場合:
  target が不在のとき: <root>/<target> を作成しネイティブ再帰コピー（<src>/<subpath>/ → <root>/<target>/）
  target が存在するとき: 何もしない（ユーザー管理に委ねる）

subpath がファイルの場合:
  target が不在のとき: 親ディレクトリを作成しネイティブコピー（<src>/<subpath> → <root>/<target>）
  target が存在するとき: 何もしない
```

- **place-once**: 初回マテリアライズ後、target が在れば触らない。ストア更新の反映は
  **`apply --recopy`**、または `reset <name> [target]` で撤去後に再 apply で行う
- `subpath` がディレクトリのとき `target` に通常ファイルが存在する場合、または `subpath` が
  ファイルのとき `target` がディレクトリの場合は構造不一致でエラーで停止する
- 世代管理の対象外。**rollback は copy 先の内容を戻り先世代へ復元しない**（place-once は
  rollback にも適用され、既存の copy 先には触らない）。配置先が不在なら apply と同じく
  place-once で配置される。他世代の内容を反映したいときはユーザーが `apply --recopy` を打つ

> **上は原文の写しで、規範は frontmatter が正**。原文が併記する次の規範は本 item の
> 担当ではない。
>
> - コピー時の mode 保存 + owner-write 付与 → [copy は元の mode を保存しつつ owner-write を付与する](copy-mode-owner-write.md)
> - src ツリー内 symlink の複製 → [copy は src ツリー内の symlink を deref せず symlink のまま複製する](copy-preserves-inner-symlinks.md)
> - foreign 実ファイルの skip を warning で可視化すること → [copy が foreign 実ファイルを skip したときは warning で可視化する](copy-foreign-skip-warning.md)
> - `--backup` 有効時に構造不一致が退避 + 配置へ変わること → [apply --backup は配置を塞ぐ記録外実体を rename 退避してから配置する](apply-backup-rename-aside.md) / [backup 退避は配置前除去の後・配置の前に置き、drift 修復の処理でも同じく実施する](backup-stage-position.md)
> - `apply --recopy` の契約 → [apply --recopy は config 内の全 copy target を src から無条件に上書き再コピーする](apply-recopy.md)、`reset` の契約 → [reset は profile を触らない FS-only teardown で配置物を無い状態へ戻す](reset-fs-only-teardown.md)
> - method の値が copy であることの意味と世代管理の有無 → [method は配置方法を選び symlink は世代管理下・copy は世代管理外になる](method-src-matrix.md)
> - `rollback` の段構成（copy 反映段を含む 4 段）そのもの → [rollback は FS を先に収束させてから profile ポインタを最後に移す](rollback-refit-then-pointer.md)

## 出典

`docs/spec.md`「配置動作仕様」→「copy モード（place-once・ユーザー管理）」節のコードブロックと、
箇条書き第 2・3・6 項。

決定の実体は [apply --recopy で全 copy target を src から無条件に上書きする](../adr/apply-recopy-overwrites-copies.md)で、place-once そのものは [copy は世代の外に place-once で置く](../adr/copy-outside-generations-place-once.md) が定めている。
