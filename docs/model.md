# sara モデル（docs/model.yaml）の読み方

`docs/model.yaml` は sara（https://github.com/cledouarec/sara）が `docs/` をナレッジグラフとして
検証するための型定義で、sara の組み込みモデルを全面的に置き換える（部分マージはされない）。
この文書は YAML だけでは読み取りにくい型・ID・フィールド・関係の意味を説明する。

- [型と prefix・配置ディレクトリ・親の一覧と ID 規約](../CLAUDE.md)（「ドキュメント」節）
- [型グラフの図と、item をどの型・どこへ張るかの規範の索引](agents/sara-graph.md)
- ADR の規則: [ADR は 1 本に 1 つの決定を 4 節で書き、justifies は規範の形成に寄与した item へ張る](quality/adr-one-decision-four-sections-justifies-by-contribution.md)・[設計判断は ADR に記録し、改訂の書き戻しと item への接続を同じ変更の中で完了させる](quality/adr-decisions-recorded-and-connected.md)・[layat への当てはめ](adr/README.md)

## 型構成

型の一覧は `docs/model.yaml` を一次の参照先とする。各型が書くものは次のとおり。

| 型 | 書くもの |
|---|---|
| `solution` | layat というプロダクトが何であるか（1 件）|
| `use_case` | 要求。ユーザーがどう使うか |
| `requirement` | 要件。何を満たすべきか |
| `design` | どう実現するか。テストハーネスの実装形は `test_plan` を満たす `design` として書く |
| `quality` | 品質方針・規約・プロセス横断のガバナンス。テストプロセスの外にある関心事 |
| `test_plan` | テスト計画活動の成果物（テストスコープ・スコープ外宣言・テストレベル・アプローチ・テスト容易性）|
| `infrastructure` | 開発・提供・稼働を支える技術基盤（CI/CD・リリース・開発環境・配信・ホスティング・クラウド）。開発基盤は `quality` を、稼働基盤は `design` を満たす |
| `adr` | 意思決定の記録。仕様ツリーから独立し、`justifies` で決めた item へ接続する |
| `sara_adr` | sara による品質管理機構についての意思決定の記録。`adr` と同じフィールドと関係を持つ |
| `risk` | 満たされない懸念。`requirement` か `design` を脅かす |
| `test_condition` | リスクを潰すためのテスト条件 |
| `test_case` | 具体的なテストケース（テスト資産 1 つに 1 件）|
| `ubiquitous_term` | ユビキタス言語の 1 語。relation を持たない根で、どの item ともつながない。索引は `docs/ubiquitous-language/index.md` |

`test_plan` から `test_case` へは `is_followed_by`（`test_case` が張る `follows` の逆向き）で届く。
`risk` の親に `test_plan` は無いままで、テスト計画がどのリスクを扱うかは、`test_case` から
`covers` と `mitigates` を辿って追跡する。

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
| `requirement` / `design` / `ubiquitous_term` / `adr` / `sara_adr` / `risk` | `section` | 型ごとの index.md で item を載せる節。必須の enum で、取りうる値は `docs/model.yaml` が持つ。見出しの文言と節の並び順は `dev/scripts/okf-index.sh` の表が持ち、表に無い値があると生成が exit 2 で止まる。値を足すときは両方を直す。`risk` の値はテストの区分と同じで、見出しと並び順は `dev/tests/test-categories.tsv` が持つ |
| `requirement` | `risk_exempt` | `requirement` 水準の `risk` を持たない理由を 1 文で書く。任意。`sara-gap` は値を持つ `requirement` を免除として扱う |
| `infrastructure` | `dod` | `docs/dev/definition-of-done.md` の項目 ID（`DOD-01` …）のリスト。任意 |
| `adr` / `sara_adr` | `origin` | 起点となったセッション等の散文。任意 |
| `adr` / `sara_adr` | `sources` | 起点となった GitHub Issue 等の出典。書き方は「OKF 互換の frontmatter」節 |
| `risk` | `likelihood` / `impact` / `level` | `high` / `medium` / `low`。採点規約と `level` の導出は [risk の level は likelihood と impact から導き、手で決めない](quality/risk-level-derived-from-likelihood-and-impact.md) |
| `test_case` | `resource` | 対応するテスト資産の正準表記（単一値）|
| `ubiquitous_term` | `name_ja` | 日本語の正名。英語のまま使う語では省く。英語の正名は組み込みの `name` に置く |
| `ubiquitous_term` | `avoid` | 執筆で使わない綴りのリスト。任意。綴りだけを書き、説明は本文に置く |

`specification` / `specification_ja` の様式（SHALL 系に揃える・強度の写像）は
[specification は SHALL 系で書き、日本語の規範文は同じ強度の規範助動詞で終える](quality/specification-shall-family-and-matching-auxiliaries.md) が持つ。

`dod` の値が実在する DoD 項目かは検証されない。sara は text フィールドの中身を見ないので、
存在しない項目 ID を書いても `sara check` は通る。

`supersedes` で指された ADR は `status` が `deprecated` になる。sara はこの対応を検証しないので、
`dev/tests/adr-supersede.sh` が frontmatter から確かめる。`deprecated` は提案を棄却した ADR にも
付くため、逆向き（`deprecated` の ADR が `supersedes` で指されるか）は確かめない。

### `test_case` の `resource`

正準表記は次の 2 規則だけで、`dev/tests/test-doc-map.sh` が
CASE とテスト資産を比べて 1:1 かを確かめる。本文の `## 対象` 節は人向けの補足で、比較の対象にしない。

- ファイルが実在するもの: リポジトリ相対パス（例: `internal/engine/engine_test.go`・
  `tests/nix-unit/structure.nix`・`tests/e2e/scenarios/03-stale.sh`・
  `tests/namaka/manifest-project/`〔ディレクトリは末尾スラッシュ〕）
- flake check: `checks.hm-module` の形（`checks.` 接頭辞）

## OKF 互換の frontmatter

item は Open Knowledge Format（OKF）v0.2 と互換のキーを持つ。対応はキー単位で取り、バンドル単位の
適合（`okf_version` の宣言・`log.md`・frontmatter の無い文書への frontmatter 追加）は追わない。
決定の理由は [OKF v0.2 と互換の frontmatter キーを item に足す](sara-adr/okf-compatible-frontmatter.md) とその後続の SDR が持つ。

| キー | 対象の型 | 書き方 |
|---|---|---|
| `type` | 全て | sara の組み込みフィールドをそのまま OKF の `type` とみなす |
| `name` | 全て | OKF の `title` とみなす。`title` は書かない |
| `description` | 全て | item の 1 文の要約。新しく起こす item では必須。既存の item は index.md の生成器が本文の最初の段落で代用する |
| `status` | 全て | `draft`（検討中でまだ有効でない）/ `stable`（有効）/ `deprecated`（もう有効でない。`requirement` / `design` では棄却した提案の記録も表す）の 3 値 |
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
- `verified` は自動では付けない。人が内容を確かめたときにだけ、手で要素を足す。`verified` は人が内容を確かめた記録で、
  要素を足せるのは人だけである。`human:` 以外の actor が `verified` に要素を足してはならない
- `human:` 以外の actor の要素が付いていた場合、基準時刻には数えるが、人の確認としては数えない

item の**基準時刻**はその item の `generated.at` と、`generated.at` 以降の `verified[].at`
（actor を問わない）のうち最も新しい時刻とする。`verified` が無い item では `generated.at` が基準時刻になる。

`generated` と `verified` は文書の鮮度と信頼を確かめる 3 つの確認に使われる。どれも結果を報告に留める。

- 再検証待ち: 上流の relation のうち `justifies` を除く 7 種で指している item の `generated.at` が、
  自身の基準時刻より新しい item を挙げる。peer の relation は対象にしない（→ [上流より古い item と generated の更新漏れは機械的に列挙する](quality/stale-items-enumerated-mechanically.md)）
- `generated` の更新漏れ: base との差分で本文が変わったのに `generated.at` が変わっていない item を挙げる
  （→ 同上）
- 人の確認が無い high の risk: `level` が high の risk のうち、`generated.at` 以降の `human:` の `verified` を
  持たないものを挙げる（→ [level が high の risk は人が確かめた記録を持つ](quality/high-risk-verified-by-human.md)）

再検証待ちと人の確認が無い high の risk は、devShell の `sara-fresh` が列挙する。`generated` の更新漏れは、`--base <ref>` を渡した `sara-fresh` が列挙する。

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

値は「`test_case` の `resource`」節の 2 規則に従う。OKF の `resource` は URI を想定するが、
layat の値は URI ではない。根もリポジトリ根で、index.md のバンドル根（`docs/`）とは異なる。

### 検証の分担

平たいキーのうち `status` と `resource` は、`docs/model.yaml` のフィールドとして sara が確かめる。
`stale_after` も `docs/model.yaml` のフィールドだが、sara は値の形を確かめない。`stale_after` と、
sara が型を持たない入れ子のキー（`generated` / `verified` / `sources`）は、
`dev/tests/okf-frontmatter.sh` が形を確かめる。本文を変えたのに `generated` を更新していない item は、
`sara-fresh --base <ref>` が base との差分から列挙して報告する。列挙はマージを止めない。

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
| `follows` | upstream | テスト計画への準拠。逆向きは `is_followed_by` |
| `depends_on` | peer | 同じ型の item どうしの依存 |
| `supersedes` | peer | `adr` / `sara_adr` から置き換えた ADR へ。文書丸ごとの失効 |
| `references` | peer | ADR ヘッダの「関連:」。改訂を伴わない参照 |

`refines` / `derives_from` / `satisfies` / `depends_on` / `justifies` / `supersedes` は組み込み由来、
`references` / `threatens` / `mitigates` / `covers` / `follows` は独自定義。
決定を変えるときは ADR を丸ごと書き直して `supersedes` で置き換える。

## item の `## 出典` の読み方

item の `## 出典` は、item へ分割した時点で原文のどこにあったかを記録している。出典が名指しする
`docs/spec.md`・`docs/design.md`・`docs/concept.md`・`docs/glossary.md`・`docs/publication-roadmap.md` は、
現在のリポジトリに無い。item から現行の仕様へ辿るには relation（`sara query`）を使う。

原文を読むときは、出典が名指しする文書を縮退させたコミットの親を見る。`docs/publication-roadmap.md` は
縮退させずに削除したので、表には削除したコミットを載せている。原文はそのコミットの親にある。

| 出典が名指しする文書 | 縮退させたコミット |
|---|---|
| `docs/spec.md` | `cf76bf89` |
| `docs/design.md` | `705766d1` |
| `docs/concept.md` | `686f9604` |
| `docs/glossary.md` | `341d4ed3` |
| `docs/publication-roadmap.md` | `5aa5b6a0` |

```bash
git show cf76bf89^:docs/spec.md   # 縮退させたコミットの親が、縮退前の原文
```
