---
id: "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
type: infrastructure
name: "トレーサビリティ検証基盤（sara によるドキュメントグラフの機械検証）"
status: stable
depends_on:
  - "INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27"
satisfies:
  - "QA-6bf957d9-17d9-4660-92b7-ebd6eeb71a8c"
  - "QA-b14e0a14-b9c9-4cc6-9b7a-0ce95d30d289"
  - "QA-bf975ac7-1e84-4f75-9fe7-8d6154d69ba1"
  - "QA-1dd63789-24ed-4a86-bf12-b62d8c526ba5"
  - "QA-611977bd-3d11-4695-9fb2-a9f7211fcfbd"
  - "QA-aa70fff6-3e0a-4163-bf4c-97f1730f7496"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-13T14:33:07+09:00"
---
# トレーサビリティ検証基盤（sara によるドキュメントグラフの機械検証）

## 構成

`docs/` を [sara](https://github.com/cledouarec/sara)（Rust 製 CLI。個人 NUR 経由で dev flake
へ入れる）のナレッジグラフとして扱い、要求・設計・リスク・テストの紐付けが一貫しているかを
機械検証する。要求とリスクの紐付けの一貫性管理が主目的で、「仕様とテストの全体像把握」を支える。

| 部品 | 役割 |
|---|---|
| `docs/model.yaml` | カスタムスキーマ。13 型と relation を定義し、sara の組み込みモデルを全面置換する |
| `sara.toml` | リポジトリ設定 |
| `sara` ジョブ | `test.yml` 内の 1 ジョブ。`sara check` で broken reference / duplicate ID / cycles を検出し、同じジョブで契約テスト（この表の `dev/tests/` の 10 件）を実行し、`sara-fresh` で実物のグラフの一覧を出す |
| `dev/tests/test-doc-map.sh` | テスト資産 ⇔ CASE の 1:1 対応を検証する（→ Issue #304）。区分と除外の正本は `dev/tests/` のデータファイルが持つ |
| `dev/tests/risk-matrix.sh` | risk の `level` が `likelihood` × `impact` のマトリクス導出と一致するかを検証する（→ Issue #303）。マトリクスの正本は `dev/tests/risk-matrix.tsv` |
| `dev/tests/sara-gap.sh` | `sara-gap` が出す未カバーの 7 節と到達率の契約、および exit code を検証する。判定は fixture 基準で、実グラフのどの item がギャップかは見ない（件数が動き続けるため契約にできない）|
| `sara-fresh` | `dev/scripts/sara-fresh.sh` を包む devShell のコマンド。上流の `generated.at` が自身の基準時刻より新しい item、`--base <ref>` との差分で本文が変わったのに `generated.at` が変わっていない item、`generated.at` 以降の `human:` の `verified` を持たない high の risk を列挙する。結果は報告に留める |
| `dev/tests/sara-fresh.sh` | `sara-fresh` の 3 節の列挙と exit code の契約を検証する。グラフはテストが一時ディレクトリに作り、実物の `docs/model.yaml` を重ねる。`--base` の節は、一時ディレクトリに `git init` したリポジトリで確かめる |
| `dev/tests/sara-new.sh` | `sara-new` の起票契約（採番・ファイル名規約の適用）を検証する（→ Issue #367）。fixture は実物の `docs/model.yaml` を重ねて使う |
| `dev/tests/adr-supersede.sh` | `supersedes` が指す ADR が `status: deprecated` かを frontmatter から検証する |
| `dev/tests/okf-frontmatter.sh` | OKF 互換の入れ子キー（`generated` / `verified` / `sources`）と `stale_after` の形を frontmatter から検証する |
| `dev/tests/okf-index.sh` | `index.md` が `dev/scripts/okf-index.sh` の生成結果と一致するかを検証する |
| `dev/tests/term-avoid.sh` | 用語 item の `avoid` にある綴りが docs・README・CLAUDE.md の散文に無いことを検証する |
| `dev/tests/link-text.sh` | リンク文字列の裸のタグと、リンク直後の name の繰り返しを検出する。文書とコードにある許可外の場所のフル ID と 8 文字省略形も検出する |

契約テストを同じジョブへ載せるのは、テストが「人が思い出したときだけ動く」状態を避けるため。
表の `dev/tests/` の 11 件のうち `adr-supersede.sh` を除く 10 件は dev flake の `checks.*` としても露出しているが、CI の `flake-check` はルート
flake を対象にするため到達しない。CI の明示ステップとして固定する。

`sara` ジョブは契約テストに続けて、実物のグラフに `sara-fresh` を実行し、一覧をログとジョブのサマリーへ出す。
pull_request では base ブランチを `--base` に渡し、workflow_dispatch では `--base` を付けない。merge-base を引けるよう、
このジョブの checkout は全履歴を取る。挙がった item は読み直しが要ることを示すだけで破損を意味しないため、
一覧に item が挙がっても（exit 1）ジョブは落とさず、実行に失敗したとき（exit 2）だけ落とす
（→ [上流より古い item と generated の更新漏れは機械的に列挙する](../quality/stale-items-enumerated-mechanically.md)）。

`sara` ジョブは `test.yml` に同居するが、[CI パイプライン（flake check の os×system マトリクスと E2E ジョブ）](ci-pipeline.md) の変更検出ジョブは
**再利用しない**。あちらの filter は nix / Go のソースを対象にするのに対し、この確認が最も役立つのは
docs のみの PR だからで、job ローカルの filter を別に置く。

filter が対象にするのは 4 種類。確認の対象の文書とその設定（`docs/**` / `sara.toml`）、この確認自身の
実行環境を決めるもの（sara 本体と dev ツールを供給する dev flake・その lock・契約テストの実装＝
`dev/tests/**` とそれが依存する `dev/scripts/**`・この workflow 自身）、そして**契約テストが
入力に取るテスト資産そのもの**（Go のテストファイル・`tests/**`・`flake.nix` の checks 定義）。
3 つ目は `test-doc-map.sh` が「テスト資産 ⇔ CASE の 1:1」を確認する以上避けられない — テスト
ファイルを追加・改名しても CASE を更新しなければジョブごとスキップされ、検出したいはずの
乖離がそのままマージされてしまう（→ Issue #304）。
4 つ目は `link-text.sh` が ID を走査するコードと設定（`internal/**`・`cmd/**`・`lib/**`・`modules/**` など）で、
これが無いとコードだけの PR で許可外の ID を見逃す。
CI 用の devShell も sara 専用のものへ分ける。

先行して試作された grep / awk による ID の比較は、参照先 ID の実在を検証しないという不足を抱えていた
（存在しない `TC-999` を書いても検出されずカバー率に計上される）。sara の broken reference 検出が
この不足を補う。

## ID 規約

→ [フル ID は frontmatter だけに書き、散文では item にリンクで触れる](../quality/full-ids-in-frontmatter-links-in-prose.md)

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

[model.yaml で組み込みモデルを全面置換し、汎用の型名を使う](../sara-adr/model-yaml-full-replacement.md)、
[sara 品質管理機構の決定は sara_adr 型に分ける](../sara-adr/sara-adr-type-for-doc-process-decisions.md)、[ADR も UUID で採番し、連番の旧 ADR は legacy_adr 型に凍結する](../sara-adr/adr-uuid-ids-and-legacy-type.md)、
[root は solution と ADR 系の型だけにする](../sara-adr/roots-solution-and-adr-types-only.md)、[sara check を strict モードで回す](../sara-adr/sara-check-strict-mode.md)、
[risk の親に test_plan を含めない](../sara-adr/risk-parent-excludes-test-plan.md)、
[infrastructure を技術基盤として再定義し root から外す](../sara-adr/infrastructure-redefined-non-root.md)、[infrastructure 型を改名しない](../sara-adr/infrastructure-name-kept.md)、
[design の satisfies 対象に test_plan を含める](../sara-adr/design-satisfies-test-plan.md)、[ADR の justifies 対象を specification を持つ全型に広げる](../sara-adr/adr-justifies-quality-and-test-plan.md)、
[item にできない外部参照は text フィールドへ逃がす](../sara-adr/external-references-as-text-fields.md)、
[提案を棄却した決定は status 不採用 で記録する](../sara-adr/adr-status-rejected.md)、[adr の status_note を持たない](../sara-adr/adr-status-note-removed.md)、
[ADR の失効は supersedes と 廃止 で表す](../sara-adr/adr-supersedes-and-deprecated-status.md)、[ID はフル UUID に統一し 8 文字省略形を使わない](../sara-adr/full-uuid-ids-no-short-form.md)、
[sara check を required status check にしない](../sara-adr/sara-check-not-required.md)、[frontmatter の外のフル ID と省略形は起票の機械化と契約テストで防ぐ](../sara-adr/frontmatter-only-ids-checked-by-contract-test.md)。設計は epic GitHub Issue #203 の grilling
セッション（2026-07-31・2026-08-01）で確定した。
