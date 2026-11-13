---
id: "REQ-38506419-c23d-4fce-b167-30715ce6a69d"
type: requirement
name: "apply --all は build 前に選択 config 間の target 衝突を検出して停止する"
description: "apply --all は build の前に、選択した config をバケットに分けて正規化後 target の重複を確かめ、重複があれば衝突 target と両 config 名を示して exit 1 で停止しなければならない"
section: cli-subcommands
status: stable
derives_from:
  - "UC-1c280dce-7c72-44c0-95ea-d06344f62a47"
specification: |
  Before any build, `apply --all` SHALL check the configs selected after the root filter
  (`--project-root` / `--home-root` / `--system-root`) for a normalized target claimed by
  more than one of them, using the `targets` obtained by the batched eval or, for the
  prebuilt set given by `--manifest <name>=<link-farm>`, derived from the entries of each
  link-farm's `manifest.json`; the prebuilt set, including the one the home-manager
  activation passes, SHALL be checked the same way. The check SHALL
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
  正規化後 target が複数の config に現れないかを確認しなければならない。`--manifest
  <name>=<link-farm>` で渡された prebuilt 集合では `targets` を各 link-farm の `manifest.json` の
  entries から導出し、home-manager の activation が渡す集合を含め同じ確認をしなければならない。
  確認は選択された
  config をバケットに分けて行わなければならない: project / home / system は rootKind ごとに
  1 つ、fixed は root 文字列値ごとに 1 つ、`--root` 指定時は全体で 1 つ。同一バケット内の
  2 つの config が正規化後 target を共有するとき、`apply --all` は衝突 target と両 config 名を
  含むエラーで停止し、exit 1 としなければならず、いずれの config の build にも配置にも
  入ってはならない。`--dryrun` も同様に停止しなければならない。選択外の config を確認しては
  ならない。名指しの `apply <name>` はこの確認を行ってはならず、そのための eval を足しては
  ならない。この確認で見えない衝突（別 entrypoint・別ツール・fixed root が偶然 project root と
  一致する場合）は、従来どおり実行時の後勝ち + foreign symlink warning で扱わなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:29:58+09:00"
---
# apply --all は build 前に選択 config 間の target 衝突を検出して停止する

## 仕様

- **検出対象**: `apply --all` で選択された config 集合内の、正規化後 target の重複。
  `targets` は一括 eval（[apply --all は rootKind を 1 回の一括 eval で取る](apply-all-batched-eval.md)）が rootKind と同時に取る。
  `--manifest <name>=<link-farm>` の prebuilt 集合（HM activation が渡すものを含む）では
  各 link-farm の `manifest.json` の entries から導出する
- **バケット規則**: project / home / system は rootKind ごと、fixed は root 文字列値ごと、
  `--root` 一律上書き時は全体 1 つ
- **error 停止**: build にも配置にも入らず exit 1（`--dryrun` も同じ）。`--json` では
  subject 登録前の失敗としてトップレベル `errors[]` に載る
- **確認しないもの**: 名指しの `apply <name>`・選択外の config
- **検出不能な衝突**: 別 entrypoint・別ツール・fixed root と project root の偶然の一致は
  従来どおり実行時の後勝ち + foreign symlink warning（[同一 target を複数 config で狙うことによる振動はユーザー責任とし warning で可視化するに留める](cross-config-target-oscillation.md) / [symlink 配置は親 dir を作り配置元/subpath を指すリンクを張り、foreign symlink は警告して後勝ちする](symlink-placement-procedure.md)）

> **上は要約で、規範は frontmatter が正**。同一 config 内の target 重複は
> [target 衝突の検出処理を同一 manifest 内と cross-config で分ける](target-collision-eval-detection.md)、HM モジュールの `layat.configs` 間の衝突は
> [単一 HM config 内の configs 間 target 衝突は activation の apply --all 前段の確認で停止する](hm-configs-target-collision-assertion.md) の担当。

## 出典

[apply --all は選択した config 間の target 重複を build 前に止める](../adr/apply-all-preflight-target-conflict.md)・[config 間の target 衝突は CLI が判定し engine に持ち込まない](../adr/cross-config-conflict-judged-by-cli.md)・[モジュールは configs 間の target 衝突を評価時に確認しない](../adr/modules-no-eval-conflict-assertion.md)
（Issue #152 で起票）。prebuilt 集合を確認の対象に含めるのは [apply --manifest でビルド済み link-farm を engine へ直接渡す](../adr/apply-manifest-prebuilt-link-farm.md)・[モジュールは configs 間の target 衝突を評価時に確認しない](../adr/modules-no-eval-conflict-assertion.md)。
