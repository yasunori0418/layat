---
id: "REQ-1be4d678-959c-44d7-a346-44bfd95af56e"
type: requirement
name: "世代は link farm derivation を layat 自前 profile へコミットして積み、前世代 manifest から古くなった配置を除去する"
status: stable
derives_from:
  - "UC-0b6f60cb-3e98-4ee7-8929-4d94a29f0af6"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
  - "UC-d39c1994-f9a5-4860-80ba-f6e584adaf14"
specification: |
  The manifest of what has been placed SHALL be embedded inside the store, as part of the
  link farm derivation produced by `lib.mkManifest` (what that derivation contains is
  stated by REQ-60e6b49c-9ba1-4552-a0ec-d340421ec281 and is NOT restated here); the last-applied record
  (REQ-9d622539-a1e5-4719-af51-e0cd85b66f35) SHALL be the only mutable JSON kept outside
  the store, and `manifest.json` SHALL be immutable.
  At run time the engine SHALL, in this order: acquire a flock keyed on the resolved
  `profileDir`; read the baseline manifest (the last-applied record when one exists,
  otherwise the store manifest of the previous generation) and diff it against the
  new generation so as to determine which entries have disappeared; place the symlinks,
  out-of-store symlinks and place-once copies, doing the new placements and replacements
  first and the stale removal last; and only once everything has succeeded, commit the
  link-farm as a new generation through the generation backend contract (the default
  nixprofile backend updates the layat nix profile with `nix-env --profile
  <profileDir>/profile --set <link-farm-drv>`), that being the commit point in every
  mode, so that a partial failure never reaches the commit and the previous generation is
  preserved. When no record exists, the previous generation SHALL be read from the
  previous generation of layat's own profile in every mode, standalone and module alike,
  and SHALL NOT depend on the host's `oldGenPath`.
  layat SHALL hold its own profile in every mode: in standalone it SHALL serve
  user-facing rollback, whereas in a module (home-manager / NixOS / nix-darwin) it SHALL
  be confined to an internal mechanism holding the previous-generation manifest and stale
  tracking. One profile SHALL correspond to one config, each config being atomic.
specification_ja: |
  「配置したもの」のマニフェストは、純粋関数 `lib.mkManifest` が生成する link farm derivation の
  一部として store 内に埋め込まなければならない（当該 derivation が何を含むかは REQ-60e6b49c-9ba1-4552-a0ec-d340421ec281 が
  規定し、ここでは繰り返さない）。store 外に持つ可変 JSON は last-applied 記録
  （REQ-9d622539-a1e5-4719-af51-e0cd85b66f35）だけでなければならず、`manifest.json` は
  不変でなければならない。
  engine は実行時に、解決後 `profileDir` 単位の flock を取得し、比較の基準の manifest
  （last-applied 記録があれば記録、無ければ前世代の store マニフェスト）と
  新世代を diff して消えた entry を判定し、symlink / out-of-store symlink / place-once copy を
  配置し（新規・張替を先に、古くなった配置の除去を最後に）、全て成功してから世代バックエンドの契約越しに
  link-farm を新世代としてコミットしなければならない（既定の nixprofile は
  `nix-env --profile <profileDir>/profile --set <link-farm-drv>` で layat の nix profile を
  更新する）。このコミットが全モード共通のコミット点であり、途中失敗はそこへ到達せず前世代を
  保つ。記録が無いとき、前世代は CLI 単体 / module を問わず全モード共通で layat 自身の
  profile の前世代から読み、ホストの `oldGenPath` に依存してはならない。layat は全モードで自前
  profile を持たなければならず、CLI 単体ではユーザー向け rollback に使い、module
  （HM / NixOS / darwin）では前世代マニフェストと古くなった配置の追跡のための内部機構に留めなければ
  ならない。1 profile は 1 config に対応しなければならず、config 単位で atomic とする。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:54:59+09:00"
---
# REQ-1be4d678-959c-44d7-a346-44bfd95af56e: 世代は link farm derivation を layat 自前 profile へコミットして積み、前世代 manifest から古くなった配置を除去する

## 仕様

- 純粋関数 `lib.mkManifest` が **link farm derivation**（`manifest.json` + 配置元への symlink
  ツリー）を生成する。「配置したもの」のマニフェストは `manifest.json` として link farm の
  一部に **store 内に**埋め込む（store 外の可変 JSON は持たない。`manifest.json` は不変）
- **engine**（Go ライブラリ）が実行時の副作用として:
  0. 解決後 `profileDir` 単位の flock を取得
  1. **前世代の store マニフェスト**（`manifest.json`）と新世代を diff し、消えた entry の
     **symlink を除去**（stale 除去）
     - 前世代は **全モード共通で layat 自身の profile の前世代**から読む（standalone も module も
       同一。ホストの oldGenPath には依存しない）
  2. symlink / out-of-store / place-once copy を配置（新規・張替を先に、stale 除去を最後に）
  3. 全て成功してから `nix-env --profile <profileDir>/profile --set <link-farm-drv>` で layat の
     nix profile を更新（コミット点・全モード）。途中失敗は 3 に到達せず前世代を保つ

**layat は全モードで自前 profile を持つ**。standalone では profile をユーザー向け rollback に
使い、module（HM/NixOS/darwin）では profile を**内部機構**（前世代マニフェスト + stale 追跡）に
留める。

| 機構 | 役割 | 適用層 | 位置 |
|---|---|---|---|
| 世代由来の store マニフェスト | stale 除去のための前回状態（不変・GC-root 済み）| 全層共通 | `manifest.json` として link farm derivation 内に埋め込み |
| layat の nix profile | 前世代の保持・世代番号・GC root | 全モード（standalone はユーザー向け / module は内部）| `profileDir` |

> **上は原文の写しで、規範は frontmatter が正**。原文が参照する次の規範は本 item の
> 担当ではない。
>
> - flock を blocking で取るか try-lock で取るか → [flock の取得は既定 blocking とし --no-wait のときだけ try-lock でスキップする](flock-acquisition-mode.md)
> - ネイティブ FS 操作で配置すること → [engine が叩く外部コマンドは git のみに限る](engine-external-command-constraint.md)、各配置方法の手順 → [symlink 配置は親 dir を作り配置元/subpath を指すリンクを張り、foreign symlink は警告して後勝ちする](symlink-placement-procedure.md) /
>   [copy は target 不在のときだけマテリアライズする place-once で世代管理の対象外とする](copy-place-once.md) / [out-of-store symlink は marker の絶対パスまたは root 相対パスを指し、版管理はリンク先マッピングのみとする](out-of-store-link-mapping-only.md)
> - 古くなった配置の除去の保守的不変条件 → [古くなった配置の除去は前世代の記録通りを指す symlink のみに限り、copy は消さず孤児を警告する](stale-removal-invariants.md)
> - `profileDir` のオンディスクレイアウトと基底 `<state>` → [profileDir は config 専用ディレクトリとし、profile リンク・世代・pending out-link をその中に並べる](profile-on-disk-layout.md)
> - project mode の世代スキップと lstat ドリフト修復 → [project mode は世代を非公開にし、project / system mode は derivation 同一なら世代を積まず lstat ドリフト修復だけ行う](project-mode-generation-skip.md)
> - module 時にユーザー向け rollback を host へ一本化すること → [module 時は rollback を host へ一本化し、layat profile は前進のみで追従する](module-mode-profile-internal.md)
> - `lib.mkManifest` が純粋関数であること・その戻り値 → [mkManifest は配置データを生成する純粋関数である](mkmanifest-pure-function.md) / [mkManifest は manifest.json と symlink farm を含む store オブジェクトを返す](mkmanifest-return-value.md)
>
> 原文が併記する「配置・cleanup アルゴリズムは home-manager の `linkGeneration` /
> `cleanup` を参考に Go で再実装する（`home.file` 自体は再利用しない）」は実装方針の注記で
> 要求ではなく、「`nix` / `git` 以外はサブプロセスを使わない」は [engine が叩く外部コマンドは git のみに限る](engine-external-command-constraint.md) の担当。

> **2026-10-11 追記（[世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う](../adr/generation-backend-contract-paths-and-ints.md) / [古くなった配置の除去は、比較の基準を記録から優先して読む](../adr/baseline-prefers-record.md)）**: 上の写しの手順 3「`nix-env --profile <profileDir>/profile
> --set` で layat の nix profile を更新」は、世代バックエンドの契約の Commit と読み替える（`nix-env` は
> 既定実装 nixprofile の内側に閉じる。[engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](generation-backend-contract.md)）。「store 外の可変 JSON は
> 持たない」は last-applied 記録（[apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](last-applied-record.md)）に限って覆し、手順 1 の前世代
> manifest は記録があれば記録の manifest を比較の基準に採る（profile の前世代から読むのは記録が無いとき）。
> 世代由来の store マニフェストと機構表は不変。

逆に、上の写しには現れないが規範文が持つものが 1 つある。**1 profile = 1 config の atomic
性**で、原文はこれを本節ではなく「CLI 仕様」→「サブコマンド体系」の `apply` の箇条書き
（「profile は config 単位で atomic」）で述べる。atomic 性は「1 config が 1 profile を持ち、
その世代がコミット点で丸ごと切り替わる」という profile の機構そのものの性質であり、機構を
規定する本 item が引き受けるのが所在として自然なため、ここで規範化した（当該箇所を分割した
[apply の config 選択は name 省略で default・明示で単一・--all で全件](apply-config-selection.md) と、これを論拠として引く [legacy entrypoint は mkShell passthru 形を canonical とし CLI の attr path を分岐させない](legacy-entrypoint-passthru-canonical.md) も規範の所在を本 item に置いている）。

## 出典

`docs/spec.md`「世代管理仕様」→「機構」節の箇条書きと機構表、および同節の
「layat は全モードで自前 profile を持つ」段落。atomic 性のみ同「CLI 仕様」→
「サブコマンド体系」の `apply` の箇条書き（上の注記を参照）。

決定の実体は [世代は link farm と nix profile に乗せる](../adr/generations-on-link-farm-and-nix-profile.md)「世代は link farm と nix profile に乗せる」と
[symlink farm を GC アンカー専用にする](../adr/farm-as-gc-anchor-only.md)「symlink farm を GC アンカー専用にする」（link farm derivation への manifest 埋め込み）・
[新規配置と張替えを先に、世代のコミットを最後に行う](../adr/placement-order-commit-last.md)「新規配置と張替えを先に、世代のコミットを最後に行う」（コミット最後）。全モードのコミットを
`nix-env --profile <profileDir>/profile --set` に統一することは旧決定が定めていた。
コミットを世代バックエンドの契約越しにしたのは [世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う](../adr/generation-backend-contract-paths-and-ints.md)「世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う」、
last-applied 記録を store 外の唯一の可変 JSON として認めたのは [last-applied 記録を profileDir と同じ鍵で state 基底に置く](../adr/last-applied-record-location.md)「last-applied 記録を profileDir と同じ鍵で state 基底に置く」。
