# cli-json

## テスト条件（`test_condition`）

* [前段の衝突の確認がバケット規則と選択集合に従って衝突を判定する](./apply-all-conflict-bucket-rules.md) - `apply --all` の前段の確認が、同じ root に解決する config 同士だけを比べ、衝突したら何も build せずに止まることを確かめる。
* [並列 build の失敗が config ごとに集約され並列度が --jobs を超えない](./apply-all-parallel-build-failure-and-bound.md) - `apply --all` の stage 1（先行並列 build）が、並列度を `--jobs` で抑えたまま、失敗を取りこぼさずに stage 2 の集約へ渡すことを確かめる。
* [配置並列の集約が完了順に依らず辞書順で、カウントが並行安全である](./apply-all-parallel-place-lexical-aggregate.md) - `apply --all` の stage 2（配置の並列実行）が、config の完了順に依らず決定的に集約することを確かめる。
* [破壊的操作の確認が対話性と同意フラグの組み合わせで決まり非対話で止まらない](./confirm-policy-interactivity.md) - 確認プロンプトを出すか、出さずに進むか、出さずに拒否するかが、対話性・同意フラグ・機械可読出力の指定の組み合わせで決まることを確かめる。
* [engine 結果から item と change への写像がインベントリ全数と実差分に一致する](./engine-result-inventory-mapping.md) - engine の実行結果を JSON ペイロードへ写す規則が、実際に起きたことと一致することを確かめる。形ではなく内容を見る条件であり、適合チェッカの外側にある。
* [エンベロープが全形状で outturn 適合を保ち subject 件数で形が揺れない](./envelope-shape-conformance.md) - emit した文書を outturn の適合チェッカ（正とする定義として embed した schema と、schema 外の lint）へ通し、出現しうる全形状で適合することを確かめる。形状とは、subject の有無、成功と失敗、主体確立前の失敗と主体起因の失敗、conflict、および `--all` の subject 0 件 / 1 件 / 複数件である。
* [失敗が item と subject とトップレベルの正しい層に 1 度だけ載り原因で分類される](./error-layer-placement.md) - 失敗の帰属層が原因の性質で決まり、かつ 1 つの失敗が 1 つの層にしか載らないことを確かめる。
* [layat: を含む文字列リテラルが stderr へ書く出力のファイルの外に無い](./error-prefix-only-at-output-boundary.md) - internal/ と cmd/layat/ の非テストの Go で、layat: を含む文字列リテラルが stderr へ書く出力のファイルだけにあることを確かめる
* [終了コードの集約が優先順位に従い skip を失敗として数えない](./exit-code-aggregation.md) - 複数の対象を適用したときの終了コードが、個々の結果からの決めごとに従って決まることを確かめる。
* [フラグの解釈と対象選択が仕様どおりで位置引数を飲み込まない](./flag-parsing-target-selection.md) - CLI の表層が「どの対象に」「どの修飾で」作用するかを決める部分を確かめる。
* [生成器の各実装が実 nix 無しで同じ契約（操作・失敗の構造）を満たす](./generator-contract-implementations.md) - 実 nix を起動せず、契約の実装（prebuilt・テストダブル・nix 生成器）と失敗型を確かめる。nix生成器は PATH の先頭に置いた nix のスタブで駆動する。
* [生成器の診断が成功・失敗とも writer へ届き、失敗は分類と案内付きで出力面ごとに提示される](./generator-diagnostics-presentation.md) - nix 生成器は PATH の先頭に置いた nix のスタブで駆動し、CLI の提示は `generator.Error` を直接与えて確かめる。実 nix での到達は e2e で確かめる。
* [生成器の選択が優先順位どおりに解決し、入力不正を E_INPUT で拒否し、--manifest の処理では環境を読まない](./generator-selection-precedence-and-input-rejection.md)
* [--help の開示ブロックがサブコマンドごとに実際の nix 呼び出しと一致する](./help-disclosure-matches-nixgen.md) - `layat --help` の開示ブロックが、各サブコマンドを実際に実行したときの nix 呼び出しと一致することを検証する。
* [info スロットと世代スロットが値の無いときキーごと不在になり空配列は保つ](./info-seat-omission.md) - 省略されるべきフィールドが、ゼロ値として出力に現れないことを確かめる。この条件が独立して必要なのは、`{}` や `null` や `""` を含む文書も schema 上は合法であり、適合チェッカでは検出されないため。ここはテストでしか守れない層である。
* [entrypoint の探索順序と nix へ渡す引数列が形ごとに正しく組み立てられる](./nix-argv-assembly.md) - nix を実際に起動せず、起動する直前の引数列までを確かめる。探索と組み立てを純粋な関数として切り出し、実 FS 上に空ファイルを置いた fixture に対して評価する。
* [apply --all が `--manifest <name>=<link-farm>` の集合に乗り、集合の形の不正を E_INPUT で拒否する](./prebuilt-set-rides-apply-all.md)
* [status の CLI がフィルタ・ストリームの排他・--json の形・error 時の続行の契約を守る](./status-cli-contracts.md) - `layat status [name]` の CLI 層が、engine の結果を契約どおりに絞り込み、名付け、出力することを検証する。
* [stdout を機械可読出力が専有し人間向けの出力が stderr に留まる](./stream-ownership.md) - どの出力がどのストリームへ出るかを、既定の契約と `--json` の契約の双方で確かめる。
* [warning が対象の有無で item と subject へ振り分けられ全種別がコードへ写る](./warning-routing.md) - warning の置き場所が、その warning の指す対象がインベントリに存在するかどうかで決まることを確かめる。判定軸はインベントリへの帰属だけであり、警告の種別ではない。インベントリに載っている対象への警告はその item が持ち、インベントリの外にある対象への警告は subject が持つ。後者は item が無い以上どの item にも紐づけられないため、警告自体が対象を自己完結して示す必要がある。

## テストケース（`test_case`）

* [apply_all_build_test.go — 並列 build の並列度上限・失敗集約・--jobs の値域・--debug prefix・集約の同一性](./apply-all-build-test.md) - `cmd/layat/apply_all_build_test.go`
* [apply_all_conflict_test.go — 前段の衝突の確認のバケット判定・選択集合・build 前停止・単一 apply では確認しないこと・--manifest 集合](./apply-all-conflict-test.md) - `cmd/layat/apply_all_conflict_test.go`
* [apply_all_failure_matrix_test.go — build 失敗 × 配置失敗 × 成功 × skip（× conflict）の部分失敗マトリクス](./apply-all-failure-matrix-test.md) - `cmd/layat/apply_all_failure_matrix_test.go`
* [apply_all_manifest_project_test.go — project と home の link-farm を apply --all --manifest で配置し root フィルタで選別する](./apply-all-manifest-project-test.md) - `cmd/layat/apply_all_manifest_project_test.go`（`runApplyAll` を `--manifest` の集合で実 engine まで駆動する。link-farm は rootKind を指定できる `writeRootKindLinkFarm` で書く）
* [apply_all_parallel_test.go — 完了順逆転時の辞書順集約・部分失敗と skip・カウントの並行安全性](./apply-all-parallel-test.md) - `cmd/layat/apply_all_parallel_test.go`
* [apply_test.go — 終了コード集約・root フィルタ・退避行の出力](./apply-test.md) - `cmd/layat/apply_test.go`
* [backup_flag_test.go — 値を省略できるフラグが位置引数を飲み込まない](./backup-flag-test.md) - `cmd/layat/backup_flag_test.go`
* [generator_error_test.go — 生成器の失敗の人間向け / --json / --debug の 3 形](./generator-error-test.md) - `cmd/layat/generator_error_test.go`（`printGeneratorError` と `generatorErrorMessage`）
* [generator_select_test.go — 選択の優先順位・strict な設定ファイル・--manifest の処理の環境非依存](./generator-select-test.md) - `cmd/layat/generator_select_test.go`
* [generator_test.go — 契約の失敗型・prebuilt 実装・テストダブル](./generator-test.md) - `internal/generator/generator_test.go`
* [gitignore_test.go — パス正規化と列挙結果の info 表現](./gitignore-test.md) - `cmd/layat/gitignore_test.go`
* [help_disclosure_test.go — --help の開示ブロックと nix 呼び出しの突合・nix を実行しないサブコマンド](./help-disclosure-test.md) - `cmd/layat/help_disclosure_test.go`
* [init_test.go — テンプレート名の受理集合と subject を持たない文書形](./init-test.md) - `cmd/layat/init_test.go`
* [list-generations_test.go — 世代一覧の info 表現と二重エンコードの禁止](./list-generations-test.md) - `cmd/layat/list-generations_test.go`
* [nixgen_test.go — entrypoint 探索順序と installable 引数の 2 形式](./nix-test.md) - `internal/generator/nixgen/nixgen_test.go`（nix 生成器。旧 `cmd/layat/nix_test.go` から移設）
* [outturn_all_test.go — --all の subject 件数不変性と部分失敗の非汚染](./outturn-all-test.md) - `cmd/layat/outturn_all_test.go`
* [outturn_payload_test.go — engine 結果からペイロードへの写像](./outturn-payload-test.md) - `cmd/layat/outturn_payload_test.go`
* [outturn_test.go — エンベロープ基礎形・id 導出・エラー分類・stdout 所有権](./outturn-test.md) - `cmd/layat/outturn_test.go`
* [prune_test.go — prune の確認ポリシー・root パス一覧の提示・ストリーム排他と --json の形状](./prune-cli-test-go.md) - `cmd/layat/prune_test.go`
* [reset_test.go — 確認ポリシーと出力ストリームの排他](./reset-test.md) - `cmd/layat/reset_test.go`
* [status_test.go — status の subject 名・フィルタ・一致 0 件・ストリームの排他・--json の形・error 時の続行](./status-cli-test-go.md) - `cmd/layat/status_test.go`
* [stderr_test.go — stderr の行の形・errors.Join の分け方・失敗の詳細の出口・layat: を含むリテラルの所在](./stderr-test.md) - cmd/layat/stderr_test.go が、stderr へ書く行の形と、layat: を含む文字列リテラルが出力のファイルの外に無いことを確かめる
* [version_test.go — バージョンの供給元一致・--version の出力書式](./version-test.md) - `cmd/layat/version_test.go`
