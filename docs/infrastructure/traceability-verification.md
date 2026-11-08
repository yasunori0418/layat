---
id: "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
type: infrastructure
name: "トレーサビリティ検証基盤（sara によるドキュメントグラフの機械検証）"
status: stable
depends_on:
  - "INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27"
satisfies:
  - "QA-6bf957d9-17d9-4660-92b7-ebd6eeb71a8c"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:08:50+09:00"
---
# INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71: トレーサビリティ検証基盤（sara によるドキュメントグラフの機械検証）

## 構成

`docs/` を [sara](https://github.com/cledouarec/sara)（Rust 製 CLI。個人 NUR 経由で dev flake
へ入れる）のナレッジグラフとして扱い、要求・設計・リスク・テストの紐付けが一貫しているかを
機械検証する。要求とリスクの紐付けの一貫性管理が主目的で、「仕様とテストの全体像把握」を支える。

| 部品 | 役割 |
|---|---|
| `docs/model.yaml` | カスタムスキーマ。11 型と relation を定義し、sara の組み込みモデルを全面置換する |
| `sara.toml` | リポジトリ設定 |
| `sara` ジョブ | `test.yml` 内の 1 ジョブ。`sara check` で broken reference / duplicate ID / cycles を検出し、同じジョブで契約テスト 4 件を実行する |
| `dev/tests/test-doc-map.sh` | テスト資産 ⇔ CASE の 1:1 対応を検証する（→ Issue #304）。区分と除外の正本は `dev/tests/` のデータファイルが持つ |
| `dev/tests/risk-matrix.sh` | risk の `level` が `likelihood` × `impact` のマトリクス導出と一致するかを検証する（→ Issue #303）。マトリクスの正本は `dev/tests/risk-matrix.tsv` |
| `dev/tests/sara-gap.sh` | `sara-gap` の未カバー 3 段の列挙契約を検証する。判定は fixture 基準で、実グラフのどの item がギャップかは見ない（件数が動き続けるため契約にできない）|
| `dev/tests/sara-new.sh` | `sara-new` の起票契約（採番・ファイル名規約の適用）を検証する（→ Issue #367）。fixture は実物の `docs/model.yaml` を重ねて使う |

契約テストを同じジョブへ載せるのは、テストが「人が思い出したときだけ動く」状態を避けるため。
4 件はいずれも dev flake の `checks.*` としても露出しているが、CI の `flake-check` はルート
flake を対象にするため到達しない。CI の明示ステップとして固定する。

`sara` ジョブは `test.yml` に同居するが、[CI パイプライン（flake check の os×system マトリクスと E2E ジョブ）](ci-pipeline.md) の変更検出ジョブは
**再利用しない**。あちらの filter は nix / Go のソースを対象にするのに対し、この確認が最も役立つのは
docs のみの PR だからで、job ローカルの filter を別に置く。

filter が対象にするのは 3 種類。確認の対象の文書とその設定（`docs/**` / `sara.toml`）、この確認自身の
実行環境を決めるもの（sara 本体と dev ツールを供給する dev flake・その lock・契約テストの実装＝
`dev/tests/**` とそれが依存する `dev/scripts/**`・この workflow 自身）、そして**契約テストが
入力に取るテスト資産そのもの**（Go のテストファイル・`tests/**`・`flake.nix` の checks 定義）。
3 つ目は `test-doc-map.sh` が「テスト資産 ⇔ CASE の 1:1」を確認する以上避けられない — テスト
ファイルを追加・改名しても CASE を更新しなければジョブごとスキップされ、検出したいはずの
乖離がそのままマージされてしまう（→ Issue #304）。
CI 用の devShell も sara 専用のものへ分ける。

先行して試作された grep / awk による ID の比較は、参照先 ID の実在を検証しないという不足を抱えていた
（存在しない `TC-999` を書いても検出されずカバー率に計上される）。sara の broken reference 検出が
この不足を補う。

## ID 規約

正式 ID は `<PREFIX>-<フル UUIDv4>`。**フル ID を書くのは frontmatter の `id:`・relation リストと、
item 本文の H1 見出しだけにする**（→ [フル ID は frontmatter と H1 見出しだけに書き、item にはリンクで触れる](../sara-adr/full-ids-in-frontmatter-and-h1.md)）。
散文で item に触れるときは Markdown リンクにし、リンク文字列にも ID を書かない。乱数 ID により並列レーンでの採番衝突を構造的に回避する。
ID をそのまま grep すれば frontmatter の `id:` と relation に当たり、`sara query` へ渡しても同じものに当たる。

`specification` / `specification_ja` / `origin` の文字列の中で item に触れるときも、
`[item の name](item のディレクトリからの相対パス)` の Markdown リンクにする。
`name` と H1 に `<name>` のような `<…>` を含む語があればコードスパンに入れる。Markdown でタグに解釈され、描画で消えるため。
散文ではリンクの直後に同じ name を鉤括弧や丸括弧で繰り返さない。

ADR も同じ形式で採番する（`ADR-<フル UUIDv4>`・`SDR-<フル UUIDv4>`）。連番の `ADR-NNNN` は
削除済みで、新 ADR の `origin` と「起点:」行に来歴として残る。

## 検証の強度

**`strict_mode = true` の非必須の確認**。確認自体は strict で、接続漏れを含む全 warning が
error となり `sara check` は exit 1 を返す（epic #203 の移行完了・warning 0 到達を受けて有効化
→ [sara check を strict モードで回す](../sara-adr/sara-check-strict-mode.md)）。合格条件は「warning 0 件」に一本化されている。

一方で [マージの条件（main の ruleset と required status checks）](merge-gate.md) の required status check には登録せず、DoD にも入れない。
これは移行中の暫定措置ではなく「**検出は機械・判断は人**」という分担を保つための恒久判断で、
張り忘れの修正が自明でないケース（item 新設を伴う等）を CI が赤いままレビューで扱える余地を
残す。再検討条件は廃止した旧 ADR が持っていた（現在の決定は → [sara check を required status check にしない](../sara-adr/sara-check-not-required.md)）。

閾値による条件を持たない点は [CI パイプライン（flake check の os×system マトリクスと E2E ジョブ）](ci-pipeline.md) の `go-coverage` ジョブと同方針で、
他 PR とのマージ順依存を避けるためでもある。

## 出典

[ドキュメントは sara のナレッジグラフとして item に分割する](../sara-adr/docs-as-sara-knowledge-graph.md)、[概要文書は item へのリンク集に縮退させる](../sara-adr/overview-docs-as-link-index.md)、
[model.yaml で組み込みモデルを全面置換し、汎用の型名を使う](../sara-adr/model-yaml-full-replacement.md)、[欠陥はグラフに置かず GitHub Issues で管理する](../sara-adr/defects-in-github-issues.md)、
[sara 品質管理機構の決定は sara_adr 型に分ける](../sara-adr/sara-adr-type-for-doc-process-decisions.md)、[ADR も UUID で採番し、連番の旧 ADR は legacy_adr 型に凍結する](../sara-adr/adr-uuid-ids-and-legacy-type.md)、
[root は solution と ADR 系の型だけにする](../sara-adr/roots-solution-and-adr-types-only.md)、[sara check を strict モードで回す](../sara-adr/sara-check-strict-mode.md)、
[テスト計画の型を品質方針の型から分けて設ける](../sara-adr/quality-and-test-plan-types.md)、[risk の親に test_plan を含めない](../sara-adr/risk-parent-excludes-test-plan.md)、
[infrastructure を技術基盤として再定義し root から外す](../sara-adr/infrastructure-redefined-non-root.md)、[infrastructure 型を改名しない](../sara-adr/infrastructure-name-kept.md)、
[design の satisfies 対象に test_plan を含める](../sara-adr/design-satisfies-test-plan.md)、[ADR の justifies 対象を specification を持つ全型に広げる](../sara-adr/adr-justifies-quality-and-test-plan.md)、
[欠陥 issue のクローズ時にテスト系 item の欠落を追補する](../sara-adr/defect-close-backfills-risks.md)、[item にできない外部参照は text フィールドへ逃がす](../sara-adr/external-references-as-text-fields.md)、
[提案を棄却した決定は status 不採用 で記録する](../sara-adr/adr-status-rejected.md)、[adr の status_note を持たない](../sara-adr/adr-status-note-removed.md)、
[ADR の失効は supersedes と 廃止 で表す](../sara-adr/adr-supersedes-and-deprecated-status.md)、[ID はフル UUID に統一し 8 文字省略形を使わない](../sara-adr/full-uuid-ids-no-short-form.md)、
[sara check を required status check にしない](../sara-adr/sara-check-not-required.md)、[フル ID は frontmatter と H1 見出しだけに書き、item にはリンクで触れる](../sara-adr/full-ids-in-frontmatter-and-h1.md)。設計は epic GitHub Issue #203 の grilling
セッション（2026-07-31・2026-08-01）で確定した。
