# manifest-eval

## テスト条件（`test_condition`）

* [アグリゲータの名前衝突の確認が衝突を落とし、衝突元を全件示す](./aggregator-collision-detection.md) - nix-unit のスイートを組み立てるマージ（`tests/nix-unit-lib.nix` の `mergeTests`）が、TP-36e90d5d-4524-4294-bc72-ee263bb02782 の定めるテスト名の一意性を機械的に強制していることをアサートする。合成した`{ file, tests }` のリストを渡し、readDir が返す実ファイル構成に依存せず throw 側の処理を実行する。
* [GC アンカー名の形式・決定性・特殊文字耐性をアサートする](./anchor-name-determinism.md) - アンカー名（target の sha256 先頭 32 hex）そのものの性質を、farm の組み立てから切り離してアサートする。抽出条件と組み立て結果は TC-1d69350e-db3c-4d74-a24e-7a3fabb31b0a の担当。
* [省略時の既定適用・明示上書き・target 辞書順での決定的配列化をアサートする](./entry-normalization-defaults.md) - attrset で宣言された entries が配列へ正規化される過程を、3 つの観点でアサートする。
* [GC アンカー対象の抽出条件と farm の組み立て結果をアサートする](./farm-anchor-derivation.md) - symlink farm が GC アンカー専用であることを、対象の抽出と組み立て結果で見る。アンカー名そのものの形式・決定性は TC-a6a14739-89e5-4eba-acee-3ed5be5a0b7e の担当で、ここでは扱わない。
* [fixed root で rootKind が fixed になり絶対パスを併記する（home marker が倒れないことを含む）](./fixed-root-absolute-path.md) - `normalizeManifest` の `root` に marker でない絶対パス文字列を渡したとき、出力の `root`オブジェクトが REQ-dd10d820-e453-4099-a47a-ffb9a7de02fb の fixed 側の規定どおりになることをアサートする。
* [manifest 文書全体をスナップショットに固定し、どの不変条件も名指ししない変更を差分として出す](./manifest-document-regression.md) - `normalizeManifest` の出力全体をスナップショットとして保存し、以後の変更が差分として現れるようにする。個別アサートは名指しした性質しか守らないため、フィールドの追加・削除・改名のように「アサートを書き忘れた箇所」の変化はここでしか捕まらない。
* [manifest 文書の形（schemaVersion / root / entry 5 フィールド）を不変条件としてアサートする](./manifest-shape-invariants.md) - `normalizeManifest` が返す文書のトップレベルと entry の形を、値ごとに名前の付いた不変条件としてアサートする。
* [marker の判別と src 種別への変換を単一 entry の境界でアサートする](./marker-resolution.md) - manifest 全体を組み立てずに、単一 entry の解決（`resolveEntry`）と型側の marker 判別述語を直接突いてアサートする。文書の形（TC-4e7cfae7-72bc-4af6-a1f5-1ead7db564b1）とは別の境界で、変換ロジックそのものを見る。
* [パス脱出判定の深さ境界を private helper 越しに内側・外側の対でアサートする](./path-escape-boundary.md) - `escapesBase` / `isUnsafe` をテスト seam 越しに直接叩き、深さ 0 の境界を内側と外側の対でアサートする。エラーの到達性は TC-e7ff0e6d-32d7-4ed6-8c2f-449dba34b2f6 の担当。
* [拒否すべき入力が評価時に throw することをアサートする](./rejection-gates.md) - 「この入力は評価が失敗する」を、エラー種別と拒否されたフィールドの名指しで直接アサートする。正常系のアサートでは、拒否条件が外れたことを検出できないため。判定ロジックそのものを private helper 越しに単体で見るのは TC-311ca3b2-d6d1-4712-9e2a-9027941d3527 の担当で、検証境界が異なる。
* [subdir 適用形が全 target に正規化済みの前置を施し、root と単体 marker を変えず、不正な subdir を評価時に拒否する](./root-subdir-sugar.md) - root マーカーの subdir 適用形（REQ-08c8f639-790c-4970-a954-c9567538cf9b）が、全 target への前置と等価な糖衣として振る舞うことを確かめる。

## テストケース（`test_case`）

* [namaka: manifest-project — manifest 文書全体のスナップショット回帰](./namaka-manifest-project.md) - `tests/namaka/manifest-project/expr.nix`（スナップショット実体は`tests/namaka/_snapshots/manifest-project`）
* [namaka: manifest-subdir — subdir 適用形で生成した manifest 文書のスナップショット回帰](./namaka-manifest-subdir.md) - `tests/namaka/manifest-subdir/expr.nix`（スナップショット実体は`tests/namaka/_snapshots/manifest-subdir`）
* [nix-unit: aggregator-merge.nix — 衝突の確認をダミー入力で検証する](./nix-unit-aggregator-merge.md) - `tests/nix-unit/aggregator-merge.nix`（`tests/nix-unit.nix` のディレクトリ列挙で自動搭載。TP-36e90d5d-4524-4294-bc72-ee263bb02782 が定めるとおりテスト名はファイル横断で一意でなければならず、本ファイルは`testAggregatorMerge*` 接頭辞を使う）
* [nix-unit: anchor-name.nix — GC アンカー名の形式と決定性](./nix-unit-anchor-name.md) - `tests/nix-unit/anchor-name.nix`（TP-403c55c7-d996-4951-8e6b-c3a7dddd387c のテスト seam `layat.__internal.anchorName` を直接叩く）
* [nix-unit: apply-command.nix — mkApplyCommand の文字列と root の確認](./nix-unit-apply-command.md) - `tests/nix-unit/apply-command.nix`（`layat.mkApplyCommand` を fake の pkgs・layatPackage・farm で純評価する。layatPackage の fake は `lib.getExe` の derivation の確認を通すため `type = "derivation"` を持つ）
* [nix-unit: defaults.nix — 既定適用・明示上書き・target 辞書順](./nix-unit-defaults.md) - `tests/nix-unit/defaults.nix`
* [nix-unit: escapes-base.nix — `..` 深さ判定の境界網羅](./nix-unit-escapes-base.md) - `tests/nix-unit/escapes-base.nix`（TP-403c55c7-d996-4951-8e6b-c3a7dddd387c のテスト seam `layat.__internal` 経由で private helper `escapesBase` / `pathChecks.isUnsafe` を直接叩く）
* [nix-unit: farm-entries.nix — アンカー対象の抽出と anchorLines の組み立て](./nix-unit-farm-entries.md) - `tests/nix-unit/farm-entries.nix`（TP-403c55c7-d996-4951-8e6b-c3a7dddd387c のテスト seam `layat.__internal.farmEntries` / `layat.__internal.anchorName` / `layat.__internal.anchorLines` を直接叩き、加えて `mkManifest` のビルドスクリプトへの配線を fake pkgs 経由で見る。anchorLines は `lib/manifest.nix` が`mkManifest` から呼ぶ生成式そのもので、テスト側に複製は持たない → Issue #289）
* [nix-unit: fixed-root.nix — fixed root の絶対パス併記](./nix-unit-fixed-root.md) - `tests/nix-unit/fixed-root.nix`（`tests/nix-unit.nix` のディレクトリ列挙で自動搭載。TP-36e90d5d-4524-4294-bc72-ee263bb02782が定めるとおり、テスト名はファイル横断で一意でなければならない。本ファイルは`testFixedRoot*` 接頭辞を使う）
* [nix-unit: gates.nix — 評価時に throw する拒否条件群](./nix-unit-gates.md) - `tests/nix-unit/gates.nix`
* [nix-unit: resolve-marker.nix — src 解決と marker 判別述語](./nix-unit-resolve-marker.md) - `tests/nix-unit/resolve-marker.nix`（TP-403c55c7-d996-4951-8e6b-c3a7dddd387c のテスト seam `layat.__internal.resolveEntry`と `lib/types.nix` の `isRootMarker` / `isOutOfStoreMarker` / `isRootRelativeMarker` を直接叩く。manifest 全体を介さず単一 entry の src 種別判定・文字列化だけを見る最小の境界。root 相対 marker だけは、`normalizeManifest` を通した entries の形も 1 件確かめる）
* [nix-unit: root-subdir.nix — root マーカーの subdir 適用形](./nix-unit-root-subdir.md) - `tests/nix-unit/root-subdir.nix`（`tests/nix-unit.nix` のディレクトリ列挙で自動搭載。TP-36e90d5d-4524-4294-bc72-ee263bb02782が定めるとおり、テスト名はファイル横断で一意でなければならない。本ファイルは`testRootSubdir*` 接頭辞を使う）
* [nix-unit: structure.nix — manifest 構造の不変条件](./nix-unit-structure.md) - `tests/nix-unit/structure.nix`（`tests/nix-unit.nix` のディレクトリ列挙で自動搭載。TP-36e90d5d-4524-4294-bc72-ee263bb02782が定めるとおり、テスト名はファイル横断で一意でなければならない）
