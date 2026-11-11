# layat 仕様書

layat が「何を満たすべきか」の全体像と、個別仕様（`requirement` item）・品質方針
（`quality` item）・テスト計画（`test_plan` item）への索引。

規範的な仕様は **すべて `docs/requirements/` の item が持ち**、開発プロセスと規約に関する
方針は `docs/quality/` の item が、何をどこまでテストするかは `docs/test-plan/` の item が
持つ。本文書は通読の起点として全体像だけを述べ、詳細は item へのリンクで示す
（README → 本文書 → item の 3 層構造）。

> **この文書の書き方（規約）**
>
> - 散文は **見出し 1 つ（h2 / h3 の最も内側）につき 10 行以内**。表・リンク列挙・コード
>   ブロックはこの制限に含めない
> - **詳細は必ず item へリンクし、本文に書き下さない**。仕様の断片をここへ写すと item と
>   二重管理になり、改訂時に片方だけ古くなる
> - `requirement` / `quality` / `test_plan` item を足したら本文書のリンク集にも足す。逆に本文書へ
>   仕様を書き足さない
> - item を横断して検索・追跡するには `sara query`（→ `CLAUDE.md` の「item を辿る（`sara query`）」節）を使う
>
> **item の `## 出典` は概要文書の現行章立てを指さない。** 出典が記録するのは分割時点
> （epic #203）の原文の記述箇所であり、節見出しに限らず表の行・blockquote・箇条書きの項も
> 指す。縮退で章立てを再編したため、出典の節名の多くは現行の見出しと一致しない。item から
> 現行仕様へ辿るのは出典ではなく relation（`sara query`）の役割。
>
> 原文を読むときは、**出典が名指しする文書**（`docs/spec.md` / `docs/design.md` /
> `docs/concept.md` のいずれか。以下 `<doc>`）の縮退コミットの親を見る。3 文書とも同じ形の
> subject（`docs(<scope>): <doc> を概要 + item リンク集へ縮退する`）を持つ。
>
> ```bash
> git log --oneline -- <doc>        # 「概要 + item リンク集へ縮退する」の <ref> を探す
> git show <ref>^:<doc>             # その親が縮退前の原文
> ```

---

## アーキテクチャ概要

layat は **CLI とエンジンの 2 層**で構成する（→ [engine を単一の固定 Go バイナリにする](adr/engine-fixed-go-binary.md), [engine は cmd/ と internal/ に置きライブラリ化しない](adr/engine-internal-cmd-layout.md)）。CLI は manifest 生成器
（→ [manifest の取得をプロセス内の Go インターフェイスの生成器越しにする](adr/manifest-generator-in-process-interface.md)、既定 `nix`）を通じて entrypoint を発見し、生成器の Build をエンジンへ渡す。エンジンは
ロック内で Build を呼び、生成器が build した link-farm のパス（既定の nixprofile では store path）を
受け取る。エンジンは配置内容（entry）を `manifest.json` から読み、ネイティブ FS 操作で配置と古くなった配置の除去を行う。
世代のコミットと切替は世代バックエンド（→ [世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う](adr/generation-backend-contract-paths-and-ints.md)、既定 nixprofile）に任せる。
配置内容をエンジンへ渡すのは `manifest.json` だけ。

- [layat は CLI とエンジンの 2 層で構成する](requirements/two-layer-architecture.md) — layat は CLI とエンジンの 2 層で構成する
- [config は Nix で書き nix build で評価する](requirements/config-written-in-nix.md) — config は Nix で書き nix build で評価する
- [engine が叩く外部コマンドは git のみに限る](requirements/engine-external-command-constraint.md) — engine が叩く外部コマンドは git のみに限る

---

## lib API

`lib`（`mkManifest` / `mkApplyCommand` / マーカー群）は nixpkgs.lib のみに依存する純データ生成器で、配置ロジックは
持たない（→ [lib は純データ生成に徹する](adr/lib-produces-data-only.md)）。入力の確認は `mkManifest` 自身が `evalModules` + `normalizeManifest` で
行い、1 か所に集約する。

- [lib は nixpkgs.lib のみに依存する純データ生成器である](requirements/lib-pure-data-generator.md) — lib は nixpkgs.lib のみに依存する純データ生成器である

### `lib.mkManifest`

- [mkManifest は配置データを生成する純粋関数である](requirements/mkmanifest-pure-function.md) — mkManifest は配置データを生成する純粋関数である
- [mkManifest の引数は必須の pkgs / entries / root と任意の identity とする](requirements/mkmanifest-arguments.md) — mkManifest の引数は必須の pkgs / entries / root と任意の identity とする
- [mkManifest は manifest.json と symlink farm を含む store オブジェクトを返す](requirements/mkmanifest-return-value.md) — mkManifest は manifest.json と symlink farm を含む store オブジェクトを返す
- [mkManifest の返り値は passthru で `rootKind` を露出する](requirements/mkmanifest-passthru-rootkind.md) — mkManifest の返り値は passthru で root kind を露出する
- [entry の識別子は属性キー = target とし一意性は Nix が担保する](requirements/entry-identity-by-target-key.md) — entry の識別子は属性キー = target とし一意性は Nix が担保する
- [root は明示必須で暗黙デフォルトを持たない](requirements/root-explicit-required.md) — root は明示必須で暗黙デフォルトを持たない
- [root は 3 マーカーと絶対パス文字列の union を取る](requirements/root-marker-kinds.md) — root は 3 マーカーと絶対パス文字列の union を取る
- [root マーカーの subdir 適用形は全 target への前置と等価な評価時の糖衣で、root を動かさない](requirements/root-marker-subdir-sugar.md) — root マーカーの subdir 適用形は全 target への前置と等価な評価時の糖衣で、root を動かさない

### マーカー構築子

- [mkOutOfStoreSymlink は out-of-store symlink を表すマーカーを返す](requirements/mkoutofstoresymlink-marker.md) — mkOutOfStoreSymlink は out-of-store symlink を表すマーカーを返す
- [mkRootRelativeSymlink は配置 root 相対の src を表すマーカーを返し、engine が target 起点の相対 symlink として張る](requirements/root-relative-symlink-marker.md) — mkRootRelativeSymlink は配置 root 相対の src を表すマーカーを返し、engine が target 起点の相対 symlink として張る
- [root マーカーは `kind` を運ぶ入れ物でパス解決は engine が行う](requirements/root-markers-runtime-resolution.md) — root マーカーは kind を運ぶ入れ物でパス解決は engine が行う

### `lib.mkApplyCommand`

- [lib.mkApplyCommand は事前展開した apply --all --manifest のコマンド行を返す](requirements/lib-apply-command.md) — lib.mkApplyCommand は事前展開した apply --all --manifest のコマンド行を返す

### 入力の確認（`evalModules` + `normalizeManifest`）

- [mkManifest 自身が evalModules で入力を確認する単一の箇所になる](requirements/mkmanifest-input-validation.md) — mkManifest 自身が evalModules で入力を確認する単一の箇所になる
- [normalizeManifest が確認・デフォルト適用・marker 変換を行い mkManifest が derivation を組む](requirements/normalize-manifest-defaults.md) — normalizeManifest が確認・デフォルト適用・marker 変換を行い mkManifest が derivation を組む
- [target / subpath のパス安全性を評価時に確認する](requirements/path-safety-validation.md) — target / subpath のパス安全性を評価時に確認する
- [target 衝突の検出処理を同一 manifest 内と cross-config で分ける](requirements/target-collision-eval-detection.md) — target 衝突の検出処理を同一 manifest 内と cross-config で分ける
- [意図が矛盾する組み合わせをクロスフィールドの確認で評価時に拒否する](requirements/normalize-manifest-cross-field-checks.md) — 意図が矛盾する組み合わせをクロスフィールドの確認で評価時に拒否する
- [marker は判別タグで識別し manifest.json には漏らさない](requirements/marker-tag-discrimination.md) — marker は判別タグで識別し manifest.json には漏らさない

---

## CLI 仕様（一次 UX）

`layat` CLI は PATH に常駐する一次 UX（→ [CLI が一次 UX で devShell に pin 版を同梱する](adr/cli-primary-ux-devshell-pinned.md)）。entrypoint を発見し、rootKind を先取り
eval して root を解決し、flock を取ってから `nix build` をロック内で回してエンジンを駆動する。

- [layat CLI は PATH 常駐の一次 UX で、project mode は devShell 同梱を canonical とする](requirements/cli-primary-ux-installation.md) — layat CLI は PATH 常駐の一次 UX で、project mode は devShell 同梱を canonical とする
- [nix experimental-features は nix 生成器の前提条件とし、自動付与せず案内エラーで停止する](requirements/nix-experimental-features-prerequisite.md) — nix experimental-features は nix 生成器の前提条件とし、自動付与せず案内エラーで停止する
- [CLI は manifest 生成器の契約（Discover / Roots / Build / DryBuild）越しに manifest を得る](requirements/manifest-generator-contract.md) — CLI は manifest 生成器の契約（Discover / Roots / Build / DryBuild）越しに manifest を得る
- [生成器は明示指定のみで選び、フラグ・環境変数・設定ファイル・既定の順に解決する](requirements/generator-selection-precedence.md) — 生成器は明示指定のみで選び、フラグ・環境変数・設定ファイル・既定の順に解決する
- [生成器は診断を CLI が渡す writer へ素通しし、失敗を構造化して返し、表示は CLI が決める](requirements/generator-diagnostics-contract.md) — 生成器は診断を CLI が渡す writer へ素通しし、失敗を構造化して返し、表示は CLI が決める

### entrypoint の発見・アドレッシング

- [entrypoint は CWD で flake.nix → shell.nix → default.nix の順に探し -f で上書きする](requirements/entrypoint-discovery-order.md) — entrypoint は CWD で flake.nix → shell.nix → default.nix の順に探し -f で上書きする
- [entrypoint は `layat.<name>` に named manifest を公開し CLI は形ごとの attr path で build する](requirements/named-manifest-addressing.md) — entrypoint は layat.\<name\> に named manifest を公開し CLI は形ごとの attr path で build する
- [config 名 default を慣例の解決先とし専用 layat 名前空間で packages を汚さない](requirements/default-config-name-and-namespace.md) — config 名 default を慣例の解決先とし専用 layat 名前空間で packages を汚さない
- [flake-parts での使い方は直書きと同一の derivation を生み CLI のアドレッシングを変えない](requirements/flake-parts-module-path.md) — flake-parts での使い方は直書きと同一の derivation を生み CLI のアドレッシングを変えない
- [legacy entrypoint は mkShell passthru 形を canonical とし CLI の attr path を分岐させない](requirements/legacy-entrypoint-passthru-canonical.md) — legacy entrypoint は mkShell passthru 形を canonical とし CLI の attr path を分岐させない
- [legacy entrypoint では相対 path の src が自動で store 化されない](requirements/legacy-src-not-auto-stored.md) — legacy entrypoint では相対 path の src が自動で store 化されない

### サブコマンド体系

- [apply の config 選択は name 省略で default・明示で単一・--all で全件](requirements/apply-config-selection.md) — apply の config 選択は name 省略で default・明示で単一・--all で全件
- [apply --all は選択された config を並列に適用し、集約表示と results\[\] を辞書順にし、部分失敗でも続行して最後に集約する](requirements/apply-all-lexical-order-continue.md) — apply --all は選択された config を並列に適用し、集約表示と results[] を辞書順にし、部分失敗でも続行して最後に集約する
- [--all は root モードフィルタで対象 config を絞れる](requirements/apply-all-root-mode-filter.md) — --all は root モードフィルタで対象 config を絞れる
- [apply --all は build 前に選択 config 間の target 衝突を検出して停止する](requirements/apply-all-cross-config-preflight.md) — apply --all は build 前に選択 config 間の target 衝突を検出して停止する
- [apply --all は build を --jobs の worker pool で先行並列実行し build 失敗 config を部分失敗として集約する](requirements/apply-all-parallel-build-jobs.md) — apply --all は build を --jobs の worker pool で先行並列実行し build 失敗 config を部分失敗として集約する
- [apply 修飾フラグは --all と合成できる](requirements/apply-modifier-flag-composition.md) — apply 修飾フラグは --all と合成できる
- [apply --dryrun は読み取り専用で conflict 検出時に非ゼロ終了する](requirements/apply-dryrun-readonly.md) — apply --dryrun は読み取り専用で conflict 検出時に非ゼロ終了する
- [apply --recopy は config 内の全 copy target を src から無条件に上書き再コピーする](requirements/apply-recopy.md) — apply --recopy は config 内の全 copy target を src から無条件に上書き再コピーする
- [apply --backup は配置を塞ぐ記録外実体を rename 退避してから配置する](requirements/apply-backup-rename-aside.md) — apply --backup は配置を塞ぐ記録外実体を rename 退避してから配置する
- [apply --manifest はビルド済み link-farm を engine へ直接適用し、--all と合成するときは name=link-farm を繰り返す](requirements/apply-manifest-direct.md) — apply --manifest はビルド済み link-farm を engine へ直接適用し、--all と合成するときは name=link-farm を繰り返す
- [reset は profile を触らない FS-only teardown で配置物を無い状態へ戻す](requirements/reset-fs-only-teardown.md) — reset は profile を触らない FS-only teardown で配置物を無い状態へ戻す
- [reset は名指し必須で profileDir 単位の blocking flock を取る](requirements/reset-named-only-and-flock.md) — reset は名指し必須で profileDir 単位の blocking flock を取る
- [reset --dryrun は副作用ゼロで削除対象を表示して終了する](requirements/reset-dryrun-no-side-effect.md) — reset --dryrun は副作用ゼロで削除対象を表示して終了する
- [prune は backref の root が実在しない roothash 系列だけを系列ごと削除する](requirements/prune-target-selection.md) — prune は backref の root が実在しない roothash 系列だけを系列ごと削除する
- [prune は dryrun・root 一覧付き確認・try-lock skip・--json の --yes 必須で削除を守る](requirements/prune-safety-gates.md) — prune は dryrun・root 一覧付き確認・try-lock skip・--json の --yes 必須で削除を守る
- [rollback と list-generations は home / system mode 限定にする](requirements/rollback-list-generations-home-only.md) — rollback と list-generations は home / system mode 限定にする
- [rollback は config の名指しを必須にする](requirements/rollback-named-only.md) — rollback は config の名指しを必須にする
- [status は記録基底を走査して last-applied 記録を読み取り専用で表示する](requirements/status-command.md) — status は記録基底を走査して last-applied 記録を読み取り専用で表示する
- [gitignore は配置 target を stdout へ列挙するだけでファイルを書き込まない](requirements/gitignore-anchored-output.md) — gitignore は配置 target を stdout へ列挙するだけでファイルを書き込まない
- [gitignore は project mode 限定で非 project config を指定したらエラーで停止する](requirements/gitignore-project-mode-only.md) — gitignore は project mode 限定で非 project config を指定したらエラーで停止する
- [gitignore は method を区別せず copy target も含めて全 target を列挙する](requirements/gitignore-includes-copy-targets.md) — gitignore は method を区別せず copy target も含めて全 target を列挙する
- [gitignore --all は projectRoot の全 config の target をソート + 重複除去して出力する](requirements/gitignore-all-dedup.md) — gitignore --all は projectRoot の全 config の target をソート + 重複除去して出力する
- [layat init は nix flake init -t への透明なラッパーとしファイルを生成しない](requirements/init-template-wrapper.md) — layat init は nix flake init -t への透明なラッパーとしファイルを生成しない
- [init のテンプレート参照はバイナリにハードコードした固定 flake ref とする](requirements/init-fixed-flake-ref.md) — init のテンプレート参照はバイナリにハードコードした固定 flake ref とする
- [template は動く example を 1 config だけ置きバリエーションはコメントで示す](requirements/init-template-contents.md) — template は動く example を 1 config だけ置きバリエーションはコメントで示す
- [--root は全モード共通で解決 root を明示上書きする](requirements/root-override-flag.md) — --root は全モード共通で解決 root を明示上書きする
- [flock の取得は既定 blocking とし --no-wait のときだけ try-lock でスキップする](requirements/flock-acquisition-mode.md) — flock の取得は既定 blocking とし --no-wait のときだけ try-lock でスキップする
- [--version は埋め込みバージョンを cobra 既定書式で表示して終了し短縮形を持たない](requirements/version-flag.md) — --version は埋め込みバージョンを cobra 既定書式で表示して終了し短縮形を持たない
- [一部の entry だけを適用する --only を提供する](requirements/no-only-flag.md) — 棄却した提案。`--only` は提供しない
- [内部実行する nix コマンドを開示し世代の切替と GC は標準の nix コマンドへ委譲する](requirements/nix-command-transparency-and-delegation.md) — 内部実行する nix コマンドを開示し世代の切替と GC は標準の nix コマンドへ委譲する

### 出力ストリームと終了コード

- [stdout は機械可読出力を専有しレポートと warning は stderr へ出す](requirements/stream-discipline.md) — stdout は機械可読出力を専有しレポートと warning は stderr へ出す
- [成功時はデフォルト沈黙とし warning と error は常時 stderr へ出す](requirements/silent-on-success.md) — 成功時はデフォルト沈黙とし warning と error は常時 stderr へ出す
- [冗長度は -v、デバッグは --debug に分離し --json と直交させる](requirements/verbose-and-debug-separation.md) — 冗長度は -v、デバッグは --debug に分離し --json と直交させる
- [終了コードは 0 = 成功 / 1 = 一般エラー / 2 = dryrun の conflict とする](requirements/exit-codes.md) — 終了コードは 0 = 成功 / 1 = 一般エラー / 2 = dryrun の conflict とする
- [apply --all --dryrun の終了コードは error を conflict より優先する](requirements/apply-all-dryrun-exit-code-priority.md) — apply --all --dryrun の終了コードは error を conflict より優先する

#### outturn 準拠の `--json`（第 2 契約・→ [--json は outturn 規約に常時準拠する](adr/outturn-conformance-principle.md)）

- [--json は outturn 規約準拠のエンベロープを出す第 2 契約とする](requirements/json-outturn-envelope.md) — --json は outturn 規約準拠のエンベロープを出す第 2 契約とする
- [--json 指定時は行指向 stdout を出さずエンベロープが stdout を専有する](requirements/json-stdout-exclusive.md) — --json 指定時は行指向 stdout を出さずエンベロープが stdout を専有する
- [エンベロープはコマンド完了時に 1 回だけ出し成立条件を満たさない実行では出さない](requirements/json-emit-timing.md) — エンベロープはコマンド完了時に 1 回だけ出し成立条件を満たさない実行では出さない
- [変更系の JSON ペイロードは engine 結果からフルインベントリと実差分を導く](requirements/json-change-payload.md) — 変更系の JSON ペイロードは engine 結果からフルインベントリと実差分を導く
- [読み取り系の JSON ペイロードは dryRun パリティと info インベントリで表す](requirements/json-readonly-payload.md) — 読み取り系の JSON ペイロードは dryRun パリティと info インベントリで表す
- [prune --json はトップレベル info の removed に削除系列を、skipped に判定か削除に至れず残した系列を載せる](requirements/prune-json-payload.md) — prune --json はトップレベル info の removed に削除系列を、skipped に判定か削除に至れず残した系列を載せる
- [--all は config ごとの SubjectResult を単一実行と同一形状で積む](requirements/json-all-subject-results.md) — --all は config ごとの SubjectResult を単一実行と同一形状で積む
- [エンベロープのエラーは主体の有無で層を分けコードを分類する](requirements/json-error-layering.md) — エンベロープのエラーは主体の有無で層を分けコードを分類する
- [item id は identity の JCS を SHA-256 した小文字 hex とする](requirements/json-item-id-derivation.md) — item id は identity の JCS を SHA-256 した小文字 hex とする
- [reset --json は --yes を必須とし無ければ fail fast する](requirements/json-reset-requires-yes.md) — reset --json は --yes を必須とし無ければ fail fast する

### 実行フロー

- [実行フローの順序は eval 先行 → flock → build とし build をロック内に閉じる](requirements/exec-flow-eval-lock-build.md) — 実行フローの順序は eval 先行 → flock → build とし build をロック内に閉じる
- [非 build コマンドも eval 先行を共通前段に持つ](requirements/non-build-commands-eval-first.md) — 非 build コマンドも eval 先行を共通前段に持つ
- [apply --all は rootKind を 1 回の一括 eval で取る](requirements/apply-all-batched-eval.md) — apply --all は rootKind を 1 回の一括 eval で取る
- [--dryrun は root を解決するが flock も pending gcroot も取らない](requirements/dryrun-no-lock-no-gcroot.md) — --dryrun は root を解決するが flock も pending gcroot も取らない
- [失敗時に残る pending gcroot は config あたり最大 1 個に有界とし回収処理を持たない](requirements/pending-gcroot-bounded.md) — 失敗時に残る pending gcroot は config あたり最大 1 個に有界とし回収処理を持たない

### 再現性スタンス

- [flake は pure eval で flake.lock が固定し legacy は impure を許容しユーザー責任とする](requirements/reproducibility-stance.md) — flake は pure eval で flake.lock が固定し legacy は impure を許容しユーザー責任とする
- [layat カスタム output は nix flake check の unknown 警告を許容し主検証は nix build で行う](requirements/flake-check-unknown-output.md) — layat カスタム output は nix flake check の unknown 警告を許容し主検証は nix build で行う

---

## entries スキーマ仕様

`entries` は **属性キー = target の attrset**で、キーが識別子になる（→ [entries を target をキーにした attrset にする](adr/entries-keyed-by-target.md)）。

- [entry submodule のフィールドは src / subpath / target / method の 4 つとする](requirements/entry-submodule-fields.md) — entry submodule のフィールドは src / subpath / target / method の 4 つとする
- [entry submodule は strict とし未知キーと旧名を評価時エラーにする](requirements/entry-submodule-strict.md) — entry submodule は strict とし未知キーと旧名を評価時エラーにする
- [src は path / set / marker の 3 種を取り store link を既定として out-of-store は marker で opt-in する](requirements/src-kinds-store-default.md) — src は path / set / marker の 3 種を取り store link を既定として out-of-store は marker で opt-in する
- [subpath は src 内の相対パスとし、リポジトリ全体は省略で表して専用トークンを設けない](requirements/subpath-omission-whole-repo.md) — subpath は src 内の相対パスとし、リポジトリ全体は省略で表して専用トークンを設けない
- [method は配置方法を選び symlink は世代管理下・copy は世代管理外になる](requirements/method-src-matrix.md) — method は配置方法を選び symlink は世代管理下・copy は世代管理外になる

---

## manifest.json スキーマ（v1・Nix↔Go 契約）

`manifest.json` は Nix とエンジンの唯一の安定契約（→ [manifest.json へ marker を `clean` enum で写す](adr/wire-format-clean-enum.md), [manifest.json を schemaVersion 1 の唯一の契約として固定する](adr/manifest-json-contract-schemaversion-v1.md)）。

- [manifest.json が Nix と engine の唯一の安定契約であり schemaVersion は 1 に固定する](requirements/manifest-single-stable-contract.md) — manifest.json が Nix と engine の唯一の安定契約であり schemaVersion は 1 に固定する
- [MVP は manifest v1 のみを発行・受理しマイグレーション機構を持たない](requirements/manifest-v1-only-no-migration.md) — MVP は manifest v1 のみを発行・受理しマイグレーション機構を持たない
- [manifest.json のトップレベルは必須の schemaVersion / root / entries と任意の identity とする](requirements/manifest-toplevel-fields.md) — manifest.json のトップレベルは必須の schemaVersion / root / entries と任意の identity とする
- [manifest.json の root は rootKind を持ち fixed のときだけ絶対パスを併記する](requirements/manifest-root-object.md) — manifest.json の root は rootKind を持ち fixed のときだけ絶対パスを併記する
- [manifest.json の entries は attrset を配列へ正規化し 5 フィールドを記録する](requirements/manifest-entries-array.md) — manifest.json の entries は attrset を配列へ正規化し 5 フィールドを記録する
- [symlink farm は GC アンカー専用でアンカーは store-backed な symlink entry に限る](requirements/farm-gc-anchor-only.md) — symlink farm は GC アンカー専用でアンカーは store-backed な symlink entry に限る
- [symlink farm の GC アンカー名は target のハッシュとする](requirements/farm-anchor-name-hash.md) — symlink farm の GC アンカー名は target のハッシュとする

---

## 配置動作仕様

symlink は配置前除去 → 配置 → 古くなった配置の除去の順で進み、途中失敗はインメモリ undo ジャーナルで
巻き戻す（→ [途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](adr/in-memory-undo-journal.md), [配置前除去を planner が計算するデータとして表す](adr/preremove-as-planner-data.md)）。copy は place-once でユーザー管理に委ねる。

### symlink モード

- [symlink 配置は親 dir を作り配置元/subpath を指すリンクを張り、foreign symlink は警告して後勝ちする](requirements/symlink-placement-procedure.md) — symlink 配置は親 dir を作り配置元/subpath を指すリンクを張り、foreign symlink は警告して後勝ちする
- [既存 symlink の張替えは unlink + symlink の 2 操作で行い冪等な再実行で収束させる](requirements/symlink-replace-unlink-symlink.md) — 既存 symlink の張替えは unlink + symlink の 2 操作で行い冪等な再実行で収束させる
- [target に通常ファイル・ディレクトリが在れば上書きせずエラーで停止する](requirements/conflict-on-real-file-or-dir.md) — target に通常ファイル・ディレクトリが在れば上書きせずエラーで停止する
- [祖先 symlink は自己記録の古くなったものだけを配置前除去し、それ以外はエラーで停止する](requirements/ancestor-symlink-nest-migration.md) — 祖先 symlink は自己記録の古くなったものだけを配置前除去し、それ以外はエラーで停止する
- [実 dir の target は全 leaf が除去可能なときだけ全体を配置前除去して symlink 化する](requirements/real-dir-target-migration.md) — 実 dir の target は全 leaf が除去可能なときだけ全体を配置前除去して symlink 化する
- [method 変更は symlink→copy のみ配置前除去で移行し、copy→symlink は移行しない](requirements/method-change-symlink-to-copy-migration.md) — method 変更は symlink→copy のみ配置前除去で移行し、copy→symlink は移行しない
- [backup 退避は配置前除去の後・配置の前に置き、drift 修復の処理でも同じく実施する](requirements/backup-stage-position.md) — backup 退避は配置前除去の後・配置の前に置き、drift 修復の処理でも同じく実施する

### 途中失敗時の巻き戻し

- [途中失敗した apply / rollback はインメモリ undo ジャーナルで全 FS 変更を巻き戻す](requirements/undo-journal-rollback.md) — 途中失敗した apply / rollback はインメモリ undo ジャーナルで全 FS 変更を巻き戻す
- [巻き戻し自体の失敗は best-effort で続行し、全件を stderr へ報告して停止する](requirements/undo-rollback-best-effort.md) — 巻き戻し自体の失敗は best-effort で続行し、全件を stderr へ報告して停止する

### copy モード・out-of-store・recopy / reset

- [copy は target 不在のときだけマテリアライズする place-once で世代管理の対象外とする](requirements/copy-place-once.md) — copy は target 不在のときだけマテリアライズする place-once で世代管理の対象外とする
- [copy が foreign 実ファイルを skip したときは warning で可視化する](requirements/copy-foreign-skip-warning.md) — copy が foreign 実ファイルを skip したときは warning で可視化する
- [copy は元の mode を保存しつつ owner-write を付与する](requirements/copy-mode-owner-write.md) — copy は元の mode を保存しつつ owner-write を付与する
- [copy は src ツリー内の symlink を deref せず symlink のまま複製する](requirements/copy-preserves-inner-symlinks.md) — copy は src ツリー内の symlink を deref せず symlink のまま複製する
- [out-of-store symlink は marker の絶対パスまたは root 相対パスを指し、版管理はリンク先マッピングのみとする](requirements/out-of-store-link-mapping-only.md) — out-of-store symlink は marker の絶対パスまたは root 相対パスを指し、版管理はリンク先マッピングのみとする
- [recopy の上書きは削除ではなく同一親内への rename 退避で行う](requirements/recopy-rename-aside.md) — recopy の上書きは削除ではなく同一親内への rename 退避で行う
- [reset の確認プロンプトは stdin が TTY のときだけ出し、非 TTY で同意が無ければ即エラー停止する](requirements/reset-confirm-non-tty.md) — reset の確認プロンプトは stdin が TTY のときだけ出し、非 TTY で同意が無ければ即エラー停止する

---

## 世代管理仕様

世代は link farm derivation を layat 自前の profile へコミットして積む（→ [世代は link farm と nix profile に乗せる](adr/generations-on-link-farm-and-nix-profile.md), [profile は専用ディレクトリレイアウトに置く](adr/profile-directory-layout.md)）。
古くなった配置の除去は前世代 manifest の記録通りを指す symlink だけに限る保守的な操作。

- [世代は link farm derivation を layat 自前 profile へコミットして積み、前世代 manifest から古くなった配置を除去する](requirements/generation-mechanism-link-farm-manifest.md) — 世代は link farm derivation を layat 自前 profile へコミットして積み、前世代 manifest から古くなった配置を除去する
- [profileDir は config 専用ディレクトリとし、profile リンク・世代・pending out-link をその中に並べる](requirements/profile-on-disk-layout.md) — profileDir は config 専用ディレクトリとし、profile リンク・世代・pending out-link をその中に並べる
- [古くなった配置の除去は前世代の記録通りを指す symlink のみに限り、copy は消さず孤児を警告する](requirements/stale-removal-invariants.md) — 古くなった配置の除去は前世代の記録通りを指す symlink のみに限り、copy は消さず孤児を警告する
- [target 除去後は空の親ディレクトリチェーンを root 境界まで保守的に剪定する](requirements/empty-parent-dir-pruning.md) — target 除去後は空の親ディレクトリチェーンを root 境界まで保守的に剪定する
- [世代操作は世代バックエンドの契約越しに行い、GC root の間引きと store 解放は nix コマンドへ分けて委譲する](requirements/generation-gc.md) — 世代操作は世代バックエンドの契約越しに行い、GC root の間引きと store 解放は nix コマンドへ分けて委譲する
- [engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う](requirements/generation-backend-contract.md) — engine は世代操作を世代バックエンドの契約（Commit / List / Switch / UserBase / SystemBase）越しに行う
- [apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](requirements/last-applied-record.md) — apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む
- [apply は last-applied 記録に出所を持たせ、記録と今回の出所が食い違えば警告して続行する](requirements/apply-origin-mismatch-warning.md) — apply は last-applied 記録に出所を持たせ、記録と今回の出所が食い違えば警告して続行する
- [rollback は FS を先に収束させてから profile ポインタを最後に移す](requirements/rollback-refit-then-pointer.md) — rollback は FS を先に収束させてから profile ポインタを最後に移す
- [module 時は rollback を host へ一本化し、layat profile は前進のみで追従する](requirements/module-mode-profile-internal.md) — module 時は rollback を host へ一本化し、layat profile は前進のみで追従する
- [project mode は世代を非公開にし、project / system mode は derivation 同一なら世代を積まず lstat ドリフト修復だけ行う](requirements/project-mode-generation-skip.md) — project mode は世代を非公開にし、project / system mode は derivation 同一なら世代を積まず lstat ドリフト修復だけ行う
- [孤児 profile は backref で逆引き可能なまま放置許容とし、`layat prune` が root 不在の系列を削除する](requirements/project-mode-orphan-profile.md) — 孤児 profile は backref で逆引き可能なまま放置許容とし、`layat prune` が root 不在の系列を削除する
- [同一 target を複数 config で狙うことによる振動はユーザー責任とし warning で可視化するに留める](requirements/cross-config-target-oscillation.md) — 同一 target を複数 config で狙うことによる振動はユーザー責任とし warning で可視化するに留める

---

## root の解決

root は評価時にパスへ展開せず、マーカーが運ぶ root の種別をエンジンが実行時に解決する
（→ [root は 3 つの marker で表す](adr/three-root-markers.md)）。

- [home mode の root は層ごとに定まった供給元から解決する](requirements/home-mode-root-resolution.md) — home mode の root は層ごとに定まった供給元から解決する
- [project mode の root は git toplevel から解決し、config 相対も CWD 相対も採らない](requirements/project-mode-root-resolution.md) — project mode の root は git toplevel から解決し、config 相対も CWD 相対も採らない
- [project mode の配置物は ephemeral とし、activation は git 状態に干渉しない](requirements/project-mode-ephemeral.md) — project mode の配置物は ephemeral とし、activation は git 状態に干渉しない
- [profileDir は home と system のみ name 直キーとし、fixed root と --root 上書きは roothash でキーする](requirements/profiledir-keying.md) — profileDir は home と system のみ name 直キーとし、fixed root と --root 上書きは roothash でキーする
- [system mode は root を / に解決し、変更系操作の前に euid 0 を確認し、profile と last-applied 記録を system 基底に置く](requirements/system-mode-root-gate-and-record-base.md) — system mode は root を / に解決し、変更系操作の前に euid 0 を確認し、profile と last-applied 記録を system 基底に置く
- [out-of-store の src は評価時の絶対パスか配置 root 相対のどちらかで、root 相対は target と同じ root に対して実行時に解決する](requirements/out-of-store-eval-time-path.md) — out-of-store の src は評価時の絶対パスか配置 root 相対のどちらかで、root 相対は target と同じ root に対して実行時に解決する

---

## モジュールオプション仕様・モジュール別動作仕様

全モジュール（HM / NixOS / nix-darwin）と devShell は、root と activation タイミングを供給して
エンジンを起動する薄い配線に徹する（→ [モジュールは root と起動の時機だけを供給して engine を起動する](adr/modules-wire-engine-kick-only.md)）。ネイティブ機構へは翻訳しない。

- [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](requirements/module-common-options.md) — 全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する
- [HM モジュール経由でも名前つき config ごとに役割分離した独立 profile を取れる](requirements/hm-named-configs-profiles.md) — HM モジュール経由でも名前つき config ごとに役割分離した独立 profile を取れる
- [layat.backup は engine 起動の配線レイヤーのオプションで manifest には影響しない](requirements/backup-wiring-layer.md) — layat.backup は engine 起動の配線レイヤーのオプションで manifest には影響しない
- [layat.recopy は engine 起動の配線レイヤーのオプションで、true のときだけ activation の起動へ --recopy を付ける](requirements/recopy-wiring-layer.md) — layat.recopy は engine 起動の配線レイヤーのオプションで、true のときだけ activation の起動へ --recopy を付ける
- [HM モジュールは `configs.<name>.subdir` が非 null の config を homeRoot の subdir 適用形で組み、全 target を `$HOME/<subdir>/` 配下に置く](requirements/hm-config-subdir-wiring.md) — HM モジュールは `configs.<name>.subdir` が非 null の config を homeRoot の subdir 適用形で組み、全 target を `$HOME/<subdir>/` 配下に置く
- [全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない](requirements/modules-are-engine-wiring.md) — 全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない
- [home-manager モジュールの engine kick は activation から全 config のビルド済み link-farm を 1 回の起動で渡し、失敗で switch を止める](requirements/hm-activation-contract.md) — home-manager モジュールの engine kick は activation から全 config のビルド済み link-farm を 1 回の起動で渡し、失敗で switch を止める
- [HM の activation は全 config の link-farm を apply --all 1 回の起動に渡し、報告は辞書順・配置の並列度は --jobs に従う](requirements/hm-activation-per-config-kick.md) — HM の activation は全 config の link-farm を apply --all 1 回の起動に渡し、報告は辞書順・配置の並列度は --jobs に従う
- [単一 HM config 内の configs 間 target 衝突は activation の apply --all 前段の確認で停止する](requirements/hm-configs-target-collision-assertion.md) — 単一 HM config 内の configs 間 target 衝突は activation の apply --all 前段の確認で停止する
- [NixOS / nix-darwin モジュールは activation から全 config のビルド済み link-farm を systemRoot で 1 回の起動に渡す](requirements/system-module-activation-contract.md) — NixOS / nix-darwin モジュールは activation から全 config のビルド済み link-farm を systemRoot で 1 回の起動に渡す
- [devShell は shellHook から engine を起動する配線で、シェル入室のたびに project mode で配置する](requirements/devshell-wiring.md) — devShell は shellHook から engine を起動する配線で、シェル入室のたびに project mode で配置する

---

## エラー仕様

設定の誤りは評価時に、実体の不整合はエンジン実行時に検出する層分けを守る。要求された操作が
成立しないときは暗黙のフォールバックを採らず、エラーで停止する。

- [設定の誤りは評価時に、実体の不整合は engine 実行時に検出する層分けを守る](requirements/error-detection-layering.md) — 設定の誤りは評価時に、実体の不整合は engine 実行時に検出する層分けを守る
- [要求された操作が成立しないときは CLI がエラーで停止し、暗黙のフォールバックを採らない](requirements/cli-stop-conditions.md) — 要求された操作が成立しないときは CLI がエラーで停止し、暗黙のフォールバックを採らない
- [配置元の実在は判定できる層で確認し、いずれの層でも停止する](requirements/src-subpath-existence-checks.md) — 配置元の実在は判定できる層で確認し、いずれの層でも停止する
- [project mode で git から root を解決できないときは engine 実行時に停止する](requirements/project-mode-root-resolution-failure.md) — project mode で git から root を解決できないときは engine 実行時に停止する
- [空の entries は正当な全クリアとして扱い、エラーにも警告にもしない](requirements/empty-entries-full-clear.md) — 空の entries は正当な全クリアとして扱い、エラーにも警告にもしない
- [conflict で停止するときは全件を対処ガイダンス付きで列挙してから 1 本の集約エラーを返す](requirements/conflict-full-report.md) — conflict で停止するときは全件を対処ガイダンス付きで列挙してから 1 本の集約エラーを返す

---

## 依存関係

- [engine は第三者依存ゼロの stdlib-only とし内部層に閉じる](requirements/engine-stdlib-only.md) — engine は第三者依存ゼロの stdlib-only とし内部層に閉じる
- [modules/common.nix は nixpkgs.lib のみに依存する](requirements/modules-common-nixpkgs-lib-only.md) — modules/common.nix は nixpkgs.lib のみに依存する
- [CLI が持ち込む依存は許可した第三者ライブラリと pin した Go に限り、いずれも固定する](requirements/cli-dependency-policy.md) — CLI が持ち込む依存は許可した第三者ライブラリと pin した Go に限り、いずれも固定する

---

## 品質

`quality` は開発プロセス・規約・ガバナンスに関する方針を持つ。`requirement` とは別系統で、
`use_case` を経由せず `solution` 直下に接続する。方針を実現する仕組みは `docs/infrastructure/` の
item が持ち、`satisfies` で `quality` へ接続する。

- [設計判断は ADR に記録し、改訂の書き戻しと item への接続を同じ変更の中で完了させる](quality/adr-decisions-recorded-and-connected.md) — 設計判断は ADR に記録し、改訂の書き戻しと item への接続を同じ変更の中で完了させる
- [リリースはバージョンの一次情報の変更で駆動し、手作業の工程を挟まない](quality/release-driven-by-source-of-truth.md) — リリースはバージョンの一次情報の変更で駆動し、手作業の工程を挟まない
- [リファレンスドキュメントは記述対象のソースから生成し、生成物を持たない](quality/reference-docs-generated-from-source.md) — リファレンスドキュメントは記述対象のソースから生成し、生成物を持たない
- [用語の正名は用語 item が固定し、執筆はそれに従う](quality/glossary-fixes-canonical-terms.md) — 用語の正名は用語 item が固定し、執筆はそれに従う
- [最新 main のビルド成果物を再ビルドなしに消費できる状態を保つ](quality/prebuilt-artifact-availability.md) — 最新 main のビルド成果物を再ビルドなしに消費できる状態を保つ
- [公開ドキュメントは英語を canonical とし、日本語版を対で保守する](quality/public-docs-english-canonical.md) — 公開ドキュメントは英語を canonical とし、日本語版を対で保守する
- [規範は 1 ファイル 1 item のグラフが持ち、概要文書は索引に縮退させる](quality/normative-items-placement-and-authoring.md) — 規範は 1 ファイル 1 item のグラフが持ち、概要文書は索引に縮退させる
- [ドキュメントグラフのトレーサビリティは機械的に検証する](quality/document-graph-mechanically-verified.md) — ドキュメントグラフのトレーサビリティは機械的に検証する
- [欠陥はトラッカーが持ちドキュメントグラフは規範のみを持つ。起票は分類語彙を経由する](quality/defects-in-tracker-graph-holds-norms.md) — 欠陥はトラッカーが持ちドキュメントグラフは規範のみを持つ。起票は分類語彙を経由する
- [傾向の計測は報告に留め、マージの条件にしない](quality/measurement-reports-without-gating.md) — 傾向の計測は報告に留め、マージの条件にしない
- [完成の定義は単一の機械パース可能な文書で持ち、項目数を上限で縛る](quality/dod-single-machine-parsable-document.md) — 完成の定義は単一の機械パース可能な文書で持ち、項目数を上限で縛る
- [マージ前の自動検証を必須にし、プラットフォーム差が出る層は全プラットフォームで通す](quality/cross-platform-verification-before-merge.md) — マージ前の自動検証を必須にし、プラットフォーム差が効く層は全プラットフォームで通す
- [コード整形と静的解析を自動検証に載せ、同じ判定を手元でも得られるようにする](quality/formatting-and-static-analysis-automated.md) — コード整形と静的解析を自動検証に載せ、同じ判定を手元でも得られるようにする
- [自動化が取り込む実行物は不変な識別子で固定し、権限は最小に絞り、不正入力では成果物を作る前に失敗する](quality/automation-supply-chain-safety.md) — 自動化が取り込む実行物は不変な識別子で固定し、権限は最小に絞り、不正入力では成果物を作る前に失敗する
- [規範の型は、それが誰を縛るかで決める](quality/norm-type-by-who-it-binds.md) — 規範の型は、それが誰を縛るかで requirement・quality・test_plan から選ぶ
- [risk はまず requirement に張り、設計を差し替えると消える懸念だけを design に張る](quality/risk-attaches-to-requirement-by-default.md) — risk はまず requirement に張り、設計を差し替えると消える懸念だけを design に張る
- [risk の level は likelihood と impact から導き、手で決めない](quality/risk-level-derived-from-likelihood-and-impact.md) — risk の likelihood と impact を決まった尺度で採点し、level はその 2 つから導く
- [specification は SHALL 系で書き、日本語の規範文は同じ強度の規範助動詞で終える](quality/specification-shall-family-and-matching-auxiliaries.md) — specification は SHALL 系で書き、specification_ja は同じ強度の規範助動詞で終える
- [テスト系 item は、テスト資産・検証テーマ・脅威の 3 段の粒度で起こす](quality/test-items-three-tier-granularity.md) — テスト系 item は CASE をテスト資産・TC を検証テーマ・RISK を脅威の単位で起こし、概要文書へ索引しない
- [コメントと利用者向け文字列は現在の動作だけを書き、決定の記録を参照しない](quality/comments-state-current-behaviour-only.md) — コメントは 3 行以内で現在の動作だけを書き、経緯と決定の記録への参照を書かない。利用者向け文字列に ADR 番号を出さない
- [フル ID は frontmatter だけに書き、散文では item にリンクで触れる](quality/full-ids-in-frontmatter-links-in-prose.md) — ID はフル UUIDv4 で frontmatter だけに書き、ファイル名は slug とし、散文では name を文字列にしたリンクで item に触れる
- [ADR は 1 本に 1 つの決定を 4 節で書き、justifies は規範の形成に寄与した item へ張る](quality/adr-one-decision-four-sections-justifies-by-contribution.md) — ADR は単独で覆せる決定ごとに 1 本を 4 節で書き、justifies は規範の形成に寄与した item へ張る
- [上流より古い item と generated の更新漏れは機械的に列挙する](quality/stale-items-enumerated-mechanically.md) — 上流より基準時刻が古い下流の item と、本文を変えたのに generated を更新していない item を機械的に列挙し、結果は報告に留める
- [level が high の risk は人が確かめた記録を持つ](quality/high-risk-verified-by-human.md) — level が high の risk は、generated.at 以降に人が内容を確かめた verified を持つ

---

## テスト計画

`test_plan` は `requirement` とは別系統で、`use_case` を経由せず `solution` 直下に接続する。

- [home-manager モジュールの配線は build sandbox 内の評価アサートで検証する](test-plan/hm-module-eval-assert.md) — home-manager モジュールの配線は build sandbox 内の評価アサートで検証する
- [非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する](test-plan/e2e-harness-scope.md) — 非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する
- [評価テストは nix-unit で不変条件を、namaka で manifest 全体のスナップショットを見る](test-plan/nix-unit-namaka-split.md) — 評価テストは nix-unit で不変条件を、namaka で manifest 全体のスナップショットを見る
- [lib.__internal で private helper をテストへ公開する](test-plan/lib-internal-test-seam.md) — lib.\_\_internal は private helper のテスト seam として公開する
- [NixOS / nix-darwin モジュールで使うときの実 activate は E2E ハーネスの対象外とする](test-plan/nixos-vm-test-future.md) — NixOS / nix-darwin モジュールで使うときの実 activate は E2E ハーネスの対象外とする
- [エンベロープの outturn 適合を Go テストと E2E の両方で検証する](test-plan/json-conformance-verification.md) — エンベロープの outturn 適合を Go テストと E2E の両方で検証する
- [評価テストの store-backed な入力は固定 outPath を持つ fake flake-input で与える](test-plan/eval-test-double.md) — 評価テストの store-backed な入力は固定 outPath を持つ fake flake-input で与える
- [原子性は実 FS の条件で故障を誘発して不変条件ごとに検証する](test-plan/fault-injection-atomicity.md) — 原子性は実 FS の条件で故障を誘発して不変条件ごとに検証する
- [エンジンとコマンド層の Go テストは nix を介さない実 FS 統合テストを主戦力とする](test-plan/go-test-layering.md) — エンジンとコマンド層の Go テストは nix を介さない実 FS 統合テストを主戦力とする

---

## 設定の書き方（本文書の対象外）

実際に動く設定例は本文書では扱わない。`templates/`（`layat init` が展開する実物）と README の
コード例を参照する。本文書が扱うのは「何を満たすべきか」であり、書き方の例ではない。

---

## 関連文書

- `README.md` / `README.ja.md` — 3 層構造の最上段（導入と使い方）
- `docs/concept.md` — コンセプト（`solution` / `use_case` への索引）
- `docs/design.md` — 設計（`design` item への索引）
- `docs/infrastructure/` — 品質方針を実現する技術基盤（CI / リリース / 配信）
- `docs/adr/` — 意思決定の記録。`requirement` / `design` / `quality` / `test_plan` などを `justifies` で裏づける
- `docs/model.yaml` — sara の型定義（item の型・関係・ID 形式）
