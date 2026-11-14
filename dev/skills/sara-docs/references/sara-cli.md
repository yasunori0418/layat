# sara CLI の使い方

sara（と `sara-new` / `sara-gap` / `sara-fresh`）は **PATH に通っている前提**で `sara ...` と直接叩く
（direnv 等で開発シェルが読み込まれた環境なら通っている）。通っていなければ環境構築を
プロジェクトの規約に従って済ませてから使う。設定はカレント（リポジトリルート）の
`sara.toml` を読むため、**リポジトリルートで実行する**。

## sara-new — item の起票（採番 + frontmatter の生成）

frontmatter は手書き・テンプレ写経をせず `sara-new` で起票する。採番・ファイル名規約の
適用・`sara init` の呼び出しを一手で済ませる:

```bash
sara-new <型> <slug> <配置ディレクトリ> -- --description <要約> [<sara init のオプション>...]

export SARA_NEW_GENERATED_BY="claude-code/<モデル ID>"   # 人なら human:<GitHub ID>
sara-new risk lock-ordering docs/risks -- \
  --name "<日本語の 1 行要約>" \
  --description "<item の 1 文の要約>" \
  --section "<テストの区分>" \
  --threatens "<REQ / DSG のフル ID>" \
  --likelihood medium --impact high --level high
```

- 返るのは機械可読の 2 行（`id:` = 採番された正式 ID、`file:` = 起票したファイルのパス）。
- `--` 以降は `sara init` へそのまま渡る。型ごとのオプション（relation・固有フィールド）は
  `sara init <型> --help` で確認する。requirement / quality / test_plan の
  `--specification` / `--specification-ja`、`requirement` / `design` / `ubiquitous_term` /
  `adr` / `sara_adr` / `risk` の `--section`、test_case の `--resource` など、モデルの
  必須フィールドはここで渡す。`section` を持つ型で `--section`（`--section=…` でもよい）を
  省くと、`sara-new` は起票せず exit 2 で落ちる。値と index.md の見出しの対応は
  `dev/scripts/okf-index.sh` の表で確かめる（`risk` は `dev/tests/test-categories.tsv`）。
- `--` 以降に `--description`（`-d` / `--description=…` でもよい）が無ければ、起票せず exit 2 で
  落ちる。環境変数 `SARA_NEW_GENERATED_BY` が空か未設定のときも exit 2 で落ちる。どちらも
  ファイルを作らない。
- 起票した item の frontmatter の末尾（閉じの `---` の直前）に `generated` を付ける。`by` は
  `SARA_NEW_GENERATED_BY` の値、`at` は起票時の `date --iso-8601=seconds` の値（UTC オフセット付き）。
- ADR は `adr` / `sara_adr` で起票する（`--status` と `--section` を渡す）。
- 生成後に本文（`# <name>` の見出し + 散文）を書き足し、`sara check` で green
  を確認する。
- 既存 item のフィールド・relation の変更は `sara edit <フル ID> --<フィールド> ...`
  でもよい（対話モードは TTY 前提なので、非対話ではフラグを明示する）。

## check — グラフ検証

```bash
sara check                 # broken reference / duplicate ID / cycle / orphan
sara check --format json   # items 配列（宣言辺つき）を機械可読で出す
```

- item を書き込んだら必ず回す。green（exit 0）が完了条件。
- strict 有効（`sara.toml` の `strict_mode = true`）なら orphan を含む全 warning が
  error になる。orphan は「upstream relation の張り忘れ」を意味することが多い。
- `--format json` の `items[].relations`（relation 名 → 張り先 ID の配列）は**宣言辺
  （primary relation）のみ**で逆辺を持たない。逆引き（「この REQ を threatens している risk は?」）が要るときは
  `sara query -d` か `sara-gap` を使う。

## query — グラフを辿る

```bash
sara query <フル ID> -u    # 上流へ（例: requirement → use_case → solution）
sara query <フル ID> -d    # 下流へ（例: requirement → design、REQ ← risk ← TC ← CASE）
sara query <フル ID> -u --depth 1
sara query <フル ID> -d --format json
```

**`sara query` はフル ID しか受け付けない**が、frontmatter の `id:` と relation リストは
フル ID で書くので、そこからそのままコピーして渡せる（省略形からの復元は要らない）。
同じ文字列で grep すれば、宣言側（frontmatter の `id:`）と参照側（relation リスト）の
両方に当たる:

```bash
rg -l REQ-2b0c2bb8-964f-4e36-a121-c6ea0d4be1c4 docs/   # 宣言・参照元の列挙
```

## report — 俯瞰

```bash
sara report coverage   # 型ごとのカバー率
sara report matrix     # トレーサビリティマトリクス（--format json で機械可読）
```

工程固定の一覧文書を作らない代わりに、全体像はこの 2 つで取る。

## diff — 参照間の比較

```bash
sara diff <git-ref>    # 指定 ref とのグラフ差分（item / relation の増減）
```

PR レビューで「このブランチでグラフに何が増えたか」を見るときに使う。

## schema — モデルの確認

```bash
sara schema            # 有効な model schema を YAML で出力
```

型・relation・フィールド定義の正本は `docs/model.yaml`。編集前に schema で現状を確認する。

## sara-gap — 未カバーの 7 節と到達率（別コマンド）

```bash
sara-gap           # text: 7 節の見出し + <フル ID>\t<name>\t<file>
sara-gap --json    # {unthreatened, unmitigated, uncovered, unfollowed, unsatisfied_quality, exempt, reach}
```

| 節 | `--json` のキー | 内容 |
|---|---|---|
| 1 | `unthreatened` | `threatens` も `risk_exempt` も無い `requirement`。`design` は列挙しない。`design` 側にだけ `risk` があれば text の行末に `(design 側に risk あり)` を付ける |
| 2 | `unmitigated` | `mitigates` されていない `risk` |
| 3 | `uncovered` | `covers` されていない `test_condition` |
| 4 | `unfollowed` | `follows` を持たない `test_case` |
| 5 | `unsatisfied_quality` | `infrastructure` に `satisfies` されていない `quality` |
| 6 | `exempt` | `risk_exempt` を持つ `requirement`。情報として出し、exit code に影響しない |
| 7 | `reach` | `requirement` / `test_plan` / `quality` ごとの `{total, reached, ratio}` |

節 1〜6 の `--json` の行は `{id, name, file}` で、`id` はフル ID。節 7 の text の行は
`<型>\t<reached>/<total>\t<割合>%`。

到達と数えるのは次のもの。

- `requirement`: `risk_exempt` を持つものを `total` から除き、`requirement` ← `risk` ← `test_condition` ← `test_case` の辺がすべてつながるもの
- `test_plan`: 1 件以上の `test_case` から `follows` されるもの
- `quality`: 1 件以上の `infrastructure` から `satisfies` されるもの

exit code は、節 1〜5 のすべてが空で `ratio` がすべて 1 なら 0、それ以外は 1。`sara check` の失敗と
JSON 形状の異常は 2。

sara 本体ではなく devShell 同梱の独立コマンド。valid なグラフに対してだけ動く
（`sara check` が落ちる間は exit 2 で一覧を出さない）。工程スキルは着手点の特定に使う。

## 別コマンド sara-fresh による再検証待ち・generated の更新漏れ・人の確認が無い high の risk の列挙

```bash
sara-fresh                     # text: 節 1・節 3 の見出し + 各行
sara-fresh --base main         # text: 節 1〜3 の見出し + 各行
sara-fresh --json [--base <ref>]  # {stale, generated_not_updated, unverified_high_risk}
```

| 節 | `--json` のキー | 内容 |
|---|---|---|
| 1 | `stale` | `refines` / `derives_from` / `satisfies` / `threatens` / `mitigates` / `covers` / `follows` で指している上流の `generated.at` が、自身の基準時刻より新しい item。text の行末に、その上流のフル ID をカンマ区切りで足す |
| 2 | `generated_not_updated` | `--base <ref>` を渡したときだけ出す。`git merge-base <ref> HEAD` のコミットと作業ツリーを比べ、本文が変わったのに `generated.at` の値が変わっていない item |
| 3 | `unverified_high_risk` | `level: high` の `risk` のうち、`by` が `human:` で始まり `at` が `generated.at` 以降の `verified` を持たないもの |

基準時刻は `generated.at` と `generated.at` 以降の `verified[].at`（actor を問わない）のうち最も新しい時刻。
`justifies` と peer の relation（`depends_on` / `supersedes` / `references`）は辿らない。`status: deprecated` の
item はどの節でも扱わない。

節 2 の本文は frontmatter の外の行で、`# ` で始まる H1 の行と `- ステータス:` で始まる行を除き、空白だけの差は
数えない。コードフェンスの中の `# ` で始まる行は H1 として除かず、本文として比べる。2 点の間で足した item と削除した item は比べない。base に同じパスが無い item は、同じ ID を持つ
base 側のファイルを旧パスとみなして比べる。未コミットの変更も比べる対象に入る。

`stale` の `--json` の要素は `{id, name, file, upstreams}` で、`upstreams` は上流ごとの
`{id, relation, at}`（`at` は上流の `generated.at`）。`generated_not_updated` と `unverified_high_risk` の
要素は `{id, name, file}`。`--base` を渡さないとき、`--json` に `generated_not_updated` のキーは無い。

exit code は、出力したどの節も空なら 0、いずれかの節に item があれば 1。引数の誤り、`sara check` の失敗、
JSON 形状の異常、frontmatter を YAML として読めない item、`verified` がリストでない item、`generated.at` や
`verified` の `at` を時刻として読めない item があるとき、`--base` の ref を解決できないとき、`--base` の ref と
HEAD の merge-base が無いときは、一覧を出さずに 2 を返す。結果は報告に留め、`verified` は人が内容を確かめた
ときにだけ手で足す。CI の `sara` ジョブは pull_request で base ブランチを `--base` に渡して実行し、exit 1 では
落ちない。
