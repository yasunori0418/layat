---
name: sara-docs
description: sara ドキュメントグラフの基盤知識・規約スキル。次のコンテキストが現れたら作業前に必ず読み込む — ① sara.toml と docs/model.yaml を持つリポジトリで docs/ 配下の item（YAML frontmatter に id / type / relation を持つ Markdown）を作成・編集・削除しようとしている。② REQ- / DSG- / RISK- / TC- / CASE- / TP- / QA- / INF- / UC- / SOL- / ADR- / SDR- / UBQ- 形式の ID、または threatens / mitigates / covers / follows / derives_from / satisfies / refines / justifies といった relation が依頼・ファイル・エラーに現れた。③ sara / sara-new / sara-gap / sara-fresh コマンドを実行しようとしている、または sara check の orphan・broken reference エラーを直そうとしている。④ 要求・設計・リスク・テスト条件・テストケース・ADR を「item として起こして」と依頼された。工程スキル（/test-plan・/test-analyze 等）が前提とする型・relation・ID・specification 規約の一次の参照先。sara を使わないリポジトリの Markdown ドキュメント一般は対象外。
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
| test_case | `CASE` | `docs/test/…` | test_condition（`covers`）/ test_plan（`follows`）|
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
4. テーマを検証する実体として test_case を起こし `covers` を張る。従う計画へは `follows` で test_plan へ張る

relation は**起こした側（下流）の frontmatter に宣言する**（`sara check` の JSON も
宣言辺だけを持つ）。前工程の item が無いまま下流を起こさない — 無ければ前工程の実行を
提案して停止するのが工程スキルの原則。既存資産から下流を先に起こす逆算は規範外の
移行専用手順で、[references/reverse-import.md](references/reverse-import.md) に隔離する。

## ID 規約（フル UUIDv4）

- ID とファイル名の形式・フル ID を書く場所・散文と `specification` での item への触れ方・ADR の採番は
  [フル ID は frontmatter だけに書き、散文では item にリンクで触れる](../../../docs/quality/full-ids-in-frontmatter-links-in-prose.md)。
- 起票は ADR を含め **`sara-new` コマンド**。採番とファイル名規約の適用、`sara init` の
  呼び出しまでを一手で済ませ、`id:` / `file:` の 2 行を返す。連番を手で振らない
  （並列レーンでの採番衝突を構造的に避けるため）。手順は
  [references/sara-cli.md](references/sara-cli.md)。
- ADR の規則のうち、粒度・本文の構成・`justifies` の選定基準は
  [ADR は 1 本に 1 つの決定を 4 節で書き、justifies は規範の形成に寄与した item へ張る](../../../docs/quality/adr-one-decision-four-sections-justifies-by-contribution.md)、
  ステータス・出典・`supersedes` と `deprecated`・採番は
  [設計判断は ADR に記録し、改訂の書き戻しと item への接続を同じ変更の中で完了させる](../../../docs/quality/adr-decisions-recorded-and-connected.md)。

## risk の評価と優先度の導出

- threatens の張り先は [risk はまず requirement に張り、設計を差し替えると消える懸念だけを design に張る](../../../docs/quality/risk-attaches-to-requirement-by-default.md)。
- likelihood・impact の採点と、`level` と test_condition の優先度の導出は
  [risk の level は likelihood と impact から導き、手で決めない](../../../docs/quality/risk-level-derived-from-likelihood-and-impact.md)。

## specification 規約

`specification`（英語）と `specification_ja`（日本語）の様式は
[specification は SHALL 系で書き、日本語の規範文は同じ強度の規範助動詞で終える](../../../docs/quality/specification-shall-family-and-matching-auxiliaries.md)。

## 機械的検出の使い分け

- **`sara check`** — グラフの整合（broken reference・duplicate ID・cycle・orphan）。
  書き込みのたびに回す。green が完了条件。
- **`sara-gap`** — valid なグラフの**未カバーと到達率**を 7 節で出す。工程の着手点リストとして使う
  （`unthreatened` = リスク識別の未着手（`threatens` も `risk_exempt` も無い `requirement`）、
  `unmitigated` = テスト分析の未着手、`uncovered` = テスト実装の未着手、
  `unfollowed` = `test_plan` への `follows` が無い `test_case`、
  `unsatisfied_quality` = `infrastructure` に `satisfies` されていない `quality`、
  `exempt` = `risk_exempt` を持つ `requirement`（情報）、
  `reach` = `requirement`・`test_plan`・`quality` の到達率）。
  exit 0 = ギャップなしで到達率がすべて 1 / 1 = ギャップありか到達率が 1 未満 / 2 = check 失敗。
- **`sara-fresh`**: valid なグラフから、`generated` と `verified` で見て確かめ直しが要る item を 2 節で出す
  （`stale` = 上流の `generated.at` が自身の基準時刻より新しい item、
  `unverified_high_risk` = `generated.at` 以降の `human:` の `verified` を持たない high の `risk`）。
  上流の item を書き換えた後に、下流で読み直す item を探すのに使う。結果は報告に留め、`verified` は自動で付けない。
  exit 0 = どの節も空 / 1 = いずれかの節に item がある / 2 = check 失敗か frontmatter を読めない item がある。

CLI の詳しい使い方（`sara-new` による起票手順を含む）は
[references/sara-cli.md](references/sara-cli.md)。
