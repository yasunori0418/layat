# engine-core

## テスト条件（`test_condition`）

* [実体が占有する target と祖先 symlink を上書きせず停止すること](./20260809-405606f0-3ac8-4cc2-998e-d4759a62a171-no-overwrite-stop.md) - target に通常ファイル・実ディレクトリが在るとき、および target の祖先コンポーネントがsymlink であるとき、engine が配置を行わず conflict として停止することを検証する。上書きは回復不能な破壊になるため、「停止する」ことに加えて「配置物を一切残さない」ことまでを条件に含む。
* [conflict が全件・種類に応じたガイダンス付きで列挙され、集約エラー 1 本で返ること](./20260809-4da40ee8-877e-405b-98ca-6bf5de926ba4-conflict-full-report.md) - conflict が複数あるとき、最初の 1 件で打ち切らず全件を列挙し、件数を含む集約エラー 1 本を返すことを検証する。ユーザーが 1 回の apply で対処すべき全てを把握できることがこの条件の狙いであり、「列挙されること」と「種類ごとに適切な対処ガイダンスが付くこと」の両方を含む。
* [out-of-store の配置前の確認がリンク先不在と確認の失敗を区別すること](./20260809-8435052a-5dcc-49e2-ac26-82f645cb6890-out-of-store-preflight.md) - out-of-store entry の配置直前に marker のリンク先の存在を確かめ、不在のときは dangling symlink を作らずに engine 実行時エラーで停止することを検証する。
* [target の現況に応じた配置分類（新規 / 自己記録の張替え / foreign 後勝ち）が仕様どおりであること](./20260809-9df804ce-35ee-44a7-87b1-17935d53fab2-placement-kinds.md) - target が「不在」「前世代 manifest に記録され記録先を指す symlink」「記録に無い symlink（foreign）」のいずれであるかに応じて、新規配置 / 無警告の張替え / 警告付き後勝ちへ振り分けられることを検証する。分類は planner の純粋な計算と engine の実 FS 実行の両方で同じ結論に至る必要があるため、実 FS 統合テストと fake FS の table-driven テストの双方から覆う。
* [dryrun が副作用を持たず、本番 apply と同数の conflict を報告すること](./20260809-a5eb7de3-a1a7-41ae-8fb9-c2aa374ac894-dryrun-readonly-parity.md) - `apply --dryrun` が FS へ一切の変更を残さないこと（配置物・profile ディレクトリのいずれも生成しないこと）を検証する。副作用は成功時に表面化しないため、実行前後の FS 状態を比較する形で条件を立てる。
* [root 解決と profile ディレクトリ準備の成功時・失敗時の処理が仕様どおりであること](./20260809-b254a5a8-7fbf-4f31-8486-3e373d66bfa7-root-resolution.md) - project mode の root が git toplevel から解決されること、git repo の外で実行したときにengine 実行時に停止すること、および root 解決に失敗したときに配置へ進まないことを検証する。`fixed` で root パスが無いときに engine が拒否することも同じ条件の射程に含む。成功時の処理も同じ条件の下に置く。`rootKind=home` が `$HOME` を絶対 root として返すこと、`--root` 上書きが rootKind によらず優先されること、相対で与えた override が cwd 起点で絶対化されて返ることを見る。
* [plan の計算が実 FS を触らない純粋な関数として fake FS で網羅検証できること](./20260809-b329cafd-06c3-4b87-8357-229b69e5ba5c-plan-purity.md) - plan の計算が (前世代 manifest, 新 manifest, root, FS プローブ) のみを入力とし、fake FS を差し込むだけで分類の全パターンを table-driven に列挙できることを検証する。実 FS へ触る実装が含まれればこの条件は成立しなくなるため、fake FS だけでテストが完結すること自体が検証対象である。
* [system mode の root 解決・基底切替・euid の条件が仕様どおりであること](./20261018-be3d3dc9-1ead-4e2e-8eea-46e1c8884562-system-mode-root-gate.md) - `rootKind=system` が root `/` に解決され、`--root` 上書きが優先されることを検証する。

## テストケース（`test_case`）

* [internal/engine/dryrun_test.go — dryrun の無副作用と除去なしの移行判断](./20260809-0c2cec08-db65-427c-9480-ac0dcacc8d31-dryrun-test-go.md) - `internal/engine/dryrun_test.go`
* [internal/engine/engine_test.go — 実 tmpdir に対する apply の統合テスト](./20260809-31fdb776-9650-4ff9-97f6-45d56b2e7177-engine-test-go.md) - `internal/engine/engine_test.go`
* [internal/planner/planner_test.go — fake FS による plan 計算の table-driven 網羅](./20260809-efe9e9c8-cc0a-4b1b-bf2b-aa1d15428252-planner-test-go.md) - `internal/planner/planner_test.go`
* [internal/engine/preflight_test.go — out-of-store の配置前の確認での非 ENOENT エラー](./20260809-f3a5fee2-1981-42c3-ba5b-81b75fd7a3ad-preflight-test-go.md) - `internal/engine/preflight_test.go`
