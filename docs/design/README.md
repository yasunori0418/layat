# layat 設計書

要求（`requirement`）とテスト計画（`test_plan`）を「どう実現するか」の全体像を述べる散文。
`design` item の一覧は、生成した [index.md](index.md) が持つ。

設計判断は **すべて `docs/design/` の item が持ち**、その決定と根拠は `docs/adr/` の ADR が
`justifies` で裏づける。本文書は通読の起点として全体像だけを散文で述べ、item の一覧は持たない。

> **この文書の書き方（規約）**
>
> - 散文は **見出し 1 つ（h2 / h3 の最も内側）につき 10 行以内**。表・リンク列挙・コード
>   ブロックはこの制限に含めない
> - **詳細は必ず item へリンクし、本文に書き下さない**。設計の断片をここへ写すと item と
>   二重管理になり、改訂時に片方だけ古くなる。**例外は「設計目標」の表**で、これは設計判断を
>   評価する軸の横断ビューであり、対応する item を持たないため本文に置く。各目標がどう
>   満たされるかは `requirement` / `design` item が持ち、表には書かない
> - 本文書へ設計を書き足さない
> - 「何を満たすべきか」は `requirement` の領分で本文書は扱わない（→ `docs/requirements/README.md`）
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

layat が何であるか・何のために作るかは `solution` / `use_case` の領分。→ `docs/README.md`。

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

---

## レイヤー構成

CLI / engine / lib / `common.nix` / 統合層の 5 段に積み、依存は呼ぶ側から呼ばれる側への一方向に
限る。配置の振る舞いは全層でエンジンが単一の源であり、各層はネイティブ機構へ翻訳しない
（→ [配置は全層で engine が行い、ネイティブ機構へ翻訳しない](../adr/engine-owns-placement-no-native-translation.md), [engine を単一の固定 Go バイナリにする](../adr/engine-fixed-go-binary.md)）。

---

## flake.nix outputs 設計

layat 自身の flake outputs は packages / templates / 各モジュール / flakeModules / lib の 5 系統。
ユーザーの entrypoint 側は `layat.<name>` に named manifest を公開し、直書きと flake-parts の
2 つの使い方が同一の derivation を生む（→ [named manifest を専用の `layat.<name>` 名前空間に公開する](../adr/layat-name-namespace.md), [flake-parts module の option に mkManifest の結果を入れる](../adr/flake-parts-option-holds-mkmanifest-result.md), [CLI が評価する attr path を legacy entrypoint の書き方で分岐させない](../adr/attr-path-independent-of-entrypoint-shape.md)）。

---

## コアロジック設計（lib データ生成 + Go エンジン）

`lib.mkManifest` が link farm derivation（`manifest.json` + symlink ツリー）を生成し、エンジンが
実行時に前世代 manifest と diff して配置・古くなった配置の除去・profile swap を行う（→ [古くなった配置の除去は manifest の差分で行う](../adr/stale-removal-by-manifest-diff.md), [manifest.json を schemaVersion 1 の唯一の契約として固定する](../adr/manifest-json-contract-schemaversion-v1.md)）。

`lib` の API・`entries` スキーマ・`manifest.json` の内容・CLI の実行モデルは仕様（`requirement`）の
領分（→ `docs/requirements/README.md`「lib API」「entries スキーマ仕様」「manifest.json スキーマ」「CLI 仕様」）。

---

## モジュール統合設計

全モジュールと devShell は root と activation タイミングを供給してエンジンを起動する薄い配線。
起動の差は link-farm の取得方法だけに閉じる（→ [モジュールは root と起動の時機だけを供給して engine を起動する](../adr/modules-wire-engine-kick-only.md), [apply --manifest でビルド済み link-farm を engine へ直接渡す](../adr/apply-manifest-prebuilt-link-farm.md)）。

各モジュールが公開するオプションと動作は仕様の領分
（→ `docs/requirements/README.md`「モジュールオプション仕様・モジュール別動作仕様」）。

---

## 使用パターン

→ `docs/README.md`「想定する使われ方」（`use_case` 7 件）。個々の設定の書き方はテンプレート
（`layat init`）と `templates/` の実物を参照する。

---

## テスト戦略

lib は Nix 評価テスト、エンジンは実 FS の tmpdir 統合テスト、非 NixOS で動く主張は実 nix の
E2E で検証する（→ [テストを sandbox 可否で flake check と E2E に分ける](../adr/test-strategy-layers.md)）。

CI での実行基盤・キャッシュ・リリースは `infrastructure` item の領分。

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
- `docs/README.md` — コンセプト（`solution` / `use_case` の全体像）
- `docs/requirements/README.md` — 仕様の全体像（`requirement`）
- `docs/adr/` — 意思決定の記録
- `docs/model.yaml` — sara の型定義（item の型・関係・ID 形式）
