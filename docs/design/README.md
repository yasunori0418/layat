# layat 設計書

要求（`requirement`）とテスト計画（`test_plan`）を「どう実現するか」の全体像と、個別設計
（`design` item）への索引。

設計判断は **すべて `docs/design/` の item が持ち**、その決定と根拠は `docs/adr/` の ADR が
`justifies` で裏づける。本文書は通読の起点として全体像だけを述べ、詳細は item へのリンクで
示す（README → 本文書 → item の 3 層構造）。

> **この文書の書き方（規約）**
>
> - 散文は **見出し 1 つ（h2 / h3 の最も内側）につき 10 行以内**。表・リンク列挙・コード
>   ブロックはこの制限に含めない
> - **詳細は必ず item へリンクし、本文に書き下さない**。設計の断片をここへ写すと item と
>   二重管理になり、改訂時に片方だけ古くなる。**例外は「設計目標」の表**で、これは設計判断を
>   評価する軸の横断ビューであり、対応する item を持たないため本文に置く。各目標がどう
>   満たされるかは `requirement` / `design` item が持ち、表には書かない
> - `design` / `infrastructure` item を足したら本文書のリンク集にも足す。逆に本文書へ設計を
>   書き足さない
> - 「何を満たすべきか」は `requirement` の領分で本文書は扱わない（→ `docs/spec.md`）
> - item を横断して検索・追跡するには `sara query`（→ `CLAUDE.md` の「item を辿る（`sara query`）」節）を使う
> - item の `## 出典` は本文書の現行章立てを指さない。原文の読み方を含め `docs/spec.md` の
>   同項が正（同項は 3 文書共通の手順として書いてある）

---

## 概要

nix store のパス（リポジトリ全体・サブディレクトリ・単一ファイル）を、root 相対の任意パスへ
symlink または copy で配置する Nix ライブラリ・モジュール群。配置ロジックはテスト可能な純粋関数 +
単一の配置エンジンとして実装し、ユーザーが配置を明示的に握る。home-manager に依存せず単体で
動作しつつ、HM / NixOS / nix-darwin のモジュールとも統合できる。ただし統合層は配置ロジックを
持たず、エンジンを起動する薄い配線に徹する（→ [モジュールは root と起動の時機だけを供給して engine を起動する](../adr/modules-wire-engine-kick-only.md)）。

layat が何であるか・何のために作るかは `solution` / `use_case` の領分。→ `docs/concept.md`。

---

## 設計目標

設計判断を評価する軸の一覧。**各軸が何を要求するか（規範）は `requirement` が持ち、どう実現するかは
`design` item が持つ**。ここは軸の名前と着眼点だけを並べ、満たし方は書かない。

| 目標 | 着眼点 |
|---|---|
| 純粋性・テスト可能性 | 配置ロジックをモジュールに依存せずテストできるか |
| 独立性 | home-manager 不在の環境で単体で動くか |
| 統合性 | 既存のモジュールシステムから起動できるか |
| 柔軟性 | モジュールシステムを介さずに使えるか |
| 取得手段非依存 | 取得方法の変化が layat の変更を要求しないか |
| 粒度 | リポジトリ全体・サブディレクトリ・単一ファイルを同じ扱いにできるか |
| 更新の独立性 | 1 つの更新が他の配置単位へ波及しないか |
| 非生成 | ファイルの内容に関与していないか |
| 世代管理 | 適用を取り消せるか |
| root 明示 | 配置先が暗黙のデフォルトに落ちていないか |

---

## プロジェクト構成

リポジトリはトップレベル 5 ディレクトリ（`lib/` / `cmd/` / `internal/` / `templates/` /
`modules/`）に分ける。配置ロジックは Go エンジンが単一の源として所有し、`lib/` はデータ生成に
徹する（→ [lib は純データ生成に徹する](../adr/lib-produces-data-only.md), [engine は cmd/ と internal/ に置きライブラリ化しない](../adr/engine-internal-cmd-layout.md)）。

- [リポジトリを lib / cmd / internal / templates / modules のトップレベル 5 ディレクトリへ分ける](./repo-layout.md) — リポジトリを lib / cmd / internal / templates / modules のトップレベル 5 ディレクトリへ分ける
- [lib は公開 API・型定義・manifest 生成・マーカー構築子・apply コマンド組み立ての 5 ファイルへ分割する](./lib-module-split.md) — lib は公開 API・型定義・manifest 生成・マーカー構築子・apply コマンド組み立ての 5 ファイルへ分割する
- [Go 側は cmd/layat に CLI 面を、internal に配置ロジックを置く 2 パッケージ構成にする](./go-package-split.md) — Go 側は cmd/layat に CLI 面を、internal に配置ロジックを置く 2 パッケージ構成にする
- [modules は common.nix に共通オプションを集約し、統合層ごとに 1 ファイルへ分ける](./modules-file-split.md) — modules は common.nix に共通オプションを集約し、統合層ごとに 1 ファイルへ分ける
- [実装スコープを CLI 単体 + project mode + home mode + system mode と、home-manager / NixOS / nix-darwin モジュールとする](./implementation-scope.md) — 実装スコープを CLI 単体 + project mode + home mode + system mode と、home-manager / NixOS / nix-darwin モジュールとする

---

## レイヤー構成

CLI / engine / lib / `common.nix` / 統合層の 5 段に積み、依存は呼ぶ側から呼ばれる側への一方向に
限る。配置の振る舞いは全層でエンジンが単一の源であり、各層はネイティブ機構へ翻訳しない
（→ [配置は全層で engine が行い、ネイティブ機構へ翻訳しない](../adr/engine-owns-placement-no-native-translation.md), [engine を単一の固定 Go バイナリにする](../adr/engine-fixed-go-binary.md)）。

- [層を CLI / engine / lib / common.nix / 統合層の 5 段に積み、依存を呼ぶ側から呼ばれる側への一方向に限る](./layer-stack-dependency.md) — 層を CLI / engine / lib / common.nix / 統合層の 5 段に積み、依存を呼ぶ側から呼ばれる側への一方向に限る
- [internal/generator は契約・prebuilt・Fake を親パッケージに、nix 実装を nixgen サブパッケージに置き、cmd 層は注入点だけを持つ](./generator-package-layout.md) — internal/generator は契約・prebuilt・Fake を親パッケージに、nix 実装を nixgen サブパッケージに置き、cmd 層は注入点だけを持つ
- [internal/generation は契約・Generation 型・FS ベースの Fake を親パッケージに、nix-env 実装を nixprofile サブパッケージに置き、engine は Backend 1 フィールドで受ける](./generation-package-layout.md) — internal/generation は契約・Generation 型・FS ベースの Fake を親パッケージに、nix-env 実装を nixprofile サブパッケージに置き、engine は Backend 1 フィールドで受ける
- [last-applied 記録は `<state>/layat/<鍵>/last-applied.json` に manifest をそのまま埋め込み、比較の基準は記録 → profile → 空の順に読む](./last-applied-record-layout.md) — last-applied 記録は `<state>/layat/<鍵>/last-applied.json` に manifest をそのまま埋め込み、比較の基準は記録 → profile → 空の順に読む

---

## flake.nix outputs 設計

layat 自身の flake outputs は packages / templates / 各モジュール / flakeModules / lib の 5 系統。
ユーザーの entrypoint 側は `layat.<name>` に named manifest を公開し、直書きと flake-parts の
2 つの使い方が同一の derivation を生む（→ [named manifest を専用の `layat.<name>` 名前空間に公開する](../adr/layat-name-namespace.md), [flake-parts module の option に mkManifest の結果を入れる](../adr/flake-parts-option-holds-mkmanifest-result.md), [CLI が評価する attr path を legacy entrypoint の書き方で分岐させない](../adr/attr-path-independent-of-entrypoint-shape.md)）。

- [layat 自身の flake outputs は packages / templates / 各モジュール / flakeModules / lib の 5 系統で構成する](./flake-outputs-attrs.md) — layat 自身の flake outputs は packages / templates / 各モジュール / flakeModules / lib の 5 系統で構成する
- [NixOS / nix-darwin モジュールは modules/kick.nix を介した system mode の配線とする](./module-stubs.md) — NixOS / nix-darwin モジュールは modules/kick.nix を介した system mode の配線とする
- [動的 entry 生成のヘルパを lib に置く](./lib-no-dynamic-entries.md) — 棄却した設計。readDir する idiom をドキュメントで示す
- [legacy entrypoint の分岐は attr path 組み立てに閉じ、nix 呼び出しヘルパを共通で再利用する](./legacy-nix-invocation-reuse.md) — legacy entrypoint の分岐は attr path 組み立てに閉じ、nix 呼び出しヘルパを共通で再利用する
- [hook 機構を engine に持たず、配線層ごとの合成 idiom で hook 需要を満たす](./hook-composition-idioms.md) — hook 機構を engine に持たず、配線層ごとの合成 idiom で hook 需要を満たす
- [buildEnv の合成環境は entry の src に置く配置 idiom で扱い、環境変数の配線は利用者側の idiom で行う](./buildenv-placement-env-wiring-idioms.md) — buildEnv の合成環境は entry の src に置く配置 idiom で扱い、環境変数の配線は利用者側の idiom で行う

---

## コアロジック設計（lib データ生成 + Go エンジン）

`lib.mkManifest` が link farm derivation（`manifest.json` + symlink ツリー）を生成し、エンジンが
実行時に前世代 manifest と diff して配置・古くなった配置の除去・profile swap を行う（→ [古くなった配置の除去は manifest の差分で行う](../adr/stale-removal-by-manifest-diff.md), [manifest.json を schemaVersion 1 の唯一の契約として固定する](../adr/manifest-json-contract-schemaversion-v1.md)）。

- [配置プリミティブを store link・copy・out-of-store symlink の 3 種に定め、世代・reset・gitignore での扱いを種別ごとに固定する](./placement-primitives.md) — 配置プリミティブを store link・copy・out-of-store symlink の 3 種に定め、世代・reset・gitignore での扱いを種別ごとに固定する
- [配置と cleanup のアルゴリズムは home-manager の linkGeneration / cleanup を範として Go で再実装する](./cleanup-from-home-manager.md) — 配置と cleanup のアルゴリズムは home-manager の linkGeneration / cleanup を範として Go で再実装する
- [配置ロジックを plan を計算する planner と FS 意味論を実行する engine に分け、dryrun と apply が同一の plan を共有する](./planner-engine-split.md) — 配置ロジックを plan を計算する planner と FS 意味論を実行する engine に分け、dryrun と apply が同一の plan を共有する
- [prune を paths の列挙・engine の判定と削除・CLI の同意解釈に分け、engine API を Reset と同型にする](./prune-orphan-series.md) — prune を paths の列挙・engine の判定と削除・CLI の同意解釈に分け、engine API を Reset と同型にする
- [rootRelative の LinkDest は target の親から src/subpath への相対パスを Entry だけから計算し、root が要るのは配置前の確認だけにする](./root-relative-linkdest.md) — rootRelative の LinkDest は target の親から src/subpath への相対パスを Entry だけから計算し、root が要るのは配置前の確認だけにする
- [status を engine の記録基底走査と CLI の絞り込み・出力に分ける](./status-scans-record-bases.md) — status を engine の記録基底走査と CLI の絞り込み・出力に分ける
- [engine のエラーは発生源で 1 回だけ op と対象パスを付けて wrap し、伝搬では素通しする](./engine-error-wrap-once-at-origin.md) — engine のエラーは発生源で 1 回だけ op と対象パスを付けて wrap し、伝搬では素通しする

`lib` の API・`entries` スキーマ・`manifest.json` の内容・CLI の実行モデルは仕様（`requirement`）の
領分（→ `docs/spec.md`「lib API」「entries スキーマ仕様」「manifest.json スキーマ」「CLI 仕様」）。

---

## モジュール統合設計

全モジュールと devShell は root と activation タイミングを供給してエンジンを起動する薄い配線。
起動の差は link-farm の取得方法だけに閉じる（→ [モジュールは root と起動の時機だけを供給して engine を起動する](../adr/modules-wire-engine-kick-only.md), [apply --manifest でビルド済み link-farm を engine へ直接渡す](../adr/apply-manifest-prebuilt-link-farm.md)）。

- [engine の起動を entrypoint 駆動とビルド済み manifest の 2 クラスに束ね、統合層の差を link-farm の取得方法だけに閉じる](./integration-layer-kick-classes.md) — engine の起動を entrypoint 駆動とビルド済み manifest の 2 クラスに束ね、統合層の差を link-farm の取得方法だけに閉じる

各モジュールが公開するオプションと動作は仕様の領分
（→ `docs/spec.md`「モジュールオプション仕様・モジュール別動作仕様」）。

---

## 使用パターン

→ `docs/concept.md`「想定する使われ方」（`use_case` 7 件）。個々の設定の書き方はテンプレート
（`layat init`）と `templates/` の実物を参照する。

---

## テスト戦略

lib は Nix 評価テスト、エンジンは実 FS の tmpdir 統合テスト、非 NixOS で動く主張は実 nix の
E2E で検証する（→ [テストを sandbox 可否で flake check と E2E に分ける](../adr/test-strategy-layers.md)）。

- [lib は nix-unit の評価テストと namaka のスナップショットの 2 手段で検証する](./lib-test-nix-unit-namaka.md) — lib は nix-unit の評価テストと namaka のスナップショットの 2 手段で検証する
- [engine は nix を使わない tmpdir 統合テストで検証し、古くなった配置の除去の安全不変条件を table-driven に重点配分する](./engine-test-go-unit-tmpdir.md) — engine は nix を使わない tmpdir 統合テストで検証し、古くなった配置の除去の安全不変条件を table-driven に重点配分する
- [E2E は tests/e2e の bash ハーネスに置き、シナリオの詳細は同ディレクトリの README へ委譲する](./e2e-harness-shape.md) — E2E は tests/e2e の bash ハーネスに置き、シナリオの詳細は同ディレクトリの README へ委譲する
- [E2E に legacy entrypoint シナリオを置き、NIX_PATH を flake.lock の nixpkgs に pin して検証する](./e2e-legacy-entrypoint.md) — E2E に legacy entrypoint シナリオを置き、NIX_PATH を flake.lock の nixpkgs に pin して検証する
- [NixOS モジュールの実 activate は runNixOSTest の checks.nixos-vm で検証し、E2E ハーネスとは別系統に置く](./nixos-vm-test-position.md) — NixOS モジュールの実 activate は runNixOSTest の checks.nixos-vm で検証し、E2E ハーネスとは別系統に置く

CI での実行基盤・キャッシュ・リリースは `infrastructure` item の領分。

- [CI パイプライン（flake check の os×system マトリクスと E2E ジョブ）](../infrastructure/ci-pipeline.md) — CI パイプライン（flake check の os×system マトリクスと E2E ジョブ）
- [バイナリキャッシュ（main push 時の cachix 投入）](../infrastructure/binary-cache.md) — バイナリキャッシュ（main push 時の cachix 投入）
- [リリース自動化（VERSION ファイル起点の bump PR・自動タグ・自動リリースノート）](../infrastructure/release-automation.md) — リリース自動化（VERSION ファイル起点の bump PR・自動タグ・自動リリースノート）
- [マージの条件（main の ruleset と required status checks）](../infrastructure/merge-gate.md) — マージの条件（main の ruleset と required status checks）
- [トレーサビリティ検証基盤（sara によるドキュメントグラフの機械検証）](../infrastructure/traceability-verification.md) — トレーサビリティ検証基盤（sara によるドキュメントグラフの機械検証）
- [ドキュメントサイト（Astro Starlight + Cloudflare Pages・ビルド時リファレンス生成）](../infrastructure/docs-site.md) — ドキュメントサイト（Astro Starlight + Cloudflare Pages・ビルド時リファレンス生成）

---

## 設計判断の所在

本節はリンク集ではなく、決定がどこにあるかの案内。主要な設計判断の決定と根拠は **ADR が持つ**
（`docs/adr/`）。`docs/model.yaml` は ADR の `justifies` の接続先として `requirement` / `design` /
`infrastructure` / `quality` / `test_plan` の 5 型を許す（現時点で実際に張られているのは `quality` を
除く 4 型）。ある item を裏づける決定は本文書の一覧ではなく `sara query <フル ID> -u` で辿る。
全 ADR をここへ並べても改訂のたびに古くなるだけなので置かない。

---

## 関連文書

- `README.md` / `README.ja.md` — 3 層構造の最上段（導入と使い方）
- `docs/concept.md` — コンセプト（`solution` / `use_case` への索引）
- `docs/spec.md` — 仕様（`requirement` / `quality` / `test_plan` item への索引）
- `docs/adr/` — 意思決定の記録
- `docs/model.yaml` — sara の型定義（item の型・関係・ID 形式）
