---
id: "ADR-553eea60-6ab4-4da7-93a9-3acf44b2fae3"
type: adr
name: "engine は cmd/ と internal/ に置きライブラリ化しない"
status: 採用
issues:
  - "#3"
origin: "Issue #3（CLI 化）と、engine / CLI の技術スタックの確定。ADR-0006「ソース配置」節・ADR-0007 §7・ADR-0011 §3 を書き直した"
justifies:
  - "REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645"
  - "REQ-79ce0a09-f9bd-4e61-ba7f-45fb5643137b"
  - "REQ-b74a118a-1272-44eb-944c-7725163211c6"
  - "DSG-1361df1a-31c4-46ed-a6ce-92938f94ac02"
  - "DSG-7d354fe0-a333-495b-9f4b-14bba316dc47"
---
# ADR-553eea60-6ab4-4da7-93a9-3acf44b2fae3: engine は cmd/ と internal/ に置きライブラリ化しない

- ステータス: 採用
- 日付: 2026-11-01
- 関連: GitHub Issue #3
- 起点: Issue #3（CLI 化）と、engine / CLI の技術スタックの確定。ADR-0006「ソース配置」節・ADR-0007 §7・ADR-0011 §3 を書き直した

## 背景

layat のバイナリは、entrypoint の発見と manifest の取得を受け持つ CLI と、`manifest.json` を入力に
配置を行う engine の 2 層からなる。engine を「ライブラリ」として切り出すとき、それが外部から import
できる再利用モジュールを指すのか、バイナリの内部の層の分離を指すのかを決める必要があった。

## 決定

engine は `internal/` に置き、外部から import できない内部の層にとどめる。CLI は `cmd/layat` に置いて
engine を import し、両者を 1 つのバイナリとして配布する。外部に対する安定した面は `manifest.json` の
入力契約だけにする。

2 層構成は REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645 に、`manifest.json` が唯一の安定した契約であることは
REQ-79ce0a09-f9bd-4e61-ba7f-45fb5643137b に、engine を内部の層に閉じることは
REQ-b74a118a-1272-44eb-944c-7725163211c6 に、ディレクトリとパッケージの構成は
DSG-1361df1a-31c4-46ed-a6ce-92938f94ac02 と DSG-7d354fe0-a333-495b-9f4b-14bba316dc47 に置く。

## 根拠

層を分ければ、安全に関わる配置と保守的な stale 除去を engine 単体でテストでき、CLI の nix の操作との
境界も明確になる。分離の目的はテストのしやすさで、再利用ではない。`internal/` にすれば外部からの
import を Go のコンパイラが禁じるので、公開 API の semver 互換を保つ負担を負わずに済む。再利用の
要求は無く、安定した面は `manifest.json` の契約で足りる。

## 棄却した代替案

### engine を公開の Go モジュールにする（`pkg/`）

外部から再利用できるが、semver 互換を保ち続ける負担が生じる。安定した面は `manifest.json` の契約で
足り、再利用の要求も無い。

### CLI と engine を別のバイナリに分ける

配布物が増える。`internal/` の層の境界と単一のバイナリで、責務の分離とテストのしやすさは両立する。

### CLI と engine をパッケージの層に分けずに書く

engine が `manifest.json` だけを受け取るという契約を保てず、評価のテストと配置のテストの境界が曖昧になる。
