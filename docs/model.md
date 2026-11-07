# sara モデル（docs/model.yaml）の読み方

`docs/model.yaml` は sara（https://github.com/cledouarec/sara）が `docs/` をナレッジグラフとして
検証するための型定義で、sara の組み込みモデルを全面的に置き換える（部分マージはされない）。
この文書は YAML だけでは読み取りにくい型・ID・フィールド・関係の意味を説明する。

- [型と prefix・配置ディレクトリ・親の一覧と ID 規約](../CLAUDE.md)（「ドキュメント」節）
- [型グラフの図と、item をどの型・どこへ張るかの判断規約](agents/sara-graph.md)
- [ADR の規則と `justifies` の選定基準](agents/adr-rules.md)・[layat への当てはめ](adr/README.md)

## 型構成

型の一覧は `docs/model.yaml` を一次の参照先とする。各型が書くものは次のとおり。

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
| `sara_adr` | sara による品質管理機構についての意思決定の記録。`adr` と同じフィールドと関係を持つ |
| `risk` | 満たされない懸念。`requirement` か `design` を脅かす |
| `test_condition` | リスクを潰すためのテスト条件 |
| `test_case` | 具体的なテストケース（テスト資産 1 つに 1 件）|
| `ubiquitous_term` | ユビキタス言語の 1 語。relation を持たない根で、どの item ともつながない。索引は `docs/glossary.md` |

`risk` の親に `test_plan` は無い。テスト計画がどのリスクを扱うかは
`test_condition` の `mitigates` を辿って追跡する。

組み込みモデルの `hardware_requirement` / `hardware_detailed_design` / `scenario` /
`system_architecture` は定義しない。組み込みの `system_requirement` / `software_requirement` /
`software_detailed_design` に当たるものは汎用名の `requirement` / `design` で表す。
型名・フィールド名は他プロジェクトでも使える汎用名に保ち、layat 固有の語彙を混ぜない。

## ID 形式

ID の表記・ファイル名・起票手順（`sara-new`）は [CLAUDE.md](../CLAUDE.md) の「ID 規約」節が持つ。
`model.yaml` の `id_format` はその規約をテンプレートで表したもので、全型が `{prefix}-{uuid4}` を使う。

sara 0.10.0 以降は `id_format` が採番・補完・検証を駆動する。パースできないテンプレートは
schema のロード時に拒否され、sara は組み込みモデルへ戻るため、カスタム型が全て unknown item
type になる。使えるプレースホルダは `{prefix}` / `{id}` / `{seq[:0N]}` / `{year}` / `{month}` /
`{day}` / `{date}` / `{uuid4}` / `{uuid7}`。

## フィールドの読み方

| 型 | フィールド | 読み方 |
|---|---|---|
| 全て | `status` | `draft` / `stable` / `deprecated` の 3 値。必須 |
| 全て | `stale_after` | 期限の時刻。任意。値の形は sara でなく `dev/tests/okf-frontmatter.sh` が確かめる |
| `requirement` / `quality` / `test_plan` | `specification` | 英語の規範文。RFC2119 キーワードを含める |
| `requirement` / `quality` / `test_plan` | `specification_ja` | `specification` と対になる日本語の規範文。sara はこのフィールドを検査しない |
| `infrastructure` | `dod` | `docs/dev/definition-of-done.md` の項目 ID（`DOD-01` …）のリスト。任意 |
| `adr` / `sara_adr` | `origin` | 起点となったセッション等の散文。任意 |
| `adr` / `sara_adr` | `sources` | 起点となった GitHub Issue 等の出典。書き方は「OKF 互換の frontmatter」節 |
| `risk` | `likelihood` / `impact` / `level` | `high` / `medium` / `low`。採点規約と `level` の導出は [sara-graph.md](agents/sara-graph.md) |
| `test_case` | `target` | 対応するテスト資産の正準表記（単一値）|
| `ubiquitous_term` | `name_ja` | 日本語の正名。英語のまま使う語では省く。英語の正名は組み込みの `name` に置く |
| `ubiquitous_term` | `avoid` | 執筆で使わない綴りのリスト。任意。綴りだけを書き、説明は本文に置く |

`specification` / `specification_ja` の様式（SHALL 系に揃える・強度の写像）は
[sara-graph.md](agents/sara-graph.md) が持つ。

`dod` の値が実在する DoD 項目かは検証されない。sara は text フィールドの中身を見ないので、
存在しない項目 ID を書いても `sara check` は通る。

`supersedes` で指された ADR は `status` が `deprecated` になる。sara はこの対応を検証しないので、
`dev/tests/adr-supersede.sh` が frontmatter から確かめる。`deprecated` は提案を棄却した ADR にも
付くため、逆向き（`deprecated` の ADR が `supersedes` で指されるか）は確かめない。

### `test_case` の `target`

正準表記は次の 2 規則だけで、`dev/tests/test-doc-map.sh` が
CASE とテスト資産を比べて 1:1 かを確かめる。本文の `## 対象` 節は人向けの補足で、比較の対象にしない。

- ファイルが実在するもの: リポジトリ相対パス（例: `internal/engine/engine_test.go`・
  `tests/nix-unit/structure.nix`・`tests/e2e/scenarios/03-stale.sh`・
  `tests/namaka/manifest-project/`〔ディレクトリは末尾スラッシュ〕）
- flake check: `checks.hm-module` の形（`checks.` 接頭辞）

## OKF 互換の frontmatter

item は Open Knowledge Format（OKF）v0.2 と互換のキーを持つ。対応はキー単位で取り、バンドル単位の
適合（`okf_version` の宣言・`log.md`・frontmatter の無い文書への frontmatter 追加）は追わない。
決定の理由は SDR-bab72f31-7cb1-47a3-bb2d-7876757d86fb とその後続の SDR が持つ。

| キー | 対象の型 | 書き方 |
|---|---|---|
| `type` | 全て | sara の組み込みフィールドをそのまま OKF の `type` とみなす |
| `name` | 全て | OKF の `title` とみなす。`title` は書かない |
| `description` | 全て | item の 1 文の要約。新しく起こす item では必須。既存の item は index.md の生成器が本文の最初の段落で代用する |
| `status` | 全て | `draft`（検討中でまだ有効でない）/ `stable`（有効）/ `deprecated`（もう有効でない）の 3 値 |
| `generated` | 全て | `by`（actor）と `at`（時刻）の組 |
| `verified` | 全て | `generated` と同じ形の要素のリスト。任意 |
| `stale_after` | 全て | 期限の時刻。任意 |
| `sources` | `adr` / `sara_adr` | `id` / `resource` / `title` の組のリスト |
| `resource` | `test_case` | 対応するテスト資産の正準表記 |

`tags` は使わない。

### `generated` と `verified`

```yaml
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T21:04:05+09:00"
verified:
  - by: "human:yasunori0418"
    at: "2026-11-08T10:00:00+09:00"
```

- actor（`by`）は、人なら `human:<GitHub ID>`、エージェントなら `claude-code/<モデル ID>` の形で書く
- 時刻（`at`）は UTC オフセット付きの ISO 8601 で書く。`Z` も UTC オフセットとして扱う
- `generated.at` は本文を最後に意味のある形で変えた時刻にする。**本文を変えたら `generated` の
  `by` と `at` を更新する**。frontmatter だけの変更や改名では更新しない。frontmatter の値を本文へ
  写しただけの変更（ステータス行の書き換えなど）でも更新しない
- 起票では `sara-new` が `generated` を付ける（`by` は環境変数 `SARA_NEW_GENERATED_BY` から取る）
- `verified` は自動では付けない。人が内容を確かめたときにだけ、手で要素を足す

### `stale_after`

UTC オフセット付きの ISO 8601 の時刻を書く。外部ツールの版や公開状況など、期限のある事実を含む
item にだけ書き手が付ける。

### `sources`

```yaml
sources:
  - id: "issue-662"
    resource: "https://github.com/yasunori0418/layat/issues/662"
    title: "<Issue のタイトル>"
```

- `resource` には Issue の URL を書く。このリポジトリの `#N` は
  `https://github.com/yasunori0418/layat/issues/N`、他のリポジトリの `owner/repo#N` は
  `https://github.com/owner/repo/issues/N` にする
- `id` は `issue-N`（他のリポジトリなら `issue-<repo>-N`）の形にし、1 つの item の中で一意にする
- `title` には Issue のタイトルを書く。`issues` から移したエントリだけは、元の文字列（`#N` /
  `owner/repo#N`）を `title` にする
- `origin` は散文のまま layat の拡張キーとして残し、`sources` には入れない

### `resource`

値の規則は「`test_case` の `target`」の 2 規則と同じ。OKF の `resource` は URI を想定するが、
layat の値は URI ではない。根もリポジトリ根で、index.md のバンドル根（`docs/`）とは異なる。

### 検証の分担

平たいキーのうち `status` と `resource` は、`docs/model.yaml` のフィールドとして sara が確かめる。
`stale_after` も `docs/model.yaml` のフィールドだが、sara は値の形を確かめない。`stale_after` と、
sara が型を持たない入れ子のキー（`generated` / `verified` / `sources`）は、
`dev/tests/okf-frontmatter.sh` が形を確かめる。本文の変更と `generated` の更新が揃っているかは
確かめず、規約で守る。

## 関係

上流向き（upstream）の関係は子から親へ張り、逆向き（downstream）は sara が対で持つ。
どの型からどの型へ張れるかは [sara-graph.md](agents/sara-graph.md) の型グラフの図が示す。

| 関係 | 向き | 意味 |
|---|---|---|
| `refines` | upstream | 具体化 |
| `derives_from` | upstream | 導出 |
| `satisfies` | upstream | 充足 |
| `justifies` | upstream | 決定の根拠 |
| `threatens` | upstream | 脅威 |
| `mitigates` | upstream | 緩和 |
| `covers` | upstream | 網羅 |
| `depends_on` | peer | 同じ型の item どうしの依存 |
| `supersedes` | peer | `adr` / `sara_adr` から置き換えた ADR へ。文書丸ごとの失効 |
| `references` | peer | ADR ヘッダの「関連:」。改訂を伴わない参照 |

`refines` / `derives_from` / `satisfies` / `depends_on` / `justifies` / `supersedes` は組み込み由来、
`references` / `threatens` / `mitigates` / `covers` は独自定義。
決定を変えるときは ADR を丸ごと書き直して `supersedes` で置き換える。
