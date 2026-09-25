---
id: "REQ-91f0a9a7-c2b2-4cda-8cce-cbf5d8c5d04d"
type: requirement
name: "apply --all は build を --jobs の worker pool で先行並列実行し build 失敗 config を部分失敗として集約する"
derives_from:
  - "UC-1c280dce-7c72-44c0-95ea-d06344f62a47"
specification: |
  After the cross-config conflict preflight, `apply --all` SHALL process the selected configs
  in two stages. Stage 1 SHALL realize every selected config's build read-only (without an
  out-link gcroot) on a worker pool that runs at most N builds at once, where N is the
  resolved value of `--jobs`; stage 2 SHALL then apply the configs in parallel as specified
  by REQ-4cbd9a0d-9f94-4747-8881-56020dc6d5af, presenting the aggregated display and
  `results[]` in lexical order, with the build still performed inside the engine's lock. `--jobs` SHALL be a flag
  local to `apply`, SHALL default to 0, which resolves to the logical CPU count, SHALL be
  rejected with an error when negative, and SHALL be rejected with an error when given
  without `--all`. A config whose stage-1 build fails SHALL NOT enter stage 2, SHALL be
  settled on its own subject with that error, SHALL count as a failure of the aggregate,
  and SHALL be reported on stderr, while the other configs SHALL proceed; `--dryrun` SHALL
  take the same two stages. The aggregate result (applied / skipped / failed counts and the
  order and content of `results[]`) SHALL NOT depend on the value of `--jobs`. Under
  `--debug`, the stage-1 nix command disclosure lines SHALL be prefixed with `[<name>] ` and
  MAY appear in any order, since the builds run concurrently; nix's own output in stage 1
  is captured and SHALL surface only in the error of a failed config.
specification_ja: |
  `apply --all` は前段の cross-config 衝突検査の後、選択された config を 2 段で処理しなければ
  ならない。stage 1 は選択された全 config の build を読み取り専用（out-link gcroot を置かない）
  で先行 realize し、同時実行数が `--jobs` の解決値 N を超えない worker pool で実行しなければ
  ならない。stage 2 はその後 REQ-4cbd9a0d-9f94-4747-8881-56020dc6d5af に従い config を並列に
  適用し、集約表示と `results[]` を辞書順にしなければならず、build は従来どおり engine の
  ロック内で行わなければならない。`--jobs` は `apply` ローカルのフラグでなければ
  ならず、既定値 0 は論理 CPU 数に解決されなければならず、負値はエラーとしなければならず、
  `--all` 無しの指定はエラーとしなければならない。stage 1 の build に失敗した config は
  stage 2 に入ってはならず、その error で自身の subject を確定し、集約の失敗に数え、stderr に
  報告しなければならず、他の config は処理を続けなければならない。`--dryrun` も同じ 2 段を
  取らなければならない。集約結果（applied / skipped / failed の件数と `results[]` の順序・
  内容）は `--jobs` の値に依存してはならない。`--debug` 下では、stage 1 の nix コマンド開示行に
  `[<name>] ` の prefix を付けなければならず、build が並行するため開示行はどの順で現れても
  よい。stage 1 の nix 自身の出力は捕捉され、失敗した config の error にのみ現れなければ
  ならない。
---
# REQ-91f0a9a7-c2b2-4cda-8cce-cbf5d8c5d04d: apply --all は build を --jobs の worker pool で先行並列実行し build 失敗 config を部分失敗として集約する

## 仕様

- **2 段構成**: stage 1 = 選択 config の build を `--no-link` で先行 realize（worker pool・
  同時実行数 ≤ `--jobs`）、stage 2 = REQ-4cbd9a0d-9f94-4747-8881-56020dc6d5af に従う並列適用
  （集約表示と `results[]` は辞書順）。engine の in-lock build は残り、
  stage 1 で store が温まっているのでキャッシュヒットになる
  （build をロック内に閉じる REQ-60c6b7ea-e936-4ce8-bd75-ad35e9c693b9 は不変）
- **`--jobs`**: `apply` ローカル（persistent ではない）・既定 0 = 論理 CPU 数・負値はエラー・
  `--all` 無しはエラー（明示の `--jobs 0` も含む）
- **build 失敗**: その config は stage 2 をスキップし、自身の subject に error で載り、失敗に
  数えられる。成功 config の subject は残る（REQ-059eb4d5-63fb-4f8e-b705-11b5e2ed4ae5）
- **決定性**: 集約結果は `--jobs` に依存しない。stage 1 で非決定なのは `--debug` 開示行の
  順序だけで、stage 1 の nix 自身の出力は捕捉されて失敗時の error に載る（stage 2 の
  即時出力の順序は REQ-4cbd9a0d-9f94-4747-8881-56020dc6d5af）

> **上は要約で、規範は frontmatter が正**。配置段階の並列化（Issue #154）の規範は
> REQ-4cbd9a0d-9f94-4747-8881-56020dc6d5af が持つ。

## 出典

ADR-0039「`apply --all` を build・配置の両段階で並列化する」の build 段階（Issue #153 で起票）。
