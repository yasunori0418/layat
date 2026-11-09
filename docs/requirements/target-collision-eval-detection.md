---
id: "REQ-5c6b07da-3d06-414d-8770-4f438234b322"
type: requirement
name: "target 衝突の検出処理を同一 manifest 内と cross-config で分ける"
status: stable
derives_from:
  - "UC-1c280dce-7c72-44c0-95ea-d06344f62a47"
specification: |
  A collision in which two distinct keys A and B explicitly override the `target` field to
  the same value SHALL be detected as a duplicate of the normalized target string and
  SHALL stop evaluation via `lib.throwIf` at eval time, not at engine runtime. A
  cross-config collision that does not ride on a single invocation (a same target across
  separate entrypoints, machines or tools) cannot be detected before placement and SHALL
  instead be handled at engine runtime as last-writer-wins plus a foreign symlink warning.
  The two SHALL remain separate paths. Where several configs do ride on a single
  invocation, as the `layat.configs` of one module configuration do through the `apply
  --all` of the activation, detection before placement is possible and SHALL NOT be
  precluded by this; that case is stated by [単一 HM config 内の configs 間 target 衝突は activation の apply --all 前段の確認で停止する](hm-configs-target-collision-assertion.md) and is
  not restated here.
specification_ja: |
  別キー A / B が `target` フィールドを同値に明示上書きした衝突は、正規化後 target
  文字列の重複として eval 時に `lib.throwIf` で検出し停止しなければならない
  （engine 実行時ではない）。単一の起動に載らない cross-config（別
  entrypoint・別マシン・別ツールに跨る場合）の同一 target 衝突は配置前に検出できず、engine
  実行時の後勝ち +
  foreign symlink warning として扱わなければならない。両者は別の処理として保たなければ
  ならない。単一の起動に載る複数 config
  （activation の `apply --all` に載る 1 つのモジュール config の `layat.configs` など）に
  ついては配置前の検出が可能であり、本 item はそれを妨げない。その場合の規範は [単一 HM config 内の configs 間 target 衝突は activation の apply --all 前段の確認で停止する](hm-configs-target-collision-assertion.md) の担当で、本 item では
  規定しない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T09:19:35+09:00"
---
# target 衝突の検出処理を同一 manifest 内と cross-config で分ける

## 仕様

属性キーによる一意性担保だけでは塞げない場合として、**別キー A/B が `target` フィールドを
同値に明示上書きした衝突**がある。これは正規化後 target 文字列の重複として eval 時に
`lib.throwIf` で検出・停止する（engine 実行時ではない）。

cross-config（別 profile・別 manifest）の同一 target 衝突は eval では検出不可で、
これは engine 実行時の後勝ち + foreign symlink warning になる。両者は扱う段階が別である。

> **上は原文の写しで、規範は frontmatter が正**。原文が「cross-config」を無条件に「eval では
> 検出不可」とするのに対し、規範文では**単一の eval に載らない場合**（別 entrypoint・別マシン・
> 別ツール）へ限定した。**旧決定（[モジュールは configs 間の target 衝突を評価時に確認しない](../adr/modules-no-eval-conflict-assertion.md) で改訂）が、1 つのモジュール config 内の `layat.configs.<A>` と
> `<B>` は全 config が同一のモジュール eval に載るため正規化後 target の衝突を静的に検出でき、
> eval 時 assertion で停止すると決定済み**で、原文の無条件な言い切りはこれを否定してしまう。
> 原文が旧決定に未追従なので、[rootKind に system を足しても schemaVersion を上げない](../adr/system-rootkind-without-version-bump.md)・[root は 3 つの marker で表す](../adr/three-root-markers.md) 由来の未追従を [root は 3 マーカーと絶対パス文字列の union を取る](root-marker-kinds.md) / [意図が矛盾する組み合わせをクロスフィールドの確認で評価時に拒否する](normalize-manifest-cross-field-checks.md) で扱ったのと
> 同じように扱う。本 item は静的検出が可能な場合を妨げないことまでを規範とし、その場合に実際に
> eval 停止する規範（旧決定）そのものは持たない。**その規範は [単一 HM config 内の configs 間 target 衝突は activation の apply --all 前段の確認で停止する](hm-configs-target-collision-assertion.md) が
> 持つ**。`docs/spec.md` の追従は本 item の担当範囲外。
>
> **2026-10-15 追記（[モジュールは configs 間の target 衝突を評価時に確認しない](../adr/modules-no-eval-conflict-assertion.md)）**: 上の「eval 時 assertion で停止」は、activation の `apply --all`
> 前段の確認で停止する形へ改められた。本 item が妨げないとする対象も「単一の eval」から「単一の
> 起動」へ読み替える（規範は frontmatter）。

## 出典

`docs/spec.md`「lib API」→「入力の確認（`evalModules` + `normalizeManifest`）」。

決定の実体は同一 manifest 内の target 衝突を eval 時に検出すると定めた
[manifest 内の target の衝突を評価時に検出する](../adr/manifest-target-conflict-eval-error.md)で、cross-config を「単一の eval に載らない場合」へ限定した
のは旧決定（単一のモジュール eval に載る config 間は静的検出が可能）。例外の条件を「単一の起動」へ改めたのは
[モジュールは configs 間の target 衝突を評価時に確認しない](../adr/modules-no-eval-conflict-assertion.md)。
