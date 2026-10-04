# sara モデル（docs/model.yaml）の読み方

`docs/model.yaml` は sara（https://github.com/cledouarec/sara）が `docs/` をナレッジグラフとして
検証するための型定義で、sara の組み込みモデルを全面的に置き換える（部分マージはされない）。
この文書は YAML だけでは読み取りにくい型・ID・フィールド・関係の意味を説明する。

- [型と prefix・配置ディレクトリ・親の一覧と ID 規約](../CLAUDE.md)（「ドキュメント」節）
- [型グラフの図と、item をどの型・どこへ張るかの判断規約](agents/sara-graph.md)
- [ADR の運用と `justifies` の選定基準](adr/README.md)

## 型構成

型は 11 個。親を持たない根は `solution` と `adr` だけで、`adr` も `justifies` を 1 本以上張るため、
接続漏れは全型で orphan warning として出る。

| 型 | 書くもの |
|---|---|
| `solution` | layat というプロダクトが何であるか（1 件）|
| `use_case` | ユーザーがどう使うか |
| `requirement` | 何を満たすべきか |
| `design` | どう実現するか。テストハーネスの実装形は `test_plan` を満たす `design` として書く |
| `quality` | 品質方針・規約・プロセス横断のガバナンス。テストプロセスの外にある関心事 |
| `test_plan` | テスト計画活動の成果物（テストスコープ・スコープ外宣言・テストレベル・アプローチ・テスト容易性）|
| `infrastructure` | 開発・提供・稼働を支える技術基盤（CI/CD・リリース・開発環境・配信・ホスティング・クラウド）。開発基盤は `quality` を、稼働基盤は `design` を満たす |
| `adr` | 意思決定の記録。仕様ツリーから独立し、`justifies` で決めた item へ接続する |
| `risk` | 満たされない懸念。`requirement` か `design` を脅かす |
| `test_condition` | リスクを潰すためのテスト条件 |
| `test_case` | 具体的なテストケース（テスト資産 1 つに 1 件）|

`risk` の親に `test_plan` は無い。テスト計画がどのリスクを扱うかは
`test_condition` の `mitigates` を辿って追跡する。

検出した欠陥（defect）は型として持たず、GitHub Issues（`bug` label）で管理する
（運用は [issue-tracker.md](agents/issue-tracker.md) の「Defect issues」節）。

組み込みモデルの `hardware_requirement` / `hardware_detailed_design` / `scenario` /
`system_architecture` は定義しない。組み込みの `system_requirement` / `software_requirement` /
`software_detailed_design` に当たるものは汎用名の `requirement` / `design` で表す。
型名・フィールド名は他プロジェクトでも使える汎用名に保ち、layat 固有の語彙を混ぜない。

## ID 形式

ID の表記・ファイル名・起票手順（`sara-new`）は [CLAUDE.md](../CLAUDE.md) の「ID 規約」節が持つ。
`model.yaml` の `id_format` はその規約をテンプレートで表したもので、`adr` だけが
`{prefix}-{seq:04}`（連番）、他の型は `{prefix}-{uuid4}` を使う。

sara 0.10.0 以降は `id_format` が採番・補完・検証を駆動する。パースできないテンプレートは
schema のロード時に拒否され、sara は組み込みモデルへ戻るため、カスタム型が全て unknown item
type になる。使えるプレースホルダは `{prefix}` / `{id}` / `{seq[:0N]}` / `{year}` / `{month}` /
`{day}` / `{date}` / `{uuid4}` / `{uuid7}`。

## フィールドの読み方

| 型 | フィールド | 読み方 |
|---|---|---|
| `requirement` / `quality` / `test_plan` | `specification` | 英語の規範文。RFC2119 キーワードを含める |
| `requirement` / `quality` / `test_plan` | `specification_ja` | `specification` と対になる日本語の規範文。sara はこのフィールドを検査しない |
| `infrastructure` | `dod` | `docs/dev/definition-of-done.md` の項目 ID（`DOD-01` …）のリスト。任意 |
| `adr` | `status` | `提案` / `採用` の 2 値 |
| `adr` | `status_note` | ステータス行のカッコ書き注記 |
| `adr` | `issues` / `origin` | item にできない外部参照（GitHub Issue 番号・起点となったセッション等）|
| `risk` | `likelihood` / `impact` / `level` | `high` / `medium` / `low`。採点規約と `level` の導出は [sara-graph.md](agents/sara-graph.md) |
| `test_case` | `target` | 対応するテスト資産の正準表記（単一値）|

`specification` / `specification_ja` の様式（SHALL 系に揃える・強度の写像）は
[sara-graph.md](agents/sara-graph.md) が持つ。

`dod` の値が実在する DoD 項目かは検証されない。sara は text フィールドの中身を見ないので、
存在しない項目 ID を書いても `sara check` は通る。

`adr` の `status` に「廃止」は無い。ADR 全体の失効を表す `supersedes` を定義していないためで、
両者は対で扱う。

### `test_case` の `target`

正準表記は次の 2 規則だけで、`dev/tests/test-doc-map.sh` が
CASE とテスト資産の 1:1 を照合する。本文の `## 対象` 節は人向けの補足で、照合しない。

- ファイルが実在するもの: リポジトリ相対パス（例: `internal/engine/engine_test.go`・
  `tests/nix-unit/structure.nix`・`tests/e2e/scenarios/03-stale.sh`・
  `tests/namaka/manifest-project/`〔ディレクトリは末尾スラッシュ〕）
- flake check: `checks.hm-module` の形（`checks.` 接頭辞）

## 関係

上流向き（upstream）の関係は子から親へ張り、逆向き（downstream）は sara が対で持つ。
`depends_on` / `revises` / `references` は同列（peer）の関係。

| 関係 | 向き | 張る元 → 先 | 意味 |
|---|---|---|---|
| `refines` | upstream | `use_case` → `solution` | 具体化 |
| `derives_from` | upstream | `requirement` → `use_case`、`quality` / `test_plan` → `solution` | 導出 |
| `satisfies` | upstream | `design` → `requirement` / `test_plan`、`infrastructure` → `quality` / `design` | 充足 |
| `justifies` | upstream | `adr` → `requirement` / `design` / `infrastructure` / `quality` / `test_plan` | 決定の根拠 |
| `threatens` | upstream | `risk` → `requirement` / `design` | 脅威 |
| `mitigates` | upstream | `test_condition` → `risk` | 緩和 |
| `covers` | upstream | `test_case` → `test_condition` | 網羅 |
| `depends_on` | peer | 同じ型の item どうし（`requirement` / `design` / `quality` / `test_plan` / `infrastructure`）| 依存 |
| `revises` | peer | `adr` → `adr` | ADR ヘッダの「改訂対象:」。節単位の部分改訂 |
| `references` | peer | `adr` → `adr` | ADR ヘッダの「関連:」。改訂を伴わない参照 |

`refines` / `derives_from` / `satisfies` / `depends_on` / `justifies` は組み込み由来、
`revises` / `references` / `threatens` / `mitigates` / `covers` は独自定義。
組み込みの `supersedes`（ADR 全体の失効）は定義せず、改訂は `revises` で表す。
