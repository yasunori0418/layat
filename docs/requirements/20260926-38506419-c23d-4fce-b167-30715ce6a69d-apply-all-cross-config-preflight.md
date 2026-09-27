---
id: "REQ-38506419-c23d-4fce-b167-30715ce6a69d"
type: requirement
name: "apply --all は build 前に選択 config 間の target 衝突を検出して停止する"
derives_from:
  - "UC-1c280dce-7c72-44c0-95ea-d06344f62a47"
specification: |
  Before any build, `apply --all` SHALL check the configs selected after the root filter
  (`--project-root` / `--home-root` / `--system-root`) for a normalized target claimed by
  more than one of them, using the `targets` obtained by the batched eval. The check SHALL
  group the selected configs into buckets: one per rootKind for project / home / system,
  one per root string value for fixed, and a single bucket for all of them when `--root`
  is given. When two configs in one bucket share a normalized target, `apply --all` SHALL
  stop with an error naming the conflicting target and both config names, SHALL exit 1,
  and SHALL NOT enter the build or the placement of any config; `--dryrun` SHALL stop the
  same way. Configs outside the selection SHALL NOT be checked. A named `apply <name>`
  SHALL NOT perform this check and SHALL NOT add any eval for it. Conflicts this check
  cannot see — across entrypoints, with other tools, or a fixed root that happens to equal
  the project root — SHALL remain resolved at runtime by the last writer winning with the
  foreign symlink warning.
specification_ja: |
  `apply --all` は build の前に、root フィルタ（`--project-root` / `--home-root` /
  `--system-root`）適用後に選択された config 群について、一括 eval で得た `targets` を用いて
  正規化後 target が複数の config に現れないかを検査しなければならない。検査は選択された
  config をバケットに分けて行わなければならない: project / home / system は rootKind ごとに
  1 つ、fixed は root 文字列値ごとに 1 つ、`--root` 指定時は全体で 1 つ。同一バケット内の
  2 つの config が正規化後 target を共有するとき、`apply --all` は衝突 target と両 config 名を
  含むエラーで停止し、exit 1 としなければならず、いずれの config の build にも配置にも
  入ってはならない。`--dryrun` も同様に停止しなければならない。選択外の config を検査しては
  ならない。名指しの `apply <name>` はこの検査を行ってはならず、そのための eval を足しては
  ならない。この検査で見えない衝突（別 entrypoint・別ツール・fixed root が偶然 project root と
  一致する場合）は、従来どおり実行時の後勝ち + foreign symlink warning で扱わなければならない。
---
# REQ-38506419-c23d-4fce-b167-30715ce6a69d: apply --all は build 前に選択 config 間の target 衝突を検出して停止する

## 仕様

- **検出対象**: `apply --all` で選択された config 集合内の、正規化後 target の重複。
  `targets` は一括 eval（REQ-535b811d-dfc5-4eac-92db-737e70eb5415）が rootKind と同時に取る
- **バケット規則**: project / home / system は rootKind ごと、fixed は root 文字列値ごと、
  `--root` 一律上書き時は全体 1 つ
- **error 停止**: build にも配置にも入らず exit 1（`--dryrun` も同じ）。`--json` では
  subject 登録前の失敗としてトップレベル `errors[]` に載る
- **検査しないもの**: 名指しの `apply <name>`・選択外の config
- **検出不能な衝突**: 別 entrypoint・別ツール・fixed root と project root の偶然の一致は
  従来どおり実行時の後勝ち + foreign symlink warning（REQ-fc1118b1-b0e8-4ddf-80f6-c70956651693 / REQ-622787dc-4512-4ce9-9c7d-7b32bbb70557）

> **上は要約で、規範は frontmatter が正**。同一 config 内の target 重複は
> REQ-5c6b07da-3d06-414d-8770-4f438234b322、HM モジュールの `layat.configs` 間の衝突は
> REQ-5923ac79-4a2d-43cd-b56c-2f1000c01b44 の担当。

## 出典

ADR-0038「同一 entrypoint 内の cross-config target 衝突を `apply --all` 前段で検出し error 停止する」
§1〜§3（Issue #152 で起票）。
