---
id: "REQ-9d622539-a1e5-4719-af51-e0cd85b66f35"
type: requirement
name: "apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む"
section: generations
status: stable
derives_from:
  - "UC-0b6f60cb-3e98-4ee7-8929-4d94a29f0af6"
specification: |
  After every successful placement the engine SHALL write a last-applied record at
  `<state>/layat/<key>/last-applied.json` for home, project and fixed configs, and at
  `<SystemStateDir>/layat/<key>/last-applied.json` (`<SystemStateDir>` defaulting to
  `/var/lib`) for system configs, where `<key>` is the same key as the profileDir
  (`<name>` for home and system without `--root`, otherwise `<roothash>/<name>` with the
  backref `.root`).
  The record SHALL contain `recordVersion` (2), `linkFarm` (the store path of the placed
  link-farm), `appliedAt` (RFC 3339), `manifest` (the placed `manifest.json` embedded
  verbatim) and, when known, `origin` (the source of the most recent apply that wrote the record, as specified by
  [apply は last-applied 記録に出所を持たせ、記録と今回の出所が食い違えば警告して続行する](apply-origin-mismatch-warning.md)), so that its content does not depend on the store. The record SHALL be written
  atomically (temporary file + rename) immediately after all filesystem operations have
  succeeded and before the generation commit; a failure to write it SHALL fail the apply,
  and the undo journal SHALL roll the filesystem back. The record file SHALL be written with
  mode 0644 so that every user can read it (read-only operations such as `apply --dryrun` do
  not require the writer's privileges). The record SHALL also be updated on
  the generation-skip path, under the same failure rule as apply. On rollback the record
  SHALL be rewritten with the manifest and link-farm of N-1 after the filesystem has
  converged and before the profile pointer is moved; a failure to write it SHALL fail the
  rollback, and its undo journal SHALL roll the filesystem back. On reset the record SHALL
  be rewritten with the manifest minus the selected entries whose target is gone (entries
  kept as foreign or drifted and entries left unreached by a failure stay in the record)
  keeping `linkFarm` as it was (when the baseline came from the profile because no record
  existed, `linkFarm` SHALL be the link-farm of that current generation), and the record
  file SHALL be deleted when every entry has been removed. Reset SHALL rewrite the record
  by the same rule also when a filesystem operation fails part-way, returning a failure to
  rewrite it together with the removal error; a failure to write it SHALL fail the reset
  (reset has no undo journal, so the filesystem stays as reset left it). When no profile
  exists but a record does, reset SHALL create the profileDir that is the flock key (and,
  for the roothash series, the backref `.root`) in the same way as apply before taking the
  lock and removing. `appliedAt` SHALL
  be the time of the operation that wrote the record. A record that is not valid JSON,
  that lacks a required field, has a field of the wrong type or contains an unknown field,
  or whose `recordVersion` is neither 1 nor 2 SHALL be an error rather than being treated as
  absent; a version 1 record SHALL be read as having no `origin`. The baseline for stale
  removal, for drift repair, for reset's removal set and for rollback's departing
  generation SHALL be taken from the record when it exists, otherwise from the current
  generation of the nix profile; when both exist and their link-farm disagree, the record
  SHALL win and one warning line SHALL be written; when neither exists the baseline SHALL
  be empty and one warning line stating that there is no last-applied record SHALL be
  written (warnings go to stderr under the output discipline of
  [成功時はデフォルト沈黙とし warning と error は常時 stderr へ出す](silent-on-success.md)). The generation skip SHALL be taken only when
  the new link-farm equals the current generation of the profile and, when a record exists,
  also its `linkFarm`; it SHALL NOT be taken when no profile exists, even if a record does;
  when the record and the profile disagree, the full apply path SHALL run with the record as
  baseline, so that a re-run after a failed commit still commits and entries placed by the
  uncommitted attempt are removed. The rollback target (N-1)
  SHALL still be read from the generation link. The `--json` contract (apart from the
  origin mismatch warning of [apply は last-applied 記録に出所を持たせ、記録と今回の出所が食い違えば警告して続行する](apply-origin-mismatch-warning.md)), the flock key
  (profileDir) and the prune targets SHALL NOT change.
specification_ja: |
  engine は配置が成功するたびに last-applied 記録を、home / project / fixed の config では
  `<state>/layat/<鍵>/last-applied.json` に、system の config では
  `<SystemStateDir>/layat/<鍵>/last-applied.json`（`<SystemStateDir>` の既定は `/var/lib`）に
  書かなければならない。`<鍵>` は profileDir と同じ鍵（`--root` なしの home と system は
  `<name>` 直、他は `<roothash>/<name>` + backref `.root`）とする。記録は `recordVersion`（2）・`linkFarm`
  （配置した link-farm の store path）・`appliedAt`（RFC 3339）・`manifest`（配置した
  `manifest.json` をそのまま埋め込む）と、分かっていれば `origin` を含まなければならず、内容が
  store に依存しないようにする。`origin` は直近に記録を書いた apply の出所で、規則は
  [apply は last-applied 記録に出所を持たせ、記録と今回の出所が食い違えば警告して続行する](apply-origin-mismatch-warning.md) が定める。記録は全 FS 操作の成功直後・世代 commit の前に、temp + rename で atomic に
  書かなければならず、書けなければ apply は失敗し undo journal が FS を巻き戻さなければ
  ならない。記録ファイルは全ユーザーが読める mode 0644 で書かなければならない（`apply --dryrun`
  などの読み取り専用の操作に書き手の権限を要さないため）。generation skip の処理でも apply と同じ失敗規則で更新しなければならない。rollback
  では FS の収束後・profile ポインタ移動の前に N-1 の manifest と link-farm で書き直さなければ
  ならず、書けなければ rollback は失敗し undo journal が FS を巻き戻さなければならない。reset
  では撤去対象のうち target が消えた entry だけを除いた manifest で書き直さなければならず
  （foreign・drift として残した entry と途中失敗で未到達の entry は記録に残り、`linkFarm` は
  保持する。記録が無く比較の基準を profile から取ったときはその current 世代の link-farm を
  `linkFarm` とする）、全撤去なら記録ファイルを削除しなければならない。reset は FS 操作の途中
  失敗でも同じ規則で記録を書き直さなければならず、書き直しの失敗は撤去の error と併せて返さ
  なければならない。書けなければ reset は失敗しなければならない（reset は undo journal を
  持たないので FS は reset が残した状態のまま）。profile が無く記録だけがあるとき、reset は
  flock の鍵である profileDir（roothash 系列は backref `.root` も）を apply と同じ手順で
  作ってから flock を取り、撤去しなければならない。`appliedAt` は記録を書いた操作の時刻としなければならない。
  JSON として不正な記録・必須フィールドの欠落や型違いの記録・未知のフィールドを含む記録・
  `recordVersion` が 1 でも 2 でもない記録は
  「記録なし」と扱わず error としなければならない。版 1 の記録は `origin` 無しとして読まなければならない。古くなった配置の除去・drift 修復・reset の撤去
  対象の比較の基準と、rollback で離れる基準世代は、記録があれば記録から、無ければ nix profile の
  current から取らなければならない。両方あって link-farm が食い違えば記録を採り warning を
  1 行出さなければならない。どちらも無ければそれらは空とし、last-applied 記録が無い旨を
  warning 1 行として出さなければならない（warning は [成功時はデフォルト沈黙とし warning と error は常時 stderr へ出す](silent-on-success.md)
  の出力規律に従い stderr へ出る）。generation skip は、新 link-farm が profile の current 世代に
  一致し、記録があればその `linkFarm` にも一致するときだけ行わなければならない。profile が
  無ければ（記録だけがあっても）skip してはならない。記録と profile が食い違うときは記録を比較の基準に通常の apply の処理を
  走らせなければならない（commit 失敗後の再実行が commit を飛ばさず、未コミットの試行が置いた
  entry が撤去されるため）。rollback の target（N-1）は引き続き世代リンクから読まなければ
  ならない。`--json` の
  契約（[apply は last-applied 記録に出所を持たせ、記録と今回の出所が食い違えば警告して続行する](apply-origin-mismatch-warning.md) の出所の食い違い警告を除く）・flock の鍵（profileDir）・prune の対象は変えてはならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:55:00+09:00"
---
# apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む

## 仕様

- **置き場と鍵**: home / project / fixed は `<state>/layat/<鍵>/last-applied.json`、system は
  `<SystemStateDir>/layat/<鍵>/last-applied.json`（既定 `/var/lib`）。鍵は profileDir と同じ規則
  （`--root` なしの home と system は `<name>` 直、他は `<roothash>/<name>` + backref `.root`）。
- **内容**: `recordVersion`（2）/ `linkFarm` / `appliedAt` / `manifest`（そのまま埋め込む）/ `origin`。
  `origin` は直近に記録を書いた apply の出所で、分からなければ省く（規則は [apply は last-applied 記録に出所を持たせ、記録と今回の出所が食い違えば警告して続行する](apply-origin-mismatch-warning.md)）。manifest を
  内容ごと持つので、記録の有効性は store の GC と無関係。
- **書き込み**: 全 FS 操作の成功直後・世代 commit の前に atomic に書く。記録は「この root に
  layat が最後に置いたもの」という FS の事実であり、commit の成否と独立。書けなければ apply は
  失敗し undo journal が FS を巻き戻す。記録ファイルは mode 0644 で書き、非 root の `apply --dryrun` でも読める。
- **更新する場面**: apply（generation skip を含む。失敗規則は同じ）・rollback（FS 収束後・
  ポインタ移動前に N-1 の manifest と link-farm。書けなければ undo journal で巻き戻す）・
  reset（撤去対象のうち target が消えた entry だけを除いた manifest。FS 操作の途中失敗でも
  書き直し、書き直しの失敗は撤去の error と併せて返す。`linkFarm` は保持。記録が無ければ
  profile の current の link-farm。全撤去でファイル削除。書けなければ reset は失敗。undo
  journal は無い。profile が無く記録だけがあるときは profileDir（roothash 系列は backref
  `.root` も）を apply と同じ手順で作ってから flock を取る）。
  `appliedAt` は書いた操作の時刻。
- **読み**: JSON として不正・必須フィールドの欠落や型違い・未知のフィールド・`recordVersion` が 1 でも 2 でもない
  記録は error（「記録なし」に倒さない）。版 1 は `origin` 無しとして読む。
- **比較の基準・基準世代の優先順位**: 記録 → nix profile の current → 空。不一致は記録を採り warning
  1 行。どちらも無ければ初回として「last-applied 記録なし」を warning 1 行。対象は apply の
  古くなった配置の除去と drift 修復、reset の撤去対象、rollback の基準世代（N）。generation skip は新
  link-farm が profile の current に一致し、記録があればその `linkFarm` にも一致するときだけ。
  profile が無ければ（記録だけがあっても）skip しない（記録は commit 前に書かれるので、commit 失敗後は記録と
  profile が食い違う。そのときは通常の処理で記録を比較の基準にして、古くなった配置の除去と commit を行う）。
  rollback の target（N-1）は世代リンクから読む。

変えないもの: `--json` の契約（出所の食い違い警告を除く）・flock の鍵（profileDir）・prune の対象。記録のレイアウトは
[last-applied 記録は `<state>/layat/<鍵>/last-applied.json` に manifest をそのまま埋め込み、比較の基準は記録 → profile → 空の順に読む](../design/last-applied-record-layout.md)、古くなった配置の除去の保守的不変条件は
[古くなった配置の除去は前世代の記録通りを指す symlink のみに限り、copy は消さず孤児を警告する](stale-removal-invariants.md)、rollback の手順は
[rollback は FS を先に収束させてから profile ポインタを最後に移す](rollback-refit-then-pointer.md)、system mode の記録基底は
[system mode は root を / に解決し、変更系操作の前に euid 0 を確認し、profile と last-applied 記録を system 基底に置く](system-mode-root-gate-and-record-base.md) の担当。

## 出典

[last-applied 記録を profileDir と同じ鍵で state 基底に置く](../adr/last-applied-record-location.md)・[last-applied 記録は manifest を内容ごと持つ](../adr/record-embeds-manifest.md)・[記録を書き直す操作ごとに linkFarm の値を定める](../adr/record-rewrite-linkfarm.md)・[古くなった配置の除去は、比較の基準を記録から優先して読む](../adr/baseline-prefers-record.md)・[世代スキップには profile に加え記録の一致も要する](../adr/generation-skip-requires-both-match.md)・[rollback の戻り先は世代リンクから読む](../adr/rollback-target-from-generation-link.md)・[apply は世代の commit の前に記録を書く](../adr/apply-writes-record-before-commit.md)・[rollback はポインタ移動の前に戻り先の manifest で記録を書き直す](../adr/rollback-rewrites-record-before-pointer.md)・[reset は撤去できなかった entry を残して記録を書き直す](../adr/reset-rewrites-record-with-survivors.md)・[記録の読みは不正な記録を error にする](../adr/record-read-strictly.md)。
system の置き場は [system mode の last-applied 記録を /var/lib/layat に置く](../adr/system-record-base-var-lib.md)。`origin` と版 2 は [last-applied 記録に出所を持たせる](../adr/record-carries-origin.md)・[記録の読みは不正な記録を error にする](../adr/record-read-strictly.md)。
