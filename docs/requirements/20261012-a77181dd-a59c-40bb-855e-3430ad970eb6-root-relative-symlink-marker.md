---
id: "REQ-a77181dd-a59c-40bb-855e-3430ad970eb6"
type: requirement
name: "mkRootRelativeSymlink は配置 root 相対の src を表すマーカーを返し、engine が target 起点の相対 symlink として張る"
derives_from:
  - "UC-01b896b4-04b9-40d0-bf9e-966eaf64c3d4"
specification: |
  `lib.mkRootRelativeSymlink` SHALL be a pure function
  (`mkRootRelativeSymlink :: string -> marker`) returning a marker that denotes an
  out-of-store symlink to a path relative to the placement root. Its argument SHALL be a
  relative path string; an absolute path and a path that escapes the root through `..`
  SHALL be rejected at Nix evaluation time. The path SHALL be resolved against the same
  root as the entry's `target` (the root the engine resolves at run time, including a
  `--root` override), and the engine SHALL place the entry as a symlink whose destination
  is the relative path from the directory of `target` to `<src>/<subpath>`. The marker
  SHALL be accepted for every root kind. Combining it with `method = "copy"` SHALL be
  rejected at evaluation time in the same way as for `mkOutOfStoreSymlink`
  (REQ-16faf428-77f3-492f-b858-222c5274cbf7), and an entry whose `<src>/<subpath>` and
  `target` are equal or one of which is an ancestor of the other SHALL be rejected at
  evaluation time as well. When `<root>/<src>/<subpath>` does not exist, the engine SHALL
  stop before placement and SHALL NOT create a dangling symlink.
specification_ja: |
  `lib.mkRootRelativeSymlink` は、配置 root 相対のパスへの out-of-store symlink を表す
  マーカーを返す純粋関数（`mkRootRelativeSymlink :: string -> marker`）でなければならない。
  引数は相対パスの文字列でなければならず、絶対パスと `..` で root の外へ出るパスは Nix
  評価時に拒否しなければならない。パスは entry の `target` と同じ root（engine が実行時に
  解決する root。`--root` 上書きを含む）に対して解決しなければならず、engine は `target` の
  親ディレクトリから `<src>/<subpath>` への相対パスを指す symlink として配置しなければ
  ならない。このマーカーは全ての root 種別で受け付けなければならない。`method = "copy"`
  との併用は `mkOutOfStoreSymlink` と同じく評価時に拒否しなければならず
  （REQ-16faf428-77f3-492f-b858-222c5274cbf7）、`<src>/<subpath>` と `target` が一致するか一方が
  他方の祖先である entry も評価時に拒否しなければならない。`<root>/<src>/<subpath>` が存在
  しないとき、engine は配置の前に停止しなければならず、dangling symlink を作っては
  ならない。
---
# REQ-a77181dd-a59c-40bb-855e-3430ad970eb6: mkRootRelativeSymlink は配置 root 相対の src を表すマーカーを返し、engine が target 起点の相対 symlink として張る

## 仕様

リポジトリ内の実体を、そのリポジトリを root とする配置先へ可変 symlink で置くためのマーカー。
`mkOutOfStoreSymlink` の引数は評価時に確定する絶対パスで、pure eval の flake は自身の
チェックアウトの絶対パスを得られない。root 相対で宣言すれば評価時に絶対パスが要らない。

```nix
".claude/skills/sara-docs" = { src = layat.lib.mkRootRelativeSymlink "dev/skills/sara-docs"; };
```

| 項目 | 内容 |
|---|---|
| 基準点 | その manifest の配置 root（`target` と同じ root。`--root` 上書きを含む）|
| リンクの形 | `target` の親ディレクトリから `<src>/<subpath>` への相対パス（上の例は `../../dev/skills/sara-docs`）|
| root 種別 | 全種別で使える |
| 評価時の拒否 | 絶対パス・`..` による root 外への脱出・`method = "copy"` との併用・`<src>/<subpath>` と `target` の一致または祖先関係 |
| 実行時の停止 | `<root>/<src>/<subpath>` が不在なら配置前に停止（dangling を作らない）|
| 版管理 | 相対のリンク先文字列のみ。指す先の内容はライブ（REQ-a8a923ad-07fb-4582-b90a-07a6e0c41baa の担当）|

`mkOutOfStoreSymlink`（REQ-eb363122-385a-499c-a074-c95efb949d07）は絶対パス専用のまま据え置き、
相対文字列を受理させない。manifest での表現（`srcKind: "rootRelative"`）は
REQ-0b0cd1e3-bfeb-45c1-978d-e2e11c568336、マーカーの判別は REQ-1dcc9a33-b0f2-43e0-8310-fc4b19e68fe7、
src と target root の時点の関係は REQ-81249072-56b8-42f5-807b-ea623c1efe99 の担当。

## 出典

ADR-0059「root 相対 out-of-store マーカー」。
