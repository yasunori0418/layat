# sara CLI の使い方

sara（と `sara-new` / `sara-gap`）は **PATH に通っている前提**で `sara ...` と直接叩く
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
  --threatens "<REQ / DSG のフル ID>" \
  --likelihood medium --impact high --level high
```

- 返るのは機械可読の 2 行（`id:` = 採番された正式 ID、`file:` = 起票したファイルのパス）。
- `--` 以降は `sara init` へそのまま渡る。型ごとのオプション（relation・固有フィールド）は
  `sara init <型> --help` で確認する。requirement / quality / test_plan の
  `--specification` / `--specification-ja`、test_case の `--resource` など、モデルの必須
  フィールドはここで渡す。
- `--` 以降に `--description`（`-d` / `--description=…` でもよい）が無ければ、起票せず exit 2 で
  落ちる。環境変数 `SARA_NEW_GENERATED_BY` が空か未設定のときも exit 2 で落ちる。どちらも
  ファイルを作らない。
- 起票した item の frontmatter の末尾（閉じの `---` の直前）に `generated` を付ける。`by` は
  `SARA_NEW_GENERATED_BY` の値、`at` は起票時の `date --iso-8601=seconds` の値（UTC オフセット付き）。
- ADR は `adr` / `sara_adr` で起票する（`--status` を渡す）。
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
| 6 | `exempt` | `risk_exempt` を持つ `requirement`（情報。exit code に影響しない）|
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
