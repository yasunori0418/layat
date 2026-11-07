# cli-json

## テスト条件（`test_condition`）

* [フラグの解釈と対象選択が仕様どおりで位置引数を飲み込まない](./20260809-0c7281f8-cffc-4a53-8fe0-f52d9ea6bd12-flag-parsing-target-selection.md) - CLI の表層が「どの対象に」「どの修飾で」作用するかを決める部分を確かめる。
* [エンベロープが全形状で outturn 適合を保ち subject 件数で形が揺れない](./20260809-1a46da17-1812-431b-966f-ec8dffd315ae-envelope-shape-conformance.md) - emit した文書を outturn の適合チェッカ（正とする定義として embed した schema と、schema 外の lint）へ通し、出現しうる全形状で適合することを確かめる。形状とは、subject の有無、成功と失敗、主体確立前の失敗と主体起因の失敗、conflict、および `--all` の subject 0 件 / 1 件 / 複数件である。
* [info スロットと世代スロットが値の無いときキーごと不在になり空配列は保つ](./20260809-4e0a14d6-342c-448b-964b-b8e87520e89b-info-seat-omission.md) - 省略されるべきフィールドが、ゼロ値として出力に現れないことを確かめる。この条件が独立して必要なのは、`{}` や `null` や `""` を含む文書も schema 上は合法であり、適合チェッカでは検出されないため。ここはテストでしか守れない層である。
* [entrypoint の探索順序と nix へ渡す引数列が形ごとに正しく組み立てられる](./20260809-69537269-32d7-4e04-8722-806b58e7db7a-nix-argv-assembly.md) - nix を実際に起動せず、起動する直前の引数列までを確かめる。探索と組み立てを純粋な関数として切り出し、実 FS 上に空ファイルを置いた fixture に対して評価する。
* [warning が対象の有無で item と subject へ振り分けられ全種別がコードへ写る](./20260809-71fdbbfb-cf81-4aba-8254-59d7cb45d5ab-warning-routing.md) - warning の置き場所が、その warning の指す対象がインベントリに存在するかどうかで決まることを確かめる。判定軸はインベントリへの帰属だけであり、警告の種別ではない。インベントリに載っている対象への警告はその item が持ち、インベントリの外にある対象への警告は subject が持つ。後者は item が無い以上どの item にも紐づけられないため、警告自体が対象を自己完結して示す必要がある。
* [失敗が item と subject とトップレベルの正しい層に 1 度だけ載り原因で分類される](./20260809-733ac4ed-dac3-4bb8-9bfd-fd1cbdc300c9-error-layer-placement.md) - 失敗の帰属層が原因の性質で決まり、かつ 1 つの失敗が 1 つの層にしか載らないことを確かめる。
* [破壊的操作の確認が対話性と同意フラグの組み合わせで決まり非対話で止まらない](./20260809-b0e0cdb8-0c56-4eb8-8d4e-86569af85dec-confirm-policy-interactivity.md) - 確認プロンプトを出すか、出さずに進むか、出さずに拒否するかが、対話性・同意フラグ・機械可読出力の指定の組み合わせで決まることを確かめる。
* [stdout を機械可読出力が専有し人間向けの出力が stderr に留まる](./20260809-b4a365cd-0b1c-40ff-92c6-91f6cffe8e98-stream-ownership.md) - どの出力がどのストリームへ出るかを、既定の契約と `--json` の契約の双方で確かめる。
* [engine 結果から item と change への写像がインベントリ全数と実差分に一致する](./20260809-cf8189c4-d680-4d7e-bcd3-810543762c50-engine-result-inventory-mapping.md) - engine の実行結果を JSON ペイロードへ写す規則が、実際に起きたことと一致することを確かめる。形ではなく内容を見る条件であり、適合チェッカの外側にある。
* [終了コードの集約が優先順位に従い skip を失敗として数えない](./20260809-ddee6cc4-bc10-4107-bc7e-288a5fb62f1f-exit-code-aggregation.md) - 複数の対象を適用したときの終了コードが、個々の結果からの決めごとに従って決まることを確かめる。
* [前段の衝突の確認がバケット規則と選択集合に従って衝突を判定する](./20260926-1de3feef-3d6e-4310-adb3-ef659f4175bc-apply-all-conflict-bucket-rules.md) - `apply --all` の前段の確認が、同じ root に解決する config 同士だけを比べ、衝突したら何も build せずに止まることを確かめる。
* [配置並列の集約が完了順に依らず辞書順で、カウントが並行安全である](./20260926-9ccb2ffe-7bf6-41ea-a9ce-df06cdda0509-apply-all-parallel-place-lexical-aggregate.md) - `apply --all` の stage 2（配置の並列実行）が、config の完了順に依らず決定的に集約することを確かめる。
* [並列 build の失敗が config ごとに集約され並列度が --jobs を超えない](./20260926-e42a5438-039b-41b3-9a39-0e8dc1f25cb3-apply-all-parallel-build-failure-and-bound.md) - `apply --all` の stage 1（先行並列 build）が、並列度を `--jobs` で抑えたまま、失敗を取りこぼさずに stage 2 の集約へ渡すことを確かめる。
* [生成器の診断が成功・失敗とも writer へ届き、失敗は分類と案内付きで出力面ごとに提示される](./20261002-4a39ca86-e866-41ec-93d2-663f8e6c1ce0-generator-diagnostics-presentation.md) - nix 生成器は PATH の先頭に置いた nix のスタブで駆動し、CLI の提示は `generator.Error` を直接与えて確かめる。実 nix での到達は e2e で確かめる。
* [生成器の選択が優先順位どおりに解決し、入力不正を E_INPUT で拒否し、--manifest の処理では環境を読まない](./20261002-8377d290-1f47-48f3-844c-bd51f257e25e-generator-selection-precedence-and-input-rejection.md)
* [生成器の各実装が実 nix 無しで同じ契約（操作・失敗の構造）を満たす](./20261002-b6d6cf76-16f9-4c9f-9b50-74104254ad6e-generator-contract-implementations.md) - 実 nix を起動せず、契約の実装（prebuilt・テストダブル・nix 生成器）と失敗型を確かめる。nix生成器は PATH の先頭に置いた nix のスタブで駆動する。
* [apply --all が --manifest <name>=<link-farm> の集合に乗り、集合の形の不正を E_INPUT で拒否する](./20261015-e54fe715-7b49-431c-a611-25e196b1d0a0-prebuilt-set-rides-apply-all.md)
* [status の CLI がフィルタ・ストリームの排他・--json の形・error 時の続行の契約を守る](./20261023-1bc3e39a-116a-4f26-96b4-1d15e6f837bd-status-cli-contracts.md) - `layat status [name]` の CLI 層が、engine の結果を契約どおりに絞り込み、名付け、出力することを検証する。
* [--help の開示ブロックがサブコマンドごとに実際の nix 呼び出しと一致する](./20261101-b4262641-776c-427d-8331-1a94b29fa568-help-disclosure-matches-nixgen.md) - `layat --help` の開示ブロックが、各サブコマンドを実際に実行したときの nix 呼び出しと一致することを検証する。

## テストケース（`test_case`）

* [reset_test.go — 確認ポリシーと出力ストリームの排他](./20260809-020fa59e-9aaa-498b-ad5f-dd49b8b48f78-reset-test.md) - `cmd/layat/reset_test.go`
* [version_test.go — バージョンの供給元一致・--version の出力書式](./20260809-3b8d14a5-192b-46ba-94bd-ce1e78c4e8de-version-test.md) - `cmd/layat/version_test.go`
* [init_test.go — テンプレート名の受理集合と subject を持たない文書形](./20260809-5ad75041-e529-4f01-a8a6-8f15e40be33e-init-test.md) - `cmd/layat/init_test.go`
* [apply_test.go — 終了コード集約・root フィルタ・退避行の出力](./20260809-81d876af-a1e3-4348-a827-81c8ba7ef39b-apply-test.md) - `cmd/layat/apply_test.go`
* [outturn_all_test.go — --all の subject 件数不変性と部分失敗の非汚染](./20260809-8b3feeb2-7b7b-482c-bba2-6db0b923647a-outturn-all-test.md) - `cmd/layat/outturn_all_test.go`
* [outturn_test.go — エンベロープ基礎形・id 導出・エラー分類・stdout 所有権](./20260809-932e218b-5bfb-4d66-98a3-41054c0d863f-outturn-test.md) - `cmd/layat/outturn_test.go`
* [outturn_payload_test.go — engine 結果からペイロードへの写像](./20260809-a9ff732c-af30-4f13-862d-92e3074d390d-outturn-payload-test.md) - `cmd/layat/outturn_payload_test.go`
* [gitignore_test.go — パス正規化と列挙結果の info 表現](./20260809-c5c6d127-c227-4f01-9a5d-fc09e7412175-gitignore-test.md) - `cmd/layat/gitignore_test.go`
* [backup_flag_test.go — 値を省略できるフラグが位置引数を飲み込まない](./20260809-cd3e3428-3fe0-4d33-a64c-1931ac11d07d-backup-flag-test.md) - `cmd/layat/backup_flag_test.go`
* [list-generations_test.go — 世代一覧の info 表現と二重エンコードの禁止](./20260809-d0505d58-ef48-4fe2-ab7c-77002109d1c4-list-generations-test.md) - `cmd/layat/list-generations_test.go`
* [nixgen_test.go — entrypoint 探索順序と installable 引数の 2 形式](./20260809-eec02695-ac60-4ea9-a825-f23f7669d132-nix-test.md) - `internal/generator/nixgen/nixgen_test.go`（nix 生成器。旧 `cmd/layat/nix_test.go` から移設）
* [prune_test.go — prune の確認ポリシー・root パス一覧の提示・ストリーム排他と --json の形状](./20260823-be10d33c-3715-41a3-8b98-951e95cb5e3f-prune-cli-test-go.md) - `cmd/layat/prune_test.go`
* [apply_all_conflict_test.go — 前段の衝突の確認のバケット判定・選択集合・build 前停止・単一 apply では確認しないこと・--manifest 集合](./20260926-16289504-60e7-4046-9deb-7e2c962c5abf-apply-all-conflict-test.md) - `cmd/layat/apply_all_conflict_test.go`
* [apply_all_failure_matrix_test.go — build 失敗 × 配置失敗 × 成功 × skip（× conflict）の部分失敗マトリクス](./20260926-4a8860b1-5a88-4225-9702-f946be9c984e-apply-all-failure-matrix-test.md) - `cmd/layat/apply_all_failure_matrix_test.go`
* [apply_all_build_test.go — 並列 build の並列度上限・失敗集約・--jobs の値域・--debug prefix・集約の同一性](./20260926-9d21aef7-487d-4193-ad79-131c569db1e8-apply-all-build-test.md) - `cmd/layat/apply_all_build_test.go`
* [apply_all_parallel_test.go — 完了順逆転時の辞書順集約・部分失敗と skip・カウントの並行安全性](./20260926-a09f5137-2386-4bea-bcf9-a82504a988e6-apply-all-parallel-test.md) - `cmd/layat/apply_all_parallel_test.go`
* [generator_test.go — 契約の失敗型・prebuilt 実装・テストダブル](./20261002-1d6ff920-79d1-4bf7-bcee-48b5c06b0134-generator-test.md) - `internal/generator/generator_test.go`
* [generator_select_test.go — 選択の優先順位・strict な設定ファイル・--manifest の処理の環境非依存](./20261002-28e6be05-f3d5-46c9-84b1-7ddfa07d253b-generator-select-test.md) - `cmd/layat/generator_select_test.go`
* [generator_error_test.go — 生成器の失敗の人間向け / --json / --debug の 3 形](./20261002-83bd9936-cfb5-4e1d-a9a7-b85f049991b8-generator-error-test.md) - `cmd/layat/generator_error_test.go`（`printGeneratorError` と `generatorErrorMessage`）
* [status_test.go — status の subject 名・フィルタ・一致 0 件・ストリームの排他・--json の形・error 時の続行](./20261023-dc611f8f-4ee4-44cb-8b34-01d68da1b918-status-cli-test-go.md) - `cmd/layat/status_test.go`
* [apply_all_manifest_project_test.go — project と home の link-farm を apply --all --manifest で配置し root フィルタで選別する](./20261025-c4ffa8b1-4058-4cba-b911-520ee970b2ae-apply-all-manifest-project-test.md) - `cmd/layat/apply_all_manifest_project_test.go`（`runApplyAll` を `--manifest` の集合で実 engine まで駆動する。link-farm は rootKind を指定できる `writeRootKindLinkFarm` で書く）
* [help_disclosure_test.go — --help の開示ブロックと nix 呼び出しの突合・nix を実行しないサブコマンド](./20261101-ad7299c1-e8af-4298-88a4-e1c61062ba30-help-disclosure-test.md) - `cmd/layat/help_disclosure_test.go`
