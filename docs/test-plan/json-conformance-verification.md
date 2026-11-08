---
id: "TP-d3000054-42d9-4bac-912a-dd3abc38d3e9"
type: test_plan
name: "エンベロープの outturn 適合を Go テストと E2E の両方で検証する"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
specification: |
  A Go test SHALL verify the emitted document against the outturn
  `conformance.NewDefaultChecker()` (the embedded canonical schema including format
  assertions, plus the out-of-schema lint MUSTs) for every shape: success, with and
  without a subject, a pre-stage error, a subject error, a conflict, and the multiple
  subjects of `--all` (partial failure, zero results, dryrun conflict). The E2E SHALL
  verify the real output of the scenarios with the `outturn-validate` CLI provided by the
  outturn input of the dev flake (omitting `-schema` selects the embedded canonical
  schema), and SHALL cross-check the item ids against the `id-vectors.json` testdata of
  the same input. Conformance alone SHALL NOT be the whole of the E2E's `--json`
  verification: since the schema constrains the shape of the envelope but not what it
  says, the E2E SHALL additionally assert the meaning of the payload structurally against
  the run that produced it — that the exit code accompanying the document is the expected
  one, that the planned changes carry the kind and reversibility the run implies, that the
  inventory of items and their statuses match, that a failing item carries the error code
  for its cause, and that the generation numbers are present or omitted according to
  whether a profile exists. Those assertions SHALL be made on extracted fields rather than
  by matching the document as text, so that they do not break on formatting or on fields
  they do not concern.
specification_ja: |
  Go テストは、emit した文書を outturn `conformance.NewDefaultChecker()`（embed した正とする
  schema〔format assertion 込み〕+ schema 外 lint MUST）で全形状について検証しなければ
  ならない（成功 / subject あり・なし / 前段エラー / 主体エラー / conflict / `--all` の
  複数 subject〔部分失敗・0 件・dryrun conflict〕）。E2E は dev flake の outturn input が
  提供する `outturn-validate` CLI（`-schema` 省略 = embed した正とする schema）でシナリオの実出力を検証し、
  item id は同 input の `id-vectors.json` testdata と比べなければならない。適合検証
  だけを E2E の `--json` 検証としてはならない。schema はエンベロープの形は縛るがその内容が
  何を述べているかは縛らないため、E2E は加えて、payload の意味をそれを生んだ実行に照らして
  構造的にアサートしなければならない。すなわち、文書に伴う終了コードが期待どおりであること、
  計画された change がその実行の含意する kind と可逆性を持つこと、item のインベントリと
  status が一致すること、失敗した item がその原因に対応するエラーコードを持つこと、そして
  世代番号が profile の有無に応じて存在または省略されること。これらのアサートは文書をテキストで
  比較するのではなく抽出したフィールドに対して行わなければならない（書式や当該アサートの
  関知しないフィールドで失敗しないようにするため）。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T22:44:41+09:00"
---
# TP-d3000054-42d9-4bac-912a-dd3abc38d3e9: エンベロープの outturn 適合を Go テストと E2E の両方で検証する

## 仕様

**適合検証**: Go テストが outturn `conformance.NewDefaultChecker()`（embed した正とする schema
〔format assertion 込み〕+ schema 外 lint MUST）で emit 文書を全形状（成功 / subject
あり・なし / 前段エラー / 主体エラー / conflict / `--all` の複数 subject〔部分失敗・
0 件・dryrun conflict〕）について検証する。E2E は dev flake の outturn input が提供する
`outturn-validate` CLI（`-schema` 省略 = embed した正とする schema）でシナリオ 01–07 の実出力を検証し、
item id は同 input の `testdata/v1/id-vectors.json`（`OUTTURN_ID_VECTORS`）と比べる。

> **上は原文の写しで、規範は frontmatter が正**。原文の「シナリオ 01–07」という
> 具体の本数は E2E 実装の現況であり、規範文では「シナリオの実出力を検証する」に留めた。

**適合検証だけでは E2E の `--json` 検証は足りない**。schema はエンベロープの**形**を縛るが、
その内容が何を述べているかは縛らない。適合していて意味が誤っている出力（実行が add したのに
change が空、conflict なのに status=success 等）は検出されずに通る。そこで E2E は、適合検証に
加えて payload の意味を実行に照らして構造的にアサートする。

| アサート対象 | 内容 |
|---|---|
| 終了コード | 文書に伴う exit code が期待どおり（成功 0 / 失敗 1 / conflict 2）|
| change の意味論 | 計画された change が実行の含意する `kind` と `reversible` を持つ |
| item インベントリ | items の本数と各 `status` が実行と一致する |
| エラーコード | 失敗した item が原因に対応する code を持つ |
| 世代番号 | profile の有無に応じて `generation.before` / `after` が存在または省略される |

これらは文書のテキストの比較ではなく抽出したフィールドに対して行う（書式や当該アサートの
関知しないフィールドの変化で失敗しないようにするため）。

item id の導出規則そのものは REQ-57137302-de29-4f71-a565-034cd5de080b の担当。

## 出典

`docs/spec.md`「CLI 仕様」→「出力ストリームと終了コード」→「outturn 準拠の `--json`
出力」のサブ項目「適合検証」。

決定の実体は ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71、ADR-feb60906-1517-46b8-8fa5-e007df4d3e11、
ADR-79af04fe-7ba0-4953-bb0d-75bda1707cd7、ADR-5d28818e-f7e1-45e8-a7fd-bbb6157b8586、
ADR-c6bc7060-7af4-4889-9ce6-5800601c9dcc、ADR-9e328362-2e90-441a-8eca-40081ce71b1b、
ADR-baf6346d-783b-4607-ab1c-c2c93c09e064、ADR-bd8a4c8a-5e7e-485a-bccf-20dab63125f2、
ADR-64cff3f4-5997-49f4-b23a-e6ebeced7df9、ADR-acc31876-df36-485e-b356-92ae60e26db7。

> **2026-08-09 追補（→ Issue #273）**: E2E の `--json` 検証範囲を追補した。原文が挙げるのは
> `outturn-validate` による適合検証と id-vectors との比較だけだったが、実装（`tests/e2e/
> lib.sh` の `run_json` / `assert_json` と各シナリオ）は終了コードの一致に加え、change の
> 意味論・item インベントリ・エラーコード・世代番号を jq で構造的にアサートしている。適合
> 検証だけを規範に置くと、この層が退行しても item に照らして検出できないため規範へ引き上げた。

> **本 item は `requirement` から `test_plan` へ移設した**（→ Issue #238。旧 ID は
> `REQ-2381d93a-732e-4437-910b-fac14d398aa0`）。エンベロープを outturn 規約準拠にすること自体は REQ-a5053191-1c6a-449b-9c5e-5ff49dc5aead が担い、
> そちらは use_case へ紐づく。本 item はその適合をどのテストレベルでどう検証するかを定める
> テストアプローチの規定であり、ユーザーの使われ方からは導かれないため、use_case を親に
> 持てず接続漏れになっていた（当時の判断は Issue #211）。テスト計画の型を新設して solution
> 直下で受けることにしたため、`derives_from` は SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e を指す。TP-229b69c0-cf5e-4fb6-a353-27e5064d93e9・
> TP-b7f1dc79-0222-4b6e-9e91-0545046e34f2 も同じ経緯で移設した。TP-403c55c7-d996-4951-8e6b-c3a7dddd387c は #238 では「テスト計画そのものではない」
> と判断して見送ったが、Issue #239 でその判断を改めて移設した（経緯は同 item の注記）。
