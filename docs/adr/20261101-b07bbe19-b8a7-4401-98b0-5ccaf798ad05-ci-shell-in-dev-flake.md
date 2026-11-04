---
id: "ADR-b07bbe19-b8a7-4401-98b0-5ccaf798ad05"
type: adr
name: "CI シェルは dev/flake.nix に公開する"
status: 採用
origin: "参照: https://github.com/yasunori0418/cryoflow の CI 構成（同一メンテナの実績構成）。ADR-0012 §2 を書き直した"
justifies:
  - "INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27"
---
# ADR-b07bbe19-b8a7-4401-98b0-5ccaf798ad05: CI シェルは dev/flake.nix に公開する

- ステータス: 採用
- 日付: 2026-11-01
- 起点: 参照: https://github.com/yasunori0418/cryoflow の CI 構成（同一メンテナの実績構成）。ADR-0012 §2 を書き直した

## 背景

CI の E2E ハーネスは、layat・git・jq・`outturn-validate` などのツールを必要とする。これらをジョブ
ごとに用意すると、手元の開発環境と版がずれうる。リポジトリは成果物を root の `flake.nix` に、
開発用の環境を `dev/flake.nix` に分けている。

## 決定

`dev/flake.nix` に CI 用の devShell `ci` を置き、E2E ハーネスが使うツールだけを入れた最小の
環境を提供する。CI はハーネスを `nix develop '.?dir=dev#ci' -c <cmd>` で実行する。

CI のジョブ構成は INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27 に置く。

## 根拠

CI と手元が同じ flake からツールを得るので、版がずれない。CI 用の devShell を `dev/flake.nix` に
置けば、開発だけで使う依存が root の flake の成果物に含まれない。同じメンテナの cryoflow と
同じ構成で、保守のやり方もそろう。

## 棄却した代替案

旧 ADR-0012 §2 は、この決定について代替案を記録していない。
