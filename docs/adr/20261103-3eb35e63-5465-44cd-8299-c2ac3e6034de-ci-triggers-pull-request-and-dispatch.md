---
id: "ADR-3eb35e63-5465-44cd-8299-c2ac3e6034de"
type: adr
name: "テスト系の CI を pull_request と workflow_dispatch で起動する"
status: 採用
origin: "参照: PR #40、https://github.com/yasunori0418/cryoflow の CI 構成。ADR-0027 §2a（トリガは pull_request と workflow_dispatch）を書き直した"
justifies:
  - "INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27"
---
# ADR-3eb35e63-5465-44cd-8299-c2ac3e6034de: テスト系の CI を pull_request と workflow_dispatch で起動する

- ステータス: 採用
- 日付: 2026-11-03
- 起点: 参照: PR #40、https://github.com/yasunori0418/cryoflow の CI 構成。ADR-0027 §2a（トリガは pull_request と workflow_dispatch）を書き直した

## 背景

`test.yml` には、flake check や E2E のようなテスト系のジョブと、テスト資産とテスト文書の
対応表を作る `test-doc-matrix` が同居している。main への変更は PR 経由でしか入らない。
どのイベントでどのジョブを回すかを決める必要がある。

## 決定

テスト系のジョブは `pull_request` と `workflow_dispatch` で起動し、main への push では回さない。
`test.yml` は `push: main` もトリガに持つが、それを使うのは `test-doc-matrix` だけで、他のジョブは
ジョブ段のガードで push のイベントを除外する。`test-doc-matrix` は、生成の入力を変える PR では
ステップ段の filter によって PR の上でも実行する。

トリガとジョブ段のガードの細目は INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27 に置く。

## 根拠

main に入る変更は PR のチェックを通っている。テスト系のジョブを main への push でもう一度回すと、
PR のチェックと重複して CI の時間を浪費する。手動の再実行は `workflow_dispatch` で足りる。

`test-doc-matrix` は全体のインベントリを要し、PR ごとに払うには重い。マージ後の状態から作れば
足りるので、main への push と `workflow_dispatch` に負担させる。GitHub Actions には push の対象
ジョブをトリガ段で絞る手段が無いため、除外はジョブ段で書く。

## 棄却した代替案

### `push`（branches-ignore: main）をトリガに併用する

PR のチェックと重複して CI の時間を浪費する。

### `test-doc-matrix` を全ての PR で回す

対応表の生成は重く、入力を変えない PR で払う価値が無い。
