---
id: "REQ-0e341430-17f0-498b-9439-65491652163a"
type: requirement
name: "rollback は FS を先に収束させてから profile ポインタを最後に移す"
status: stable
derives_from:
  - "UC-0b6f60cb-3e98-4ee7-8929-4d94a29f0af6"
specification: |
  `layat rollback <name>` SHALL return the placement to the previous generation. Because
  layat places into an arbitrary root rather than into the profile directory itself, moving
  the profile pointer alone would not change the filesystem, so re-placement SHALL be
  mandatory. The diff for stale removal SHALL take
  the manifest of the departing generation N — the present state of the filesystem — as the
  `baseline`, read from the last-applied record when one exists and otherwise from the
  current generation of the profile ([apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](last-applied-record.md)), and the manifest of the
  generation N-1 being returned to, read from its generation link, as the `target`, and
  the planner SHALL be computed from that pair and reflected onto the filesystem in the
  same order as an apply: pre-placement removal, then placement and replacement, then copy
  reflection, then stale removal. The copy reflection SHALL be place-once — it SHALL copy
  only where the destination is absent and SHALL NOT touch an existing copy destination,
  `rollback` having no equivalent of `--recopy`. Where generation N migrated a
  self-recorded ancestor symlink into nested
  entries below it, the pre-placement removal SHALL remove, before placement, the
  self-recorded symlink that blocks the ancestor of the destination N-1, through the same
  execution path as the pre-placement removal of an apply, stopping with an error on drift
  in the same way. Only at the very end SHALL the profile pointer be moved to N-1 through
  the switch operation of the generation backend contract
  ([engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](generation-backend-contract.md)), the default nixprofile backend running
  `nix-env --profile <profileDir>/profile --switch-generation N-1`; moving the pointer
  first would shift the baseline to N-2 and make stale removal wrong,
  so the filesystem SHALL converge first and the pointer move SHALL come last. After the
  filesystem has converged and before the pointer is moved, the last-applied record SHALL
  be rewritten with the manifest and link-farm of N-1. The apply
  engine SHALL be reused by substituting the pair (`baseline`, `target`).
specification_ja: |
  `layat rollback <name>` は配置を前世代へ戻さなければならない。layat は profile
  dir 自体ではなく任意 root に配置するため、profile ポインタの移動だけでは FS が変わらず、
  再配置が必須である。古くなった配置の除去の diff は、`baseline` を離れる世代 N の manifest（FS の現状。
  last-applied 記録があれば記録から、無ければ profile の current 世代から読む。
  [apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](last-applied-record.md)）、`target` を戻る世代 N-1 の manifest（世代リンクから
  読む）とし、この組で planner を計算して apply と同順
  （配置前除去 → 配置 / 張替え → copy 反映 → 古くなった配置の除去）で FS へ反映しなければならない。
  copy 反映は place-once でなければならず、配置先が不在のときだけコピーしなければならず、
  既存の copy 先に触れてはならない（`rollback` に `--recopy` 相当は無い）。世代 N が自己記録の
  祖先 symlink を配下ネスト entries へ移行した世代であれば、配置前除去は戻り先 N-1 の祖先を
  塞いでいる自己記録 symlink を配置前に除去しなければならない（apply の配置前除去と同一の実行
  処理・drift 時は同じくエラー停止）。最後に世代バックエンドの契約
  （[engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](generation-backend-contract.md)）の切替操作で profile ポインタを N-1 へ移さなければ
  ならない（既定の nixprofile は `nix-env --profile <profileDir>/profile --switch-generation N-1`
  を実行する）。
  ポインタを先に動かすと基準世代が N-2 へずれ古くなった配置の除去が誤るため、FS 収束を先に、
  ポインタ移動を最後にしなければならない。FS の収束後・ポインタ移動の前に、last-applied
  記録を N-1 の manifest と link-farm で書き直さなければならない。apply エンジンは（`baseline`, `target`）の
  差し替えで再利用しなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T13:39:28+09:00"
---
# REQ-0e341430-17f0-498b-9439-65491652163a: rollback は FS を先に収束させてから profile ポインタを最後に移す

## 仕様

**standalone（home / system mode）**: `layat rollback <name>` で前世代に戻す。layat は profile dir 自体では
なく任意 root に配置するため、profile ポインタ移動だけでは FS が変わらず**再配置が必須**。
stale 除去の diff は次の基準・順序で行う:

1. `baseline` = 現世代 N の manifest（FS の現状）／ `target` = 戻る世代 N-1 の manifest
2. `(baseline, target)` で planner を計算し、apply と同順（**PreRemove → place/replace →
   copy 反映 → stale 除去**）で FS へ反映する。copy 反映は place-once（配置先が不在のときだけ
   コピーし、既存の copy 先には触れない。`rollback` に `--recopy` 相当は無い）。N が自己記録の
   祖先 symlink を配下ネスト entries へ移行した
   世代なら、`plan.PreRemove` は戻り先 N-1 の祖先 symlink を塞いでいる自己記録 symlink を配置前
   除去する（apply の PreRemove と同一実行経路・drift 時は同じく error 停止）
3. **最後に** `nix-env --profile <profileDir>/profile --rollback`（または
   `--switch-generation N-1`）で profile ポインタを移す

ポインタを先に動かすと基準世代が N-2 へずれ stale 除去が誤るため、FS 収束を先に・ポインタ移動を
最後にする。apply エンジンを `(baseline, target)` 差し替えで再利用する。

> **上は原文の写しで、規範は frontmatter が正**。原文が併記する次の点は本 item の担当ではない。
>
> - 任意世代切替・世代 GC に `nix-env --profile` 系 / `nix-collect-garbage` を使うこと →
>   [世代操作は世代バックエンドの契約越しに行い、GC root の間引きと store 解放は nix コマンドへ分けて委譲する](generation-gc.md)
> - module で `layat rollback` を公開せず host へ一本化すること → [module 時は rollback を host へ一本化し、layat profile は前進のみで追従する](module-mode-profile-internal.md)
> - project mode で rollback を公開しないこと → [project mode は世代を非公開にし、project / system mode は derivation 同一なら世代を積まず lstat ドリフト修復だけ行う](project-mode-generation-skip.md)
> - `rollback` が名指し必須で `--all` に対応しないこと・`list-generations` が home / system mode
>   限定であること → [rollback は名指し必須で --all に対応しない](rollback-named-only.md) / [rollback と list-generations は home / system mode 限定にする](rollback-list-generations-home-only.md)
> - 途中失敗時に undo ジャーナルで巻き戻すこと → [途中失敗した apply / rollback はインメモリ undo ジャーナルで全 FS 変更を巻き戻す](undo-journal-rollback.md)
> - 配置前除去の判定そのもの → [祖先 symlink は自己記録の古くなったものだけを配置前除去し、それ以外はエラーで停止する](ancestor-symlink-nest-migration.md)
> - copy の place-once そのもの（配置先が不在のときだけコピーする分類）→
>   [copy は target 不在のときだけマテリアライズする place-once で世代管理の対象外とする](copy-place-once.md)
> - `apply --recopy` の契約（既存の copy 先を無条件で上書きする処理）→
>   [apply --recopy は config 内の全 copy target を src から無条件に上書き再コピーする](apply-recopy.md)
>
> **2026-10-11 追記（[世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う](../adr/generation-backend-contract-paths-and-ints.md)）**: 上の写しの手順 3「`nix-env --profile <profileDir>/profile --rollback`
> （または `--switch-generation N-1`）」は、世代バックエンドの契約の Switch で N-1 へ移す、と読み替える
> （`nix-env` は既定実装 nixprofile の内側に閉じる。[engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](generation-backend-contract.md)）。
> ポインタ移動を最後にする順序は変わらない。
>
> **2026-10-11 追記（[古くなった配置の除去は、比較の基準を記録から優先して読む](../adr/baseline-prefers-record.md)）**: 上の写しの手順 1「`baseline` = 現世代 N の manifest」は、離れる世代 N の
> manifest を last-applied 記録があれば記録から、無ければ profile の current 世代から読む、と読み替える。
> `target`（N-1）は引き続き世代リンクから読み、FS の収束後・ポインタ移動の前に記録を N-1 の manifest と
> link-farm で書き直す。記録の規範は [apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](last-applied-record.md) の担当。

## 出典

`docs/spec.md`「世代管理仕様」→「ロールバック」節の standalone（home mode）の箇条書き。

決定の実体は [rollback は FS を収束させてからポインタを動かす](../adr/rollback-baseline-and-pointer-last.md)（rollback diff 基準）で、
戻り先の祖先 symlink を配置前除去する処理は [祖先 symlink は自己記録の古くなったものだけを移行する](../adr/ancestor-symlink-recorded-stale-only.md) が定めている。ポインタ移動を世代
バックエンドの契約越しにしたのは [世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う](../adr/generation-backend-contract-paths-and-ints.md)、比較の基準を
last-applied 記録から優先して読むのは [古くなった配置の除去は、比較の基準を記録から優先して読む](../adr/baseline-prefers-record.md)・[rollback はポインタ移動の前に戻り先の manifest で記録を書き直す](../adr/rollback-rewrites-record-before-pointer.md)。
