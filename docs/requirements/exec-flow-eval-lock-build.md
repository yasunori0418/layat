---
id: "REQ-60c6b7ea-e936-4ce8-bd75-ad35e9c693b9"
type: requirement
name: "実行フローの順序は eval 先行 → flock → build とし build をロック内に閉じる"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The order of the execution flow SHALL be eval first, then flock, then build. Because
  `profileDir` is not determined until after root resolution, and root resolution requires
  the `rootKind` of `manifest.json`, the CLI SHALL obtain `rootKind` in advance with a
  cheap `nix eval` (of the root kind only) *before* building, SHALL determine `profileDir`,
  SHALL then acquire the flock, and SHALL enclose the build inside the lock. This
  simultaneously resolves the circularity of an undetermined `profileDir` and the
  contention over the `.pending` out-link by a build outside the lock.

  The flow SHALL be: (0) entrypoint discovery, overridable by `-f`; (1) a preliminary eval
  of the root kind, followed by root resolution (project = git rev-parse, home = `$HOME`,
  system = `/`, or a fixed path, overridable by `--root`) and determination of
  `profileDir`, `rootKind` being fixed at evaluation time as a passthru of `mkManifest`
  while the actual resolution of the git toplevel / `$HOME` happens at engine runtime;
  then (2) driving the engine: (a) acquire the flock keyed on the determined `profileDir`
  (whether that acquisition blocks or is a try-lock is stated by the `--no-wait`
  specification and is not restated here), concurrent execution against the same
  `profileDir` being the user's responsibility and last-writer-wins on collision;
  (b) inside the lock, run
  `nix build` with `--out-link <profileDir>/.pending` and obtain the link-farm store path
  via `os.Readlink`, the out-link establishing an indirect gcroot that closes the GC
  window between placement and the generation commit, and the build being inside the lock so that
  out-link contention cannot structurally occur; (c) read the baseline manifest — the
  last-applied record when one exists, otherwise the `manifest.json` of the current
  generation of `profileDir` ([apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](last-applied-record.md)) — an absent one
  meaning the first run with zero removal targets; (d) in project mode, skip stacking a
  new generation when the generation-skip condition stated by
  [project mode は世代を非公開にし、project / system mode は derivation 同一なら世代を積まず lstat ドリフト修復だけ行う](project-mode-generation-skip.md) holds, while still lstat-checking each target and re-linking
  only the drifted entries rather than making it a complete no-op, and then writing the
  last-applied record, a failure to write it being unwound by the undo journal; (e) diff the old and
  new `manifest.json`, remove before placement the self-recorded stale entries that block
  placement (ancestor symlink, real directory target, method change), place the new and
  re-linked ones, reflect the copy entries, and then perform conservative stale removal
  with native filesystem operations, any mid-way failure among these four stages unwinding
  every filesystem change made by this run via the in-memory undo journal back to the
  pre-apply state, and once they have succeeded write the last-applied record
  ([apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](last-applied-record.md)), a failure to write it being unwound by the same
  undo journal;
  (f) commit the link-farm as a new generation through the generation backend contract
  ([engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](generation-backend-contract.md)), the commit point, the default nixprofile
  backend running `nix-env --profile <profileDir>/profile --set <link-farm>`; and
  (g) after the commit succeeds, delete `<profileDir>/.pending` as a plain filesystem
  operation that cleans up the temporary build output link.

  Mapped onto this flow, `apply --manifest <link-farm>` SHALL enter at stage 2a and run
  through 2g, skipping stages 0 to 1 and stage 2b; no pending out-link SHALL be
  established there, since nothing is built. What that mode does instead of the skipped
  stages is stated by the `apply --manifest` specification and is not restated here.
specification_ja: |
  実行フローの順序は「eval 先行 → flock → build」でなければならない。profileDir は
  root 解決後にしか確定せず、root 解決には `manifest.json` の `rootKind` が要るため、
  これを安価な `nix eval`（root kind のみ）でビルド前に先取りし、profileDir を確定してから
  flock を取り、build をロック内に閉じる。これで profileDir 未確定の循環と、ロック外
  build の `.pending` out-link 競合が同時に解消する。

  フローは次のとおりでなければならない。(0) entrypoint 発見（`-f` 上書き）。(1) root kind の
  先取り
  eval → root 解決（project = git rev-parse / home = `$HOME` / system = `/` / 固定パス、
  `--root` 上書き）→ profileDir 確定。rootKind は mkManifest の passthru として eval 時に
  確定し、git toplevel / `$HOME` の実体解決は engine 実行時に行う。(2) engine を駆動する:
  (a) 確定 profileDir をキーに flock を取得する（この取得が blocking か try-lock かは
  `--no-wait` の仕様の担当で、本 item では規定しない）。同一 profileDir への同時実行は
  ユーザー責任で衝突時は後勝ちとする。
  (b) ロック内で `nix build` を `--out-link <profileDir>/.pending` 付きで実行し、
  `os.Readlink` で link-farm store path を得る。out-link が indirect gcroot を張り、
  配置〜世代コミットの GC 窓を塞ぐ。build がロック内なので out-link 競合は構造的に起きない。
  (c) 比較の基準の manifest を読む（last-applied 記録があれば記録、無ければ profileDir の
  current 世代の `manifest.json`。[apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](last-applied-record.md)。どちらも無ければ初回 = 削除対象ゼロ）。
  (d) project mode かつ [project mode は世代を非公開にし、project / system mode は derivation 同一なら世代を積まず lstat ドリフト修復だけ行う](project-mode-generation-skip.md) が定める世代スキップの条件を満たすなら
  新世代は積まない。ただし各 target を
  lstat で確認し、ドリフトした entry だけ再張りする（完全 no-op にしない）。その後 last-applied
  記録を書き、書けなければ undo ジャーナルで巻き戻す。(e)
  `manifest.json` を新旧 diff し、配置を塞ぐ自己記録の古くなった配置（祖先 symlink・実 dir target・
  method 変更）を配置前除去し、新規 / 張替を配置し、copy 反映を行い、古くなった配置の保守的な除去
  （ネイティブ FS）を行う。この 4 段のいずれかが途中失敗すると、この run が行った FS 変更を
  全てインメモリ undo ジャーナルで巻き戻し pre-apply 状態へ戻す。4 段が成功したら last-applied
  記録（[apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](last-applied-record.md)）を書き、書けなければ同じ undo ジャーナルで
  巻き戻す。(f) 世代バックエンドの
  契約（[engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](generation-backend-contract.md)）越しに link-farm を新世代としてコミットする
  （コミット点。既定の nixprofile は `nix-env --profile <profileDir>/profile --set <link-farm>`
  を実行する）。(g) コミット成功後に `<profileDir>/.pending` を素の FS 操作で削除する
  （build 出力の一時リンクの片付け）。

  このフロー上で `apply --manifest <link-farm>` は 2a から入って 2g までを通り、0〜1 と
  2b を skip しなければならない。build しないため pending out-link はそこで張ってはならない。
  skip した段の代わりに何を行うかは `apply --manifest` の仕様の担当で、本 item では
  規定しない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T13:39:28+09:00"
---
# REQ-60c6b7ea-e936-4ce8-bd75-ad35e9c693b9: 実行フローの順序は eval 先行 → flock → build とし build をロック内に閉じる

## 仕様

**順序は「eval 先行 → flock → build」**。profileDir は root 解決後にしか確定せず
（project / `--root` 時は `<roothash>`）、root 解決には `manifest.json` の `rootKind` が
要る。これを安価な `nix eval`（root kind のみ）で**ビルド前に**先取りし、profileDir を
確定してから flock を取り、**build をロック内**に閉じる。これで profileDir 未確定の循環と、
ロック外 build の `.pending` out-link 競合（並行 apply の奪い合い）が同時に解消する。
`profileDir` は config 専用ディレクトリで profile リンクは `<profileDir>/profile`・
pending out-link は `<profileDir>/.pending`。

```
layat apply <name> [-f <ep>] [--root <p>]
  0. entrypoint 発見（-f 上書き）
  1. root kind を先取り eval:
     nix eval <ep>#layat.<system>.<name>.rootKind（legacy は per-system 次元なし: nix eval -f <ep> layat.<name>.rootKind）
     → root 解決（kind: project=git rev-parse / home=$HOME / system=/ / 固定パス、--root 上書き）
     → profileDir 確定（home: <name> / project: <roothash>/<name>。--root 明示時は全モード <roothash>/<name>）
       ※ rootKind は mkManifest の passthru として eval 時に確定（git toplevel / $HOME の実体解決は engine 実行時）
  2. engine を駆動:
     a. flock を取得（キー = 確定 profileDir）。
        明示 apply / rollback は blocking（LOCK_EX・取得まで待ち「他の apply 完了待ち」を表示）。
        shellHook 経路（--no-wait）は try-lock（LOCK_NB）で保持中ならスキップし、stderr に1行通知する
        （例: `layat: another apply in progress, skipped (run \`layat apply\` manually)`・シェル入室はブロックしない）。
        同一 profileDir への同時実行はユーザー責任で衝突時は後勝ち
     b. ロック内で nix build <ep>#layat.<system>.<name> --out-link <profileDir>/.pending
        （legacy は nix build -f <ep> layat.<name> --out-link <profileDir>/.pending）
        → os.Readlink で link-farm store path を得る。out-link が indirect gcroot を張り
          配置〜--set の GC 窓を塞ぐ。build がロック内なので out-link 競合は構造的に起きない
     c. profileDir の前世代 manifest.json を読む（無ければ初回 = 削除対象ゼロ）
     d. project mode かつ新 link-farm が前世代と同一なら新世代は積まない（世代スキップ）。ただし各 target を lstat 検査し、
        ドリフトした entry だけ再張りする（完全 no-op にしない）
     e. manifest.json を新旧 diff → 配置を塞ぐ自己記録 stale（祖先 symlink・実 dir target・method 変更）を配置前除去
        （PreRemove・migration）→ 新規/張替を配置 → copy 反映 → 保守的 stale 除去（ネイティブ FS）
        ※ e の 4 段（PreRemove / 配置 / copy 反映 / stale 除去）のいずれかが途中失敗すると、この run が行った
           FS 変更を全てインメモリ undo ジャーナルで巻き戻し pre-apply 状態へ戻す
     f. nix-env --profile <profileDir>/profile --set <link-farm>（サブプロセス・コミット点）
     g. --set 成功後に <profileDir>/.pending を削除（世代リンクが gcroot を引き継ぐ）
```

**`apply --manifest <link-farm>` は 0〜1（entrypoint 発見・rootKind 先取り eval）と 2b
（ロック内 `nix build`）を skip する**。ビルド済み link-farm を engine へ直接渡し、
rootKind は link-farm 内 `manifest.json` から engine が読む。2a（flock 取得）以降〜2g は
通常 apply と同一。pending out-link は build しないため張らない。

> **上のフロー図は原文の写しで、規範は frontmatter が正**。図が触れる次の規範は本 item の
> 担当ではなく、本 item はそれらがフロー上のどの段に位置するかだけを規定する。
>
> - profileDir のレイアウト（`<roothash>/<name>` 等）→ [profileDir は config 専用ディレクトリとし、profile リンク・世代・pending out-link をその中に並べる](profile-on-disk-layout.md)、世代スキップ /
>   ドリフト修復の詳細 → [project mode は世代を非公開にし、project / system mode は derivation 同一なら世代を積まず lstat ドリフト修復だけ行う](project-mode-generation-skip.md)
> - 配置前除去（PreRemove）→ [祖先 symlink は自己記録の古くなったものだけを配置前除去し、それ以外はエラーで停止する](ancestor-symlink-nest-migration.md) / [実 dir の target は全 leaf が除去可能なときだけ全体を配置前除去して symlink 化する](real-dir-target-migration.md) / [method 変更は symlink→copy のみ配置前除去で移行し、copy→symlink は移行しない](method-change-symlink-to-copy-migration.md)、古くなった配置の保守的な
>   除去 → [古くなった配置の除去は前世代の記録通りを指す symlink のみに限り、copy は消さず孤児を警告する](stale-removal-invariants.md)、undo ジャーナル → [途中失敗した apply / rollback はインメモリ undo ジャーナルで全 FS 変更を巻き戻す](undo-journal-rollback.md)
> - `--backup` の退避段（PreRemove の後・配置の前）→ [backup 退避は配置前除去の後・配置の前に置き、drift 修復の処理でも同じく実施する](backup-stage-position.md) /
>   [apply --backup は配置を塞ぐ記録外実体を rename 退避してから配置する](apply-backup-rename-aside.md)。同段はフラグ有効時のみ実行されるため、**本 item は e の段数に含めない**
>   （undo ジャーナルの対象段としては [途中失敗した apply / rollback はインメモリ undo ジャーナルで全 FS 変更を巻き戻す](undo-journal-rollback.md) が退避を含む 5 段で数える）
> - root 解決の各モードの規範 → 「root の解決」節の担当（#209-PR5）
> - `apply --manifest` の契約（引数・併用エラー・entrypoint 発見と eval / build を
>   行わないこと・rootKind を link-farm 内 `manifest.json` から読むこと）→ [apply --manifest はビルド済み link-farm を engine へ直接適用し、--all と合成するときは name=link-farm を繰り返す](apply-manifest-direct.md)。
>   本 item はそれがフロー上のどの段に入り・どの段を skip するかだけを規定する
> - 2a の flock を blocking で取るか try-lock で取るか（`LOCK_EX` / `LOCK_NB`・
>   skip 時の stderr 通知）→ [flock の取得は既定 blocking とし --no-wait のときだけ try-lock でスキップする](flock-acquisition-mode.md)。本 item はフロー上の位置（2a）だけを規定する
>
> **2026-10-11 追記（[世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う](../adr/generation-backend-contract-paths-and-ints.md) / [.pending の後片付けを gcroot を知らない engine の FS 操作として行う](../adr/pending-cleanup-without-gcroot.md)）**: 上の図の f「`nix-env --profile <profileDir>/profile --set`」は、
> 世代バックエンドの契約の Commit と読み替える（`nix-env` は既定実装 nixprofile の内側に閉じる。
> [engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](generation-backend-contract.md)）。g の `.pending` 削除は build 出力の一時リンクの片付けで、
> engine の素の FS 操作。gcroot の引き継ぎは nixprofile の世代リンクの性質で、engine は gcroot を扱わない。
>
> **2026-10-11 追記（[古くなった配置の除去は、比較の基準を記録から優先して読む](../adr/baseline-prefers-record.md)）**: 上の図の c「profileDir の前世代 manifest.json を読む」は、last-applied 記録が
> あれば記録の manifest を比較の基準に採り、無ければ profileDir の current 世代から読む、と読み替える。記録は
> e の後・f の前に書き、世代スキップの処理では d のドリフト修復の後に書く。d の世代スキップの判定元は [project mode は世代を非公開にし、project / system mode は derivation 同一なら世代を積まず lstat ドリフト修復だけ行う](project-mode-generation-skip.md)、記録の規範は [apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](last-applied-record.md) の担当。

## 出典

`docs/spec.md`「CLI 仕様」→「実行フロー」の本文・コードブロックと、
`apply --manifest` の skip を述べる最終箇条書き。

決定の実体は [eval を先に行い、flock を取ってから build する](../adr/eval-first-then-flock-then-build.md)で、out-link に
よる indirect gcroot は [store path は out-link で受け取る](../adr/store-path-via-out-link.md)、同一 profileDir 同時実行の後勝ちは [同一 profileDir への同時実行は後勝ちにする](../adr/same-profiledir-target-last-wins.md)、
`apply --manifest` の skip は [apply --manifest でビルド済み link-farm を engine へ直接渡す](../adr/apply-manifest-prebuilt-link-farm.md)。2a の flock を blocking で取るか try-lock で取るか
（shellHook の skip 通知を含む）は [flock の取得は既定 blocking とし --no-wait のときだけ try-lock でスキップする](flock-acquisition-mode.md) の担当で、その決定元となる ADR は
本 item の出典には含めない。世代のコミットを世代バックエンドの契約越しにし、`.pending` の
削除を engine の FS 操作として位置づけたのは [世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う](../adr/generation-backend-contract-paths-and-ints.md)・[.pending の後片付けを gcroot を知らない engine の FS 操作として行う](../adr/pending-cleanup-without-gcroot.md)、比較の基準を
last-applied 記録から優先して読むのは [古くなった配置の除去は、比較の基準を記録から優先して読む](../adr/baseline-prefers-record.md)。
