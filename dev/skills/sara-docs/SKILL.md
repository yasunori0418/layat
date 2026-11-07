---
name: sara-docs
description: sara ドキュメントグラフの基盤知識・規約スキル。次のコンテキストが現れたら作業前に必ず読み込む — ① sara.toml と docs/model.yaml を持つリポジトリで docs/ 配下の item（YAML frontmatter に id / type / relation を持つ Markdown）を作成・編集・削除しようとしている。② REQ- / DSG- / RISK- / TC- / CASE- / TP- / QA- / INF- / UC- / SOL- / ADR- / SDR- / UBQ- 形式の ID、または threatens / mitigates / covers / derives_from / satisfies / refines / justifies といった relation が依頼・ファイル・エラーに現れた。③ sara / sara-new / sara-gap コマンドを実行しようとしている、または sara check の orphan・broken reference エラーを直そうとしている。④ 要求・設計・リスク・テスト条件・テストケース・ADR を「item として起こして」と依頼された。工程スキル（/test-plan・/test-analyze 等）が前提とする型・relation・ID・specification 規約の一次の参照先。sara を使わないリポジトリの Markdown ドキュメント一般は対象外。
license: MIT
---

# sara-docs: sara ドキュメントグラフの基盤規約

sara（Requirements Knowledge Graph CLI）で管理する `docs/` の読み書きに共通する規約を
1 箇所に集めた基盤スキル。工程スキル（/test-plan・/test-analyze 等）はこの規約を前提に
書かれており、ここと矛盾する出力をしない。

## 前提

- **sara は必須**。このスキルの対象リポジトリは `sara.toml` + `docs/model.yaml` を持ち、
  `sara check` が通ることを常時の不変条件とする（strict 有効なら orphan も error）。
  item を書き込んだら必ず `sara check` を回して green を確認してから完了とする。
- **1 ファイル 1 item**。工程単位の固定文書（`test-plan.md` のような 1 本もの）は作らない。
  全体像は文書ではなく `sara query` / `sara report matrix` で辿る。
- モデルは次の型セットに固定する（model.yaml がこの構造であることを前提にする）。

## 型セット

| 型 | prefix | ディレクトリ | 親（upstream relation）|
|---|---|---|---|
| solution | `SOL` | `docs/solution/` | なし（根）|
| use_case | `UC` | `docs/use-cases/` | solution（`refines`）|
| requirement | `REQ` | `docs/requirements/` | use_case（`derives_from`）|
| design | `DSG` | `docs/design/` | requirement / test_plan（`satisfies`）|
| quality | `QA` | `docs/quality/` | solution（`derives_from`）|
| test_plan | `TP` | `docs/test-plan/` | solution（`derives_from`）|
| infrastructure | `INF` | `docs/infrastructure/` | quality / design（`satisfies`）|
| adr | `ADR` | `docs/adr/` | なし（分離型。`justifies` で 1 本以上接続）|
| sara_adr | `SDR` | `docs/sara-adr/` | なし（adr と同じ。sara 品質管理機構の決定）|
| risk | `RISK` | `docs/risks/` | requirement / design（`threatens`）|
| test_condition | `TC` | `docs/test/…` | risk（`mitigates`）|
| test_case | `CASE` | `docs/test/…` | test_condition（`covers`）|
| ubiquitous_term | `UBQ` | `docs/ubiquitous-language/` | なし（relation を持たない根）|

テスト系 item（TC / CASE）を `docs/test/` 配下でどう区分けするか（サブディレクトリの
切り方）と、CASE の `resource`（テスト資産の正準表記）の規則はプロジェクト規約
（CLAUDE.md 等）に委ねる。

親を持たない根は solution と ADR の 2 型（adr / sara_adr）と ubiquitous_term のみ。ubiquitous_term を除く
他の型は upstream relation を 1 本以上張る
（張り忘れは strict check の orphan error になる）。frontmatter は手書きせず
**`sara-new` で生成する**（`sara init` を包んで採番とファイル名規約の適用まで行う。
relation・型別フィールドはオプションで渡す。手順は
[references/sara-cli.md](references/sara-cli.md)）。型・フィールドの正本は各リポジトリの
`docs/model.yaml`（`sara schema` で確認できる）。

### OKF 互換のキー（`description` と `generated`）

- `sara-new` は `--` 以降に `--description`（item の 1 文の要約）が無いと起票を拒否する（exit 2）
- `sara-new` は frontmatter に `generated`（`by` と `at`）を付ける。`by` は環境変数
  `SARA_NEW_GENERATED_BY` から取り、未設定なら起票を拒否する（exit 2）。人なら
  `human:<GitHub ID>`、エージェントなら `claude-code/<モデル ID>` を設定してから呼ぶ
- **本文を変えたら `generated` の `by` と `at` を更新する**（`at` は UTC オフセット付きの
  ISO 8601。`date --iso-8601=seconds` の値）。frontmatter だけの変更や改名では更新しない
- 他のキー（`status` / `verified` / `stale_after` / `sources` / `resource`）の書き方は
  各リポジトリの `docs/model.md` の「OKF 互換の frontmatter」節に従う

用語（ubiquitous_term）の引き方・執筆での使い方・起こし方は
[references/ubiquitous-language.md](references/ubiquitous-language.md)。

## forward の張り方（規範）

item は**上流から下流へ、親が先に在る順**で起こす。テスト系の系統なら:

1. REQ / DSG（仕様・設計）が在る
2. それを脅かす risk を起こし `threatens` を張る
3. risk を潰すテーマとして test_condition を起こし `mitigates` を張る
4. テーマを検証する実体として test_case を起こし `covers` を張る

relation は**起こした側（下流）の frontmatter に宣言する**（`sara check` の JSON も
宣言辺だけを持つ）。前工程の item が無いまま下流を起こさない — 無ければ前工程の実行を
提案して停止するのが工程スキルの原則。既存資産から下流を先に起こす逆算は規範外の
移行専用手順で、[references/reverse-import.md](references/reverse-import.md) に隔離する。

## ID 規約（フル UUIDv4）

| 用途 | 形式 | 例 |
|---|---|---|
| 正式 ID（frontmatter `id:`・relation リスト）| `<PREFIX>-<フル UUIDv4>` | `REQ-2b0c2bb8-964f-4e36-a121-c6ea0d4be1c4` |
| ファイル名 | `<YYYYMMDD>-<フル UUIDv4>-<slug>.md` | `20260802-2b0c2bb8-964f-4e36-a121-c6ea0d4be1c4-mkmanifest-pure-function.md` |
| 散文中の参照 | `<PREFIX>-<フル UUIDv4>` | `REQ-2b0c2bb8-964f-4e36-a121-c6ea0d4be1c4` |

- **どの面でもフル ID をそのまま書く**（→ SDR-722ce265-2aaf-4a90-915d-32d26b7a38bb）。8 文字省略形は使わない。relation
  リスト・ファイル名・散文の表記が一致するので、ID をそのまま grep しても `sara query` へ
  渡しても同じものに当たる。
- 起票は ADR を含め **`sara-new` コマンド**。採番とファイル名規約の適用、`sara init` の
  呼び出しまでを一手で済ませ、`id:` / `file:` の 2 行を返す。連番を手で振らない
  （並列レーンでの採番衝突を構造的に避けるため）。手順は
  [references/sara-cli.md](references/sara-cli.md)。
- ADR も UUID で採番する（`ADR-<フル UUIDv4>` / `SDR-<フル UUIDv4>`）。連番の `ADR-NNNN` は
  削除済みで、新 ADR の `origin` と「起点:」行に来歴として残る。ADR の規則（粒度・`supersedes` と `廃止`・
  `justifies` の選定基準）は `docs/agents/adr-rules.md`。

## risk の評価と優先度の導出

- **threatens の張り先は requirement が既定**。設計を差し替えたら消える懸念（設計選択に
  固有のリスク）だけを design へ張る。判断は risk 単位でなく edge ごとに行う。
- **risk の `level` は likelihood × impact から導出する**（手で判断しない）:

  | likelihood \ impact | high | medium | low |
  |---|---|---|---|
  | **high** | high | high | medium |
  | **medium** | high | medium | low |
  | **low** | medium | low | low |

  プロジェクトが独自の導出規約・契約テストを持つ場合はそちらが正本。
- **test_condition の優先度はフィールドにしない**。mitigates 先 risk の `level` から
  導出する（複数なら最高 level。例外は item 本文に根拠付きで記載）。

## specification 規約

`specification` フィールドを持つ型（requirement / quality / test_plan）は:

- **`specification` は英語**で書く。sara が RFC2119 キーワードの存在を検証するため。
  綴りは **SHALL 系に固定**（SHALL / SHALL NOT。MUST / MUST NOT は書かない。
  SHOULD / MAY はそのまま）。
- 対になる日本語の規範文を **`specification_ja` に併記**する。規範を述べる文は平叙形でなく
  規範助動詞で終える（SHALL = 〜しなければならない / SHALL NOT = 〜してはならない /
  SHOULD = 〜すべきである / MAY = 〜してもよい）。強度を 2 フィールド間でずらさない。

## 機械的検出の使い分け

- **`sara check`** — グラフの整合（broken reference・duplicate ID・cycle・orphan）。
  書き込みのたびに回す。green が完了条件。
- **`sara-gap`** — valid なグラフの**未カバー 3 段**（threatens されていない REQ/DSG・
  mitigates されていない risk・covers されていない TC）。工程の着手点リストとして使う
  （unthreatened = リスク識別の未着手、unmitigated = テスト分析の未着手、
  uncovered = テスト実装の未着手）。exit 0 = ギャップなし / 1 = あり / 2 = check 失敗。

CLI の詳しい使い方（`sara-new` による起票手順を含む）は
[references/sara-cli.md](references/sara-cli.md)。
