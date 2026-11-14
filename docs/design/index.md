# 設計（`design`）

## プロジェクト構成

* [Go 側は cmd/layat に CLI 面を、internal に配置ロジックを置く 2 パッケージ構成にする](./go-package-split.md) - [layat は CLI とエンジンの 2 層で構成する](../requirements/two-layer-architecture.md) が定める CLI と engine の 2 層を、そのまま Go のパッケージ境界へ落とす。`packages.layat` は両者（`cmd/layat` + `internal`）をまとめてビルドした 1 バイナリになる。
* [実装スコープを CLI 単体 + project mode + home mode + system mode と、home-manager / NixOS / nix-darwin モジュールとする](./implementation-scope.md) - [layat CLI は PATH 常駐の一次 UX で、project mode は devShell 同梱を canonical とする](../requirements/cli-primary-ux-installation.md) という要求がある以上、**最初に完成させるべきは CLI と project mode**になる。home mode（`homeRoot`）を同時に含めるのは、そこが rollback・世代管理という layat 固有機構を通る唯一の使い方であり、ここを外すと世代機構が実装されないまま残るため。
* [lib は公開 API・型定義・manifest 生成・マーカー構築子・apply コマンド組み立ての 5 ファイルへ分割する](./lib-module-split.md) - 分割の軸は「公開 API 面 / 型 / 生成ロジック / マーカー / コマンド組み立て」の 5 つで、次の理由による。
* [modules は common.nix に共通オプションを集約し、統合層ごとに 1 ファイルへ分ける](./modules-file-split.md) - **`common.nix` を options 定義のみに絞り、全モジュールが import する**のが分割の要。[全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](../requirements/module-common-options.md) は「モジュールごとに集合が分岐しない」ことを求めるが、これは各モジュールが自前で options を定義していると守れない（片方にだけ足す・デフォルトがずれる、といった分岐が構造的に起きる）。定義を 1 ファイルに置いて import させれば、集合の同一性がファイル参照の一意性として担保される。
* [リポジトリを lib / cmd / internal / templates / modules のトップレベル 5 ディレクトリへ分ける](./repo-layout.md) - `flake.nix` を唯一の entrypoint とし、その下にトップレベル 5 ディレクトリを並置する。

## レイヤー構成

* [internal/generation は契約・Generation 型・FS ベースの Fake を親パッケージに、nix-env 実装を nixprofile サブパッケージに置き、engine は Backend 1 フィールドで受ける](./generation-package-layout.md) - internal/generation は契約・Generation 型・世代リンクの走査・FS ベースの Fake を親パッケージに置き、nix-env 実装を nixprofile サブパッケージに閉じ込め、engine は Backend を 1 フィールドで受けて nil を error にする
* [internal/generator は契約・prebuilt・Fake を親パッケージに、nix 実装を nixgen サブパッケージに置き、cmd 層は注入点だけを持つ](./generator-package-layout.md) - internal/generator は契約・prebuilt 実装・Fake を親パッケージに、nix 固有の知識を nixgen サブパッケージに置き、cmd 層は newGenerator() で選んだ生成器を各サブコマンドへ注入するだけにする
* [last-applied 記録は `<state>/layat/<鍵>/last-applied.json` に manifest をそのまま埋め込み、比較の基準は記録 → profile → 空の順に読む](./last-applied-record-layout.md) - `origin` は prebuilt 生成器（`--manifest`）なら `{ "generator": "prebuilt" }` で、`entrypoint` を持たない。`identity` は配置した manifest の `identity` の写しで、manifest が持たなければ省く。reset / rollback が版 1 の記録を書き直す場合や、記録が無く profile から比較の基準を取って書き直す場合のように、出所が分からないときは `origin` 自体を省く。
* [層を CLI / engine / lib / common.nix / 統合層の 5 段に積み、依存を呼ぶ側から呼ばれる側への一方向に限る](./layer-stack-dependency.md) - **依存は呼ぶ側から呼ばれる側への一方向に限り、逆流させない**ことを層構造の規約とする。上の図は配置の駆動順（CLI が起点、統合層が起動配線）で層を並べたもので、層の上下は依存の向きを表さない。図中の連結子も一様ではなく、`import` は依存そのものを指すが`manifest.json を渡して起動` / `起動配線` は起動の関係を指す。依存の向きは図から読み取るものではなく、下の 3 点が定める。

## flake.nix outputs 設計

* [buildEnv の合成環境は entry の src に置く配置 idiom で扱い、環境変数の配線は利用者側の idiom で行う](./buildenv-placement-env-wiring-idioms.md) - 「パッケージ群を `PATH` に載せたい」需要に対し、layat は専用の関数（`mkEnv` 等）を持たない（→ [layat.lib.mkEnv で buildEnv 合成環境を提供する](../adr/mkenv-buildenv.md)）。`pkgs.buildEnv` の結果を entry の `src` に置けば、既存の配置プリミティブでそのまま満たせるため。
* [layat 自身の flake outputs は packages / templates / 各モジュール / flakeModules / lib の 5 系統で構成する](./flake-outputs-attrs.md) - 5 系統は、layat が提供するものの**使い方の違い**で分けている。要求は「何を提供するか」を定めるが、それを flake のどの attr 名で公開するかは設計側の判断になる。
* [hook 機構を engine に持たず、配線層ごとの合成 idiom で hook 需要を満たす](./hook-composition-idioms.md) - 「配置の前後で任意の処理を実行したい」（配置後のサービスリロード・通知・後処理）需要に対し、**engine はイベント hook 機構を提供しない**（→ [engine の実行イベントにユーザー定義スクリプトの hook を掛ける](../adr/engine-event-hooks.md)）。代わりに、需要を満たす手段を**配線層ごとの合成 idiom** としてここに示す。
* [legacy entrypoint の分岐は attr path 組み立てに閉じ、nix 呼び出しヘルパを共通で再利用する](./legacy-nix-invocation-reuse.md) - flake entrypoint と legacy entrypoint（`shell.nix` / `default.nix`）は、CLI が叩くnix コマンドの形が異なる。
* [NixOS / nix-darwin モジュールは modules/kick.nix を介した system mode の配線とする](./module-stubs.md) - `nixosModules.default`（`modules/nixos.nix`）と `darwinModules.default`（`modules/nix-darwin.nix`）は、`common.nix` の共通オプションを import し、`modules/kick.nix` が返す起動文字列を activation へ置くだけの配線である。

## コアロジック設計（lib データ生成 + Go エンジン）

* [配置と cleanup のアルゴリズムは home-manager の linkGeneration / cleanup を範として Go で再実装する](./cleanup-from-home-manager.md) - engine の配置・古くなった配置の除去のアルゴリズムは、home-manager の `linkGeneration` / `cleanup` を**参考にして Go で再実装する**。`home.file` そのものは再利用しない。
* [engine のエラーは発生源で 1 回だけ op と対象パスを付けて wrap し、伝搬では素通しする](./engine-error-wrap-once-at-origin.md) - 適用範囲は `internal/engine/` のエラーとする。エラーの組み立ては標準ライブラリの `fmt.Errorf` と`errors.Is` だけで行う（→ [engine は第三者依存ゼロの stdlib-only とし内部層に閉じる](../requirements/engine-stdlib-only.md)）。
* [配置プリミティブを store link・copy・out-of-store symlink の 3 種に定め、世代・reset・gitignore での扱いを種別ごとに固定する](./placement-primitives.md) - entry が置く物の種別（配置プリミティブ）は 3 つに限る。種別は entry の `method`（`symlink` / `copy`）と、`src` が out-of-store marker かどうかの 2 軸で決まる。2 × 2 のうち「`copy` × out-of-store marker」は意図が矛盾するので評価時に拒否する（→ [意図が矛盾する組み合わせをクロスフィールドの確認で評価時に拒否する](../requirements/normalize-manifest-cross-field-checks.md)）。残る 3 マスが次の 3 種になる。
* [配置ロジックを plan を計算する planner と FS 意味論を実行する engine に分け、dryrun と apply が同一の plan を共有する](./planner-engine-split.md) - `internal/` の配置ロジックを 2 層に分ける。
* [prune を paths の列挙・engine の判定と削除・CLI の同意解釈に分け、engine API を Reset と同型にする](./prune-orphan-series.md) - `layat prune` を 3 層に分ける。既存コマンドと同じ層境界（→ [Go 側は cmd/layat に CLI 面を、internal に配置ロジックを置く 2 パッケージ構成にする](go-package-split.md)）に乗せ、prune 専用の縦割りを作らない。
* [rootRelative の LinkDest は target の親から src/subpath への相対パスを Entry だけから計算し、root が要るのは配置前の確認だけにする](./root-relative-linkdest.md) - rootRelative の LinkDest は target の親から src/subpath への相対パスを Entry だけから計算する純関数とし、root は配置前に配置元の実在を確かめるときだけ使う
* [status を engine の記録基底走査と CLI の絞り込み・出力に分ける](./status-scans-record-bases.md) - `layat status` を 3 つに分担する。既存コマンドと同じ層境界（→ [Go 側は cmd/layat に CLI 面を、internal に配置ロジックを置く 2 パッケージ構成にする](go-package-split.md)）に乗せる。

## モジュール統合設計

* [engine の起動を entrypoint 駆動とビルド済み manifest の 2 クラスに束ね、統合層の差を link-farm の取得方法だけに閉じる](./integration-layer-kick-classes.md) - 統合層は 5 つ（CLI 単体 / home-manager / devShell / NixOS / nix-darwin）あるが、**engine の起動方法は 2 クラスに束ねられる**。配置から世代コミットまでは両クラスで同一の engine の処理を通り、異なるのは link-farm をどう手に入れるかだけになる。

## テスト戦略

* [E2E は tests/e2e の bash ハーネスに置き、シナリオの詳細は同ディレクトリの README へ委譲する](./e2e-harness-shape.md) - [非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する](../test-plan/e2e-harness-scope.md) が要求する一気通貫 E2E の実装形を次のように定める。
* [E2E に legacy entrypoint シナリオを置き、NIX_PATH を flake.lock の nixpkgs に pin して検証する](./e2e-legacy-entrypoint.md) - [非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する](../test-plan/e2e-harness-scope.md) がハーネスに課すシナリオ 5 件（project mode / home mode / 古くなった配置の除去 / copy place-once・out-of-store / HM module）に加えて、**legacy entrypoint シナリオ**をE2E に置く。
* [engine は nix を使わない tmpdir 統合テストで検証し、古くなった配置の除去の安全不変条件を table-driven に重点配分する](./engine-test-go-unit-tmpdir.md) - engine（`internal/`）は Go ユニットテスト + tmpdir 統合テストで検証する。統合テストは実 FS を使うが、source は偽物・nix は使わない。
* [lib は nix-unit の評価テストと namaka のスナップショットの 2 手段で検証する](./lib-test-nix-unit-namaka.md) - 2 手段を併用するのは、lib が生成するものが構造化データ 1 個であることに由来する。
* [NixOS モジュールの実 activate は runNixOSTest の checks.nixos-vm で検証し、E2E ハーネスとは別系統に置く](./nixos-vm-test-position.md) - NixOS モジュールを使う場面の実 activate は、**`runNixOSTest`（nixpkgs の VM テストフレームワーク）のflake check `checks.nixos-vm` で検証する**。`tests/e2e/` の bash ハーネス（[E2E は tests/e2e の bash ハーネスに置き、シナリオの詳細は同ディレクトリの README へ委譲する](e2e-harness-shape.md)）のスコープ外に置く。

## 有効でない item（`deprecated`）

* [動的 entry 生成のヘルパを lib に置く](./lib-no-dynamic-entries.md) - この設計は棄却した。「リポジトリ内のファイルを列挙して entries を組み立てる」需要に対し、**lib は関数を提供しない**。既に realise 済みの store パスまたは flake input を `builtins.readDir`する idiom をドキュメント側（[`README.md` の「Generating entries dynamically」](../../README.md#generating-entries-dynamically)）で示すに留める。
