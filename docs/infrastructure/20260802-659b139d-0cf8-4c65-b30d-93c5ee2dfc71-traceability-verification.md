---
id: "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
type: infrastructure
name: "トレーサビリティ検証基盤（sara によるドキュメントグラフの機械検証）"
depends_on:
  - "INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27"
satisfies:
  - "QA-6bf957d9-17d9-4660-92b7-ebd6eeb71a8c"
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

`sara` ジョブは `test.yml` に同居するが、CI パイプライン（INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27）の変更検出ジョブは
**再利用しない**。あちらの filter は nix / Go のソースを対象にするのに対し、この検査が最も効くのは
docs のみの PR だからで、job ローカルの filter を別に置く。

filter が拾うのは 3 種類。検査対象の文書とその設定（`docs/**` / `sara.toml`）、この検査自身の
実行環境を決めるもの（sara 本体と dev ツールを供給する dev flake・その lock・契約テストの実装＝
`dev/tests/**` とそれが依存する `dev/scripts/**`・この workflow 自身）、そして**契約テストが
入力に取るテスト資産そのもの**（Go のテストファイル・`tests/**`・`flake.nix` の checks 定義）。
3 つ目は `test-doc-map.sh` が「テスト資産 ⇔ CASE の 1:1」を検査する以上避けられない — テスト
ファイルを追加・改名しても CASE を更新しなければジョブごとスキップされ、検出したいはずの
乖離がそのままマージされてしまう（→ Issue #304）。
CI 用の devShell も sara 専用のものへ分ける。

先行して試作された grep / awk による ID 突合は、参照先 ID の実在を検証しないという穴を抱えていた
（存在しない `TC-999` を書いても検出されずカバー率に計上される）。sara の broken reference 検出が
この穴を塞ぐ。

## ID 規約

正式 ID は `<PREFIX>-<フル UUIDv4>`。ファイル名・散文中の参照を含め**どの面でもフル ID を
そのまま書く**（→ SDR-722ce265-2aaf-4a90-915d-32d26b7a38bb）。乱数 ID により並列レーンでの採番衝突を構造的に回避する。
表記が全面で一致するので、目に付いた ID をそのまま grep しても `sara query` へ渡しても
同じものに当たる。

ADR も同じ形式で採番する（`ADR-<フル UUIDv4>`・`SDR-<フル UUIDv4>`）。連番の `ADR-NNNN` は
削除済みで、新 ADR の `origin` と「起点:」行に来歴として残る。

## 検証の強度

**`strict_mode = true` の非必須チェック**。検査自体は strict で、orphan を含む全 warning が
error となり `sara check` は exit 1 を返す（epic #203 の移行完了・warning 0 到達を受けて有効化
→ SDR-bf25a81b-d588-4c7e-9365-fe514ee4d626）。合格条件は「warning 0 件」に一本化されている。

一方でマージゲート（INF-8b97573f-d4d6-4abf-85e2-d859afbd96c6）の required status check には登録せず、DoD にも入れない。
これは移行中の暫定措置ではなく「**検出は機械・判断は人**」という分担を保つための恒久判断で、
接続漏れの修正が自明でないケース（item 新設を伴う等）を CI が赤いままレビューで扱える余地を
残す。再検討条件は廃止した旧 ADR が持っていた（現在の決定は → SDR-6cc54665-c239-4235-92a8-f73e26c5a5ef）。

閾値ゲートを持たない点は CI パイプライン（INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27）の `go-coverage` ジョブと同方針で、
他 PR とのマージ順依存を避けるためでもある。

## 出典

SDR-5ea31001-96f8-4324-a40b-647cdfa5cc9d、SDR-536848bb-8901-4e73-8c66-861de3f91bcd、
SDR-2457aa00-fab7-4c99-af53-2e6643c21c79、SDR-49733f2e-ed79-480d-bfb9-6361a7f95aed、
SDR-c15921a9-7aea-407c-a34b-7465783b1af1、SDR-e093c6b2-75fd-4c91-a8f0-5a89e005e4ed、
SDR-a0095ee8-570c-4f77-a6c9-bbc162a5c4cb、SDR-bf25a81b-d588-4c7e-9365-fe514ee4d626、
SDR-34916a83-0cff-4524-a8ff-5a3195af52c3、SDR-f3f9491f-d6f4-4742-82da-4933ad3d8db1、
SDR-29ef8bb4-9349-47db-9e79-926bc71bbdd0、SDR-933a9191-f60e-45bc-bebc-bc09af8154cc、
SDR-35636015-710a-4cf8-b6ba-48ca047efed0、SDR-1d295034-9bab-4e30-9ead-173e5f56c396、
SDR-9bec5976-67cd-407a-afa9-3d394a6204ca、SDR-9f2ef734-a9a7-4b93-92a7-e8ed1c7cf8fe、
SDR-d5304533-00ae-423b-a459-7c67c1cf093c、SDR-fa7eff8e-788f-48e7-951d-4cfbf95d5a1f、
SDR-3ff9f355-a667-476e-b10d-087248cdbfcf、SDR-722ce265-2aaf-4a90-915d-32d26b7a38bb、
SDR-6cc54665-c239-4235-92a8-f73e26c5a5ef。設計は epic GitHub Issue #203 の grilling
セッション（2026-07-31・2026-08-01）で確定した。
