# layat

フェッチ済みの git リポジトリをユーザー環境の任意パスへ symlink / copy で配置する Nix ライブラリ・モジュール群。設定生成は行わない。

## ドキュメント

`docs/` は **README → 概要文書 → item の 3 層構造**。規範的な内容は item（1 ファイル 1 主張の
Markdown + YAML frontmatter）が持ち、概要文書は通読の入口として全体像と item への索引を担う。

### 概要文書（通読の入口）

- `docs/concept.md` — コンセプト（solution / use_case への索引）+ 設計の哲学・既存ツールとの比較・north-star・設計の変遷
- `docs/design.md` — 設計（design item への索引）
- `docs/spec.md` — 仕様（requirement item への索引）

**これらは概要に縮退済み**。各文書の冒頭に書き方の規約（セクションあたりの散文の行数上限・
詳細は item へリンクし本文に書き下さない）があり、それに従う。仕様・設計の内容を足すときは
item を新設し、概要文書にはリンクを 1 行足すに留める。

### item 群（規範的な内容の所在）

| ディレクトリ | 型 | prefix | 親 |
|---|---|---|---|
| `docs/solution/` | solution | `SOL` | なし（根）|
| `docs/use-cases/` | use_case | `UC` | solution |
| `docs/requirements/` | requirement | `REQ` | use_case |
| `docs/design/` | design | `DSG` | requirement / test_plan |
| `docs/quality/` | quality | `QA` | solution |
| `docs/test-plan/` | test_plan | `TP` | solution |
| `docs/infrastructure/` | infrastructure | `INF` | quality / design |
| `docs/risks/` | risk | `RISK` | requirement / design |
| `docs/test/<対象>/` | test_condition | `TC` | risk |
| `docs/test/<対象>/` | test_case | `CASE` | test_condition / test_plan |
| `docs/adr/` | adr | `ADR` | なし（分離型）|
| `docs/sara-adr/` | sara_adr | `SDR` | なし（分離型）|
| `docs/ubiquitous-language/` | ubiquitous_term | `UBQ` | なし（relation を持たない根）|

型・関係・フィールドの定義は `docs/model.yaml` が持つ。親を持たない根は **solution と
ADR の 2 型（adr / sara_adr）と ubiquitous_term のみ**で、他の型は upstream relation を 1 本以上張る
（ubiquitous_term は relation を 1 本も持たない。張り忘れは `sara check` の接続漏れの warning で出る）。
検出した欠陥（defect）は型として持たず、GitHub Issues（`bug` label）で管理する
（→ [欠陥はグラフに置かず GitHub Issues で管理する](docs/sara-adr/defects-in-github-issues.md)。運用規約は `docs/agents/issue-tracker.md` の「Defect issues」節）。

risk を `requirement` と `design` のどちらに張るかの使い分けは [risk はまず requirement に張り、設計を差し替えると消える懸念だけを design に張る](docs/quality/risk-attaches-to-requirement-by-default.md)。

### テスト系 item の粒度（`docs/risks/` と `docs/test/<対象>/`）

`docs/test/<対象>/` の `<対象>` は**機能単位の 8 区分**。requirement 単位は 130 超の REQ との
1:N が錯綜するため採らない（→ Issue #273 の決定事項）。

**区分の一覧と並び順の規範は `dev/tests/test-categories.tsv`**（機械可読・契約テストと対応表
生成が共用する SSOT → Issue #304）。下の表はそれに test_plan 列を添えた読み物で、区分を
増減するときはデータファイル・`docs/test/` のディレクトリ・この表の 3 つを揃える
（前 2 つのずれは `dev/tests/test-doc-map.sh` が検出する）。

個々のテスト資産がどの区分に属するかは**どの表でも持たない**。CASE ファイルの置き場所
（`docs/test/<区分>/`）から導く。パス prefix では決まらない（`internal/engine/*_test.go` は
5 区分・`tests/e2e/scenarios/*` は 4 区分へ割れる）ため、prefix 表を持つと二重管理になる。

右端は各区分の CASE の上に立つ test_plan（逆算の起点。複数あるものは担当が分属する）。

| `<対象>` | 含むテスト資産 | 上位の test_plan |
|---|---|---|
| `manifest-eval` | `tests/nix-unit/` と `tests/namaka/` 配下の全テスト | [評価テストは nix-unit で不変条件を、namaka で manifest 全体のスナップショットを見る](docs/test-plan/nix-unit-namaka-split.md) / [評価テストの store-backed な入力は固定 outPath を持つ fake flake-input で与える](docs/test-plan/eval-test-double.md) / [lib.__internal で private helper をテストへ公開する](docs/test-plan/lib-internal-test-seam.md) |
| `engine-core` | `internal/engine/` の engine / dryrun / preflight、`internal/planner/` | [エンジンとコマンド層の Go テストは nix を介さない実 FS 統合テストを主戦力とする](docs/test-plan/go-test-layering.md) |
| `copy` | copytree / copy、e2e `04-copy` | [エンジンとコマンド層の Go テストは nix を介さない実 FS 統合テストを主戦力とする](docs/test-plan/go-test-layering.md) / [非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する](docs/test-plan/e2e-harness-scope.md) |
| `migration-stale` | preremove_generalization / staleremove、e2e `03-stale` | [エンジンとコマンド層の Go テストは nix を介さない実 FS 統合テストを主戦力とする](docs/test-plan/go-test-layering.md) / [非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する](docs/test-plan/e2e-harness-scope.md) |
| `atomicity` | undo / undo_journal / backup、engine の lock + `internal/lock/` | [原子性は実 FS の条件で故障を誘発して不変条件ごとに検証する](docs/test-plan/fault-injection-atomicity.md) / [エンジンとコマンド層の Go テストは nix を介さない実 FS 統合テストを主戦力とする](docs/test-plan/go-test-layering.md)（backup の退避ポリシー）|
| `generations` | generations / reset（engine）/ result_extensions / drift / prune、`internal/paths/`、`internal/generation/`、e2e `02-home` / `08-prune` | [エンジンとコマンド層の Go テストは nix を介さない実 FS 統合テストを主戦力とする](docs/test-plan/go-test-layering.md)（`internal/paths/` の純ロジックを含む）/ [非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する](docs/test-plan/e2e-harness-scope.md) |
| `cli-json` | `cmd/layat/` の全テストファイル、`internal/generator/` | [エンジンとコマンド層の Go テストは nix を介さない実 FS 統合テストを主戦力とする](docs/test-plan/go-test-layering.md)（CLI 層の判断）/ [エンベロープの outturn 適合を Go テストと E2E の両方で検証する](docs/test-plan/json-conformance-verification.md)（エンベロープ適合と payload 意味論）|
| `integration` | `checks.hm-module` / `checks.nixos-module` / `checks.darwin-module` / `checks.nixos-vm`、e2e `01-project` / `05-hm` / `06-init-templates` / `07-legacy` / `14-flake-parts`、`internal/gitutil/`、`internal/manifest/` | [非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する](docs/test-plan/e2e-harness-scope.md) / [home-manager モジュールの配線は build sandbox 内の評価アサートで検証する](docs/test-plan/hm-module-eval-assert.md) / [エンジンとコマンド層の Go テストは nix を介さない実 FS 統合テストを主戦力とする](docs/test-plan/go-test-layering.md)（`internal/` の Go テスト）|

区分外: `go-vet` / `golangci-lint` / カバレッジ計測は quality の担当。**CASE を
持たないテスト資産の規範は `dev/tests/test-doc-exclusions.tsv`**（除外理由付き。ここに無い
資産が CASE 無しで現れると契約テストが落ちる）。

既存のテスト実装から item を起こすときの粒度（CASE・TC・RISK の 3 段）と、risk / TC / CASE を
概要文書へ索引しない規則は [テスト系 item は、テスト資産・検証テーマ・脅威の 3 段の粒度で起こす](docs/quality/test-items-three-tier-granularity.md)。

### ID 規約（フル UUIDv4）

| 用途 | 形式 | 例 |
|---|---|---|
| 正式 ID（frontmatter `id:`・relation リスト）| `<PREFIX>-<フル UUIDv4>` | `REQ-2b0c2bb8-964f-4e36-a121-c6ea0d4be1c4` |
| ファイル名 | `<slug>.md` | `mkmanifest-pure-function.md` |
| 散文中の参照 | `[item の name](相対パス)` | `[mkManifest は配置データを生成する純粋関数である](../requirements/mkmanifest-pure-function.md)` |

ID とファイル名の形式・フル ID を書く場所・散文と `specification` での item への触れ方・ADR の採番は
[フル ID は frontmatter だけに書き、散文では item にリンクで触れる](docs/quality/full-ids-in-frontmatter-links-in-prose.md)。

- 起票は ADR を含め devShell 同梱の `sara-new` コマンドで行う。採番（`sara init` が引く
  UUIDv4）とファイル名規約の適用を一手で済ませ、`id:` と `file:` の 2 行を返す。連番を手で
  振らない（並列レーンでの採番衝突を構造的に避けるため）

  ```bash
  sara-new <型> <slug> <配置ディレクトリ> -- --description <要約> [<sara init のオプション>...]
  sara-new requirement lock-ordering docs/requirements -- --description '<1 文の要約>'
  ```

ADR の粒度・本文の構成・`justifies` の選定基準は [ADR は 1 本に 1 つの決定を 4 節で書き、justifies は規範の形成に寄与した item へ張る](docs/quality/adr-one-decision-four-sections-justifies-by-contribution.md)、
ステータス・出典・`supersedes` と `deprecated`・採番は [設計判断は ADR に記録し、改訂の書き戻しと item への接続を同じ変更の中で完了させる](docs/quality/adr-decisions-recorded-and-connected.md)、
layat への当てはめは `docs/adr/README.md`。`specification`（英語）と `specification_ja`（日本語）の
様式は [specification は SHALL 系で書き、日本語の規範文は同じ強度の規範助動詞で終える](docs/quality/specification-shall-family-and-matching-auxiliaries.md)。

### OKF 互換の frontmatter

item は OKF v0.2 と互換のキー（`description` / `status` / `generated` / `verified` / `stale_after` /
`sources` / `resource`）を持つ。キーの意味と書き方は `docs/model.md` の「OKF 互換の frontmatter」節。
起票では `sara-new` が `--description` を求め、`generated` を環境変数 `SARA_NEW_GENERATED_BY` の
actor で埋める。本文を変えたら `generated` の `by` / `at` を更新する。

`docs/index.md` と型ごとのディレクトリ・`docs/test/` 配下の `index.md` は `dev/scripts/okf-index.sh` の生成物で、
手で編集しない。item を足したり `name`・`description`・本文の冒頭を変えたりしたら、同じスクリプトで
再生成してコミットする（食い違いは契約テスト `dev/tests/okf-index.sh` が検出する）。

### item を辿る（`sara query`）

`sara` は devShell 経由で使う（`nix develop ./dev --command sara ...`）。

```bash
sara check                                 # グラフ全体の検証（broken reference / duplicate ID / cycles）
sara query <フル ID> -u                    # 上流を辿る（requirement → use_case → solution）
sara query <フル ID> -d                    # 下流を辿る（requirement → design / risk → test_condition → test_case）
sara report coverage                       # カバー率
sara report matrix                         # トレーサビリティマトリクス
```

**`sara query` はフル ID しか受け付けない**が、frontmatter の `id:` と relation リストはフル ID で書くので、
そこからそのままコピーして渡せる（変換は要らない）。同じ文字列で grep すると、宣言と relation の参照元が分かる。

```bash
rg -l REQ-2b0c2bb8-964f-4e36-a121-c6ea0d4be1c4 docs/   # `id:` で宣言している item と relation の参照元を列挙
```

## 開発コマンド

```bash
nix flake check          # 評価エラー・型チェック
nix build .#<package>    # パッケージビルド
nix run .#<script>       # activation スクリプト実行
nix develop ./dev --command sara check   # ドキュメントグラフの検証
nix develop '.?dir=dev#sara' -c dev/tests/term-avoid.sh   # avoid の綴りの検出（textlint + prh）
nix develop '.?dir=dev#sara' -c sara-fresh [--json]   # 上流より古い item と人の確認が無い high の risk の列挙（報告のみ）

# テストコード ⇔ CASE 対応（→ Issue #304）
nix develop '.?dir=dev#sara' -c dev/tests/test-doc-map.sh   # 契約テスト（純静的・毎 PR で回る）
nix develop ./dev -c dev/scripts/test-inventory.sh --static # テスト資産の列挙（ファイル粒度）
nix develop ./dev -c dev/scripts/test-doc-matrix.sh out.md  # 対応表の生成（重い・CI は main push）
```

## 規約

- **実装フェーズ**。main ブランチへの直接コミットは禁止。作業は必ずブランチを切り、PR 経由でマージする。
- `flake.nix` は `flake-parts.lib.mkFlake` ベース
- `lib/` は nixpkgs のみに依存する。home-manager / NixOS / nix-darwin への依存を持ち込まない
- **ドキュメントの配置ルール・ID 規約はこのプロジェクトの規約が優先する**（→「ドキュメント」節）。
  要件は `docs/requirements/`、リスクは `docs/risks/`、テスト成果物は `docs/test/<対象>/` へ
  1 ファイル 1 item で置き、起票は ADR を含め `sara-new` で行う。
  スキル既定の出力先・採番方式（散文への ID 直書き・連番）がこれと食い違う場合はプロジェクト規約に従う
- **コメント規約**: [コメントと利用者向け文字列は現在の動作だけを書き、決定の記録を参照しない](docs/quality/comments-state-current-behaviour-only.md)
- ユーザーに確認・質問する際は、テキストで質問を投げず **AskUserQuestion ツールを積極的に使う**。設計判断の確認・曖昧な依頼の解釈確認・代替案の選択などで使い、各質問は推奨オプションを先頭に置く

## Agent skills

### Issue tracker

Issue は GitHub Issues で管理し、`gh` コマンドで操作する。epic・sub-issue・defect の運用は
`docs/agents/issue-tracker.md`。

### Triage labels

トリアージのラベルは `needs-triage` / `needs-info` / `ready-for-agent` / `ready-for-human` / `wontfix` の 5 つ。
意味と、構造ラベル（`tracking` / `epic` / `task`）は `docs/agents/triage-labels.md`。

### Ubiquitous language

用語は `docs/glossary.md` の索引から `docs/ubiquitous-language/` の item を辿る。執筆では item の
`avoid` にある綴りを使わず、必要な語の item が無ければ起こす。手順は dev/skills/sara-docs の
`references/ubiquitous-language.md`。
`avoid` の綴りは契約テスト `dev/tests/term-avoid.sh`（textlint + prh）が検出する。対象は docs・README・CLAUDE.md で、ADR・SDR・用語 item は除く。

### Sara document graph

`docs/` は sara のナレッジグラフ（`docs/model.yaml` + `sara check`）。risk を
`requirement` と `design` のどちらに張るかの使い分けは [risk はまず requirement に張り、設計を差し替えると消える懸念だけを design に張る](docs/quality/risk-attaches-to-requirement-by-default.md)。
