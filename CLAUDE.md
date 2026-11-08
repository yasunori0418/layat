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
| `docs/test/<対象>/` | test_case | `CASE` | test_condition |
| `docs/adr/` | adr | `ADR` | なし（分離型）|
| `docs/sara-adr/` | sara_adr | `SDR` | なし（分離型）|
| `docs/ubiquitous-language/` | ubiquitous_term | `UBQ` | なし（relation を持たない根）|

型・関係・フィールドの定義は `docs/model.yaml` が持つ。親を持たない根は **solution と
ADR の 2 型（adr / sara_adr）と ubiquitous_term のみ**で、他の型は upstream relation を 1 本以上張る
（ubiquitous_term は relation を 1 本も持たない。張り忘れは `sara check` の接続漏れの warning で出る）。
検出した欠陥（defect）は型として持たず、GitHub Issues（`bug` label）で管理する
（→ SDR-49733f2e-ed79-480d-bfb9-6361a7f95aed。運用規約は `docs/agents/issue-tracker.md` の「Defect issues」節）。

risk を `requirement` と `design` のどちらに張るかの使い分けは `docs/agents/sara-graph.md`。

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
| `manifest-eval` | `tests/nix-unit/` と `tests/namaka/` 配下の全テスト | TP-36e90d5d-4524-4294-bc72-ee263bb02782 / TP-d3d06fe4-6940-4df8-b111-bb4096d5444f / TP-403c55c7-d996-4951-8e6b-c3a7dddd387c |
| `engine-core` | `internal/engine/` の engine / dryrun / preflight、`internal/planner/` | TP-e7c25263-6d2d-4a37-8275-26906889d912 |
| `copy` | copytree / copy、e2e `04-copy` | TP-e7c25263-6d2d-4a37-8275-26906889d912 / TP-229b69c0-cf5e-4fb6-a353-27e5064d93e9 |
| `migration-stale` | preremove_generalization / staleremove、e2e `03-stale` | TP-e7c25263-6d2d-4a37-8275-26906889d912 / TP-229b69c0-cf5e-4fb6-a353-27e5064d93e9 |
| `atomicity` | undo / undo_journal / backup、engine の lock + `internal/lock/` | TP-deb05610-44bc-4962-8939-952392e5fbd0 / TP-e7c25263-6d2d-4a37-8275-26906889d912（backup の退避ポリシー）|
| `generations` | generations / reset（engine）/ result_extensions / drift / prune、`internal/paths/`、`internal/generation/`、e2e `02-home` / `08-prune` | TP-e7c25263-6d2d-4a37-8275-26906889d912（`internal/paths/` の純ロジックを含む）/ TP-229b69c0-cf5e-4fb6-a353-27e5064d93e9 |
| `cli-json` | `cmd/layat/` の全テストファイル、`internal/generator/` | TP-e7c25263-6d2d-4a37-8275-26906889d912（CLI 層の判断）/ TP-d3000054-42d9-4bac-912a-dd3abc38d3e9（エンベロープ適合と payload 意味論）|
| `integration` | `checks.hm-module` / `checks.nixos-module` / `checks.darwin-module` / `checks.nixos-vm`、e2e `01-project` / `05-hm` / `06-init-templates` / `07-legacy` / `14-flake-parts`、`internal/gitutil/`、`internal/manifest/` | TP-229b69c0-cf5e-4fb6-a353-27e5064d93e9 / TP-0734996e-aea9-4229-8075-89a64bdf9f79 / TP-e7c25263-6d2d-4a37-8275-26906889d912（`internal/` の Go テスト）|

区分外: `go-vet` / `golangci-lint` / カバレッジ計測は quality の担当。**CASE を
持たないテスト資産の規範は `dev/tests/test-doc-exclusions.tsv`**（除外理由付き。ここに無い
資産が CASE 無しで現れると契約テストが落ちる）。

**逆算階層の粒度**（既存のテスト実装から item を起こすときの単位）:

- **CASE** = テストファイル / e2e シナリオ / flake check 単位。frontmatter の `resource`
  （必須・単一値）に対象を正準表記で書き、本文の `## 対象` 節に人間向けの補足を書く。
  主な検証内容（テスト関数のテーマ列挙）も本文へ。テスト関数単位には割らない。
  CASE ⟷ テスト資産の 1:1 は `dev/tests/test-doc-map.sh` が強制する
- **TC** = 検証テーマ単位（対象あたり数件）。同じ不変条件・観点を検証する CASE 群を束ねる
- **RISK** = TC を束ねる脅威単位（対象あたり 2〜5 件）。「このテーマが壊れると何が起きるか」を
  requirement / design への `threatens` で表す
- 関係の向き: CASE -covers→ TC -mitigates→ RISK -threatens→ REQ / DSG

**risk / TC / CASE は `docs/spec.md` へ索引しない**（`sara query` / `sara report` で辿る）。
`docs/spec.md` のリンク集に載せるのは requirement / quality / test_plan まで。

### ID 規約（フル UUIDv4）

| 用途 | 形式 | 例 |
|---|---|---|
| 正式 ID（frontmatter `id:`・relation リスト）| `<PREFIX>-<フル UUIDv4>` | `REQ-2b0c2bb8-964f-4e36-a121-c6ea0d4be1c4` |
| ファイル名 | `<slug>.md` | `mkmanifest-pure-function.md` |
| 散文中の参照 | `<PREFIX>-<フル UUIDv4>` | `REQ-2b0c2bb8-964f-4e36-a121-c6ea0d4be1c4` |

**ID を書くときはフル ID をそのまま書く**（→ SDR-6c6c0681-4d5e-429e-8ee9-acf8dd94ad54）。8 文字省略形は使わない。
散文・relation リストの表記が一致するので、ID をそのまま grep すれば frontmatter の `id:` と参照元に当たり、
`sara query` へ渡しても同じものに当たる（省略形から正式 ID を探す手順が要らない）。

- 起票は ADR を含め devShell 同梱の `sara-new` コマンドで行う。採番（`sara init` が引く
  UUIDv4）とファイル名規約の適用を一手で済ませ、`id:` と `file:` の 2 行を返す。連番を手で
  振らない（並列レーンでの採番衝突を構造的に避けるため）

  ```bash
  sara-new <型> <slug> <配置ディレクトリ> -- --description <要約> [<sara init のオプション>...]
  sara-new requirement lock-ordering docs/requirements -- --description '<1 文の要約>'
  ```

- ADR も UUID で採番する（`ADR-<フル UUIDv4>`、sara 品質管理機構の決定は `SDR-<フル UUIDv4>`）。
  連番の `ADR-NNNN` は削除済みで、新 ADR の `origin` と「起点:」行に来歴として残る。
  ADR の規則は `docs/agents/adr-rules.md`、layat への当てはめは `docs/adr/README.md`
- `specification` フィールドを持つ型（requirement / quality / test_plan）は**英語で書く**。
  sara が RFC2119 キーワード（MUST / SHALL / SHOULD …）の存在をハードコードで検証するため。
  対になる日本語の規範文は `specification_ja` に併記する。ただし検証が効くのは requirement
  だけ（型名がハードコードされている）で、quality / test_plan の様式はレビューで守る

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

**`sara query` はフル ID しか受け付けない**が、散文も relation リストもフル ID で書くので、
目に付いた ID をそのままコピーして渡せる（変換は要らない）。同じ文字列で参照元も辿れる。

```bash
rg -l REQ-2b0c2bb8-964f-4e36-a121-c6ea0d4be1c4 docs/   # 宣言している item ファイルと参照元を列挙
```

## 開発コマンド

```bash
nix flake check          # 評価エラー・型チェック
nix build .#<package>    # パッケージビルド
nix run .#<script>       # activation スクリプト実行
nix develop ./dev --command sara check   # ドキュメントグラフの検証
nix develop '.?dir=dev#sara' -c dev/tests/term-avoid.sh   # avoid の綴りの検出（textlint + prh）

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
  要求は `docs/requirements/`、リスクは `docs/risks/`、テスト成果物は `docs/test/<対象>/` へ
  1 ファイル 1 item で置き、起票は ADR を含め `sara-new` で行う。
  スキル既定の出力先・採番方式（散文への ID 直書き・連番）がこれと食い違う場合はプロジェクト規約に従う
- **コメント規約**:
  - 連続するコメント行の散文は 3 行以内（箇条書きも散文に数える）
  - 現在の動作だけを書く。経緯と ADR / Issue / PR 参照を書かない
  - 利用者向け文字列（ヘルプ・エラーメッセージ等）に ADR 番号を出さない
  - コードブロックを残してよいのは判定表・利用者向けサンプル・実行例だけ
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
`requirement` と `design` のどちらに張るかの使い分け規約は `docs/agents/sara-graph.md`。
