---
id: "CASE-932e218b-5bfb-4d66-98a3-41054c0d863f"
type: test_case
name: "outturn_test.go — エンベロープ基礎形・id 導出・エラー分類・stdout 所有権"
status: stable
resource: "cmd/layat/outturn_test.go"
covers:
  - "TC-1a46da17-1812-431b-966f-ec8dffd315ae"
  - "TC-733ac4ed-dac3-4bb8-9bfd-fd1cbdc300c9"
  - "TC-b4a365cd-0b1c-40ff-92c6-91f6cffe8e98"
  - "TC-b0e0cdb8-0c56-4eb8-8d4e-86569af85dec"
  - "TC-0c7281f8-cffc-4a53-8fe0-f52d9ea6bd12"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:34+09:00"
---
# CASE-932e218b-5bfb-4d66-98a3-41054c0d863f: outturn_test.go — エンベロープ基礎形・id 導出・エラー分類・stdout 所有権

## 対象

`cmd/layat/outturn_test.go`

## 検証内容

- **エンベロープ基礎形の適合** — subject の有無 × 成功 / 失敗、主体確立前の失敗と主体起因の
  失敗を表で網羅し、ツール名とバージョン・コマンド名・dryrun フラグが載ること。stdout に
  JSON 1 文書と改行だけが出ること
- **item id の他実装互換** — outturn が配布する id ベクタを全件再現すること、entry 用の
  identity の形が同じ id を導くこと
- **タイムスタンプの書式** — 日付と時刻を `T` で区切り、オフセットを明示して描画すること
  （UTC 以外の時間帯は `+09:00` のような明示オフセット、UTC は `Z`）。適合チェッカは書式が
  妥当かは見るが、どの時間帯をどう描画したかまでは区別しない
- **フラグの登録** — `--json` が最上位の persistent flag として登録され、短縮形を持たないこと
- **エラー分類** — prebuilt 以外の生成器の roots / build 段の失敗が `E_LAYAT_BUILD` になり、
  原因が not-found でもそれに優先すること・そのマーカーが再ラップされても残ること。
  discover 段と prebuilt の失敗は原因の error チェーンから従来の規則で分類すること（not-found
  の原因は `E_NOTFOUND`、原因が無ければ汎用分類・→ [生成器の失敗を --json のコード E_LAYAT_BUILD 1 本に分類する](../../adr/generator-failure-single-build-code.md)）。入力不正のマーカーが
  `E_INPUT` になり not-found の原因に優先すること（未知の `--generator` / `LAYAT_GENERATOR` の値、
  `--manifest` と `-f` の併用、`--all` と `<link-farm>` 形の `--manifest` の併用・→ [入力の不正を E_INPUT に分類する](../../adr/input-errors-classified-e-input.md)）。ファイルシステム由来の sentinel が汎用の
  形状判定に優先すること、排他制御由来の分類、いずれにも当たらない失敗の汎用分類
- **`--json` 下の stdout 所有権** — 行指向の出力（計画・世代一覧・パス列挙）が全て止まり、
  既定の契約では出ること。stderr の人間向け通知は両契約で維持されること
- **run の開始と公開の網羅** — 各コマンドが run を開始して公開すること、登録済みの実行系
  サブコマンドのどれにもその配線があること、help / completion は run を開始せず文書も出さない
  こと
- **確認の合成** — 機械可読出力の指定が端末でもプロンプトを禁じ、同意フラグの無い実行が
  即座に止まること
- **出力の書き込み失敗** — 書き込みに失敗する出力先で失敗が伝播すること
- **実コマンドの処理の端から端まで** — 一時ディレクトリで実際にコマンドを実行し、entrypoint が
  無いことによる実際の失敗を起こして、run が公開されること・実際の stdout をエンベロープが
  専有すること・適合を保つことを通しで確かめること（構築済みの run に対する表駆動の検証とは
  別の層）

表駆動の検証は固定クロックとバッファを用い、時刻とホスト環境に依存しない（上の実コマンドの
処理だけは実際の stdout と一時ディレクトリを使う）。他の CASE が使う run のコンストラクタと
デコーダをこのファイルが定義する。
