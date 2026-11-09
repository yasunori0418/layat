---
id: "REQ-fc1118b1-b0e8-4ddf-80f6-c70956651693"
type: requirement
name: "同一 target を複数 config で狙うことによる振動はユーザー責任とし warning で可視化するに留める"
status: stable
derives_from:
  - "UC-1c280dce-7c72-44c0-95ea-d06344f62a47"
specification: |
  Where two different configs aim at the same target, the drift repair may oscillate on
  every re-entry of the `shellHook` — A places it, B detects it as foreign and takes it,
  A takes it back — which is an active oscillation rather than a single instance of the
  last writer winning. Not aiming at the same target from several configs SHALL be the
  responsibility of the user, and layat SHALL make the situation visible through the
  foreign symlink warning and SHALL NOT hold any mechanism that detects and stops it at
  engine runtime. Where several configs ride on a single invocation, as the `layat.configs`
  of one module configuration do through the `apply --all` of the activation, a detection
  before placement is possible and SHALL NOT be precluded by this; that case is stated by [単一 HM config 内の configs 間 target 衝突は activation の apply --all 前段の確認で停止する](hm-configs-target-collision-assertion.md) and is not restated here.
  As an exception, a conflict between configs selected by one `apply --all` of a single
  entrypoint SHALL be stopped before the build by the preflight of [apply --all は build 前に選択 config 間の target 衝突を検出して停止する](apply-all-cross-config-preflight.md).
  The foreign warning during an oscillation will keep appearing under the high frequency
  at which a `shellHook` runs, and this SHALL be regarded as correct, being the signal of
  a misconfiguration; the warning SHALL be outside the scope of silence on success and
  SHALL therefore appear regardless of `-v`. The MVP SHALL NOT hold any mechanism for
  suppressing or aggregating it, the situation being resolved by removing the duplicate
  target from the configs.
specification_ja: |
  別 config が同一 target を狙うと、ドリフト修復が「A が置く → B が foreign 検知して奪う →
  A が再奪取」と `shellHook` 再入のたびに振動しうる（単発の後勝ちではなく能動的オシレーション）。
  「同一 target を複数 config で狙わない」はユーザー責任とし、layat は foreign symlink warning で
  可視化するに留め、engine 実行時に検知して止める機構を持ってはならない。単一の起動に載る
  複数 config（activation の `apply --all` に載る 1 つのモジュール config の `layat.configs` など）
  については配置前の検出が可能であり、本 item はそれを妨げない。その場合の規範は [単一 HM config 内の configs 間 target 衝突は activation の apply --all 前段の確認で停止する](hm-configs-target-collision-assertion.md) の担当で、本 item
  では規定しない。例外として、同一 entrypoint の `apply --all` で選択された config 間の衝突は
  [apply --all は build 前に選択 config 間の target 衝突を検出して停止する](apply-all-cross-config-preflight.md) の前段の確認で build 前に止まらなければならない。振動中の foreign warning は
  `shellHook` の高頻度実行で出続けるが、これは設定ミスのシグナルとして正しいものと
  みなさなければならない。この warning は成功時沈黙の対象外でなければならず、`-v` の有無に
  関わらず常時出さなければならない。MVP では抑制 / 集約機構を持ってはならず、
  config の同一 target 重複を解消して直す。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:31:01+09:00"
---
# 同一 target を複数 config で狙うことによる振動はユーザー責任とし warning で可視化するに留める

## 仕様

- **cross-config 同一 target の振動はユーザー責任**: 別 config A / B が同一 target を狙うと、
  lstat 修復が「A が置く → B が foreign 検知して奪う → A が再奪取」と `shellHook` 再入のたびに
  振動しうる（単発「後勝ち」ではなく能動的オシレーション）。「同一 target を複数 config で
  狙わない」をユーザー責任とし、foreign symlink warning で可視化する。検知して止める機構は
  持たない
- **振動中の foreign warning は `shellHook` 高頻度実行で出続けるが、これは設定ミスのシグナルと
  して正しい**。warning は成功時沈黙の対象外なので `-v` の有無に関わらず常時出る。抑制 / 集約
  機構は MVP で持たず、config の同一 target 重複を解消して直す（document-only）

> **上は原文の写しで、規範は frontmatter が正**。foreign symlink warning そのものと単発の
> 後勝ちは [symlink 配置は親 dir を作り配置元/subpath を指すリンクを張り、foreign symlink は警告して後勝ちする](symlink-placement-procedure.md)、lstat ドリフト修復そのものは [project mode は世代を非公開にし、project / system mode は derivation 同一なら世代を積まず lstat ドリフト修復だけ行う](project-mode-generation-skip.md)、warning が成功時沈黙の
> 対象外であることの出力規律そのものは [成功時はデフォルト沈黙とし warning と error は常時 stderr へ出す](silent-on-success.md) の担当。同一 config 内での target 衝突を
> 評価時に検出することは [target 衝突の検出処理を同一 manifest 内と cross-config で分ける](target-collision-eval-detection.md) の担当（本 item は config をまたぐ場合を扱う）。
>
> **「検知して止める機構を持たない」を engine 実行時へ限定した理由**: **旧決定（[モジュールは configs 間の target 衝突を評価時に確認しない](../adr/modules-no-eval-conflict-assertion.md) で改訂）が、
> 1 つのモジュール config 内の `layat.configs.<A>` と `<B>` は全 config が同一のモジュール eval に
> 載るため正規化後 target の衝突を静的に検出でき、eval 時 assertion で停止すると決定済み**で、
> 原文の無条件な言い切りはこれを否定してしまう（原文が旧決定に未追従・[target 衝突の検出処理を同一 manifest 内と cross-config で分ける](target-collision-eval-detection.md) と
> 同じ扱い）。本 item は静的検出が可能な場合を妨げないことまでを規範とし、その場合に実際に
> eval 停止する規範そのものは持たない（**その規範は [単一 HM config 内の configs 間 target 衝突は activation の apply --all 前段の確認で停止する](hm-configs-target-collision-assertion.md) が持つ**）。
> `docs/spec.md` の追従は本 item の担当範囲外。
>
> **2026-10-15 追記（[モジュールは configs 間の target 衝突を評価時に確認しない](../adr/modules-no-eval-conflict-assertion.md)）**: 上の「eval 時 assertion で停止」は、activation の `apply --all`
> 前段の確認で停止する形へ改められた（規範は frontmatter・[単一 HM config 内の configs 間 target 衝突は activation の apply --all 前段の確認で停止する](hm-configs-target-collision-assertion.md)）。

## 出典

`docs/spec.md`「世代管理仕様」→「project mode の世代」節の、世代スキップ時の FS の確認に付随する
入れ子箇条書き 2 項。

決定の実体は [配置の振動はユーザー責任とし、文書で注意するに留める](../adr/oscillation-is-user-responsibility.md)（cross-config 同一 target をユーザー責任と
すること・振動 warning を抑制しないこと）。
「検知して止める機構を持たない」を engine 実行時へ限定したのは旧決定（単一のモジュール eval に
載る config 間は静的検出が可能）。例外の条件を「単一の起動」へ改めたのは
[モジュールは configs 間の target 衝突を評価時に確認しない](../adr/modules-no-eval-conflict-assertion.md)。
