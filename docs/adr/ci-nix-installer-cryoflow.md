---
id: "ADR-ab400e58-afbd-40a0-8eae-239d3cfda211"
type: adr
name: "CI の nix インストーラは cryoflow 踏襲"
section: ci-release-docs
status: deprecated
origin: "参照: https://github.com/yasunori0418/cryoflow の CI 構成（同一メンテナの実績構成）。ADR-0012 §1 を書き直した"
justifies:
  - "INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-13T00:10:58+09:00"
---
# CI の nix インストーラは cryoflow 踏襲

> **廃止（2026-11-13）**: 本 ADR の決定は次の後継へ移った。以下の本文は当時の記録として読む。
>
> - nix のインストーラと setup-nix composite action → [CI の nix は install-nix-action で入れ、setup-nix composite action にまとめる](ci-nix-via-install-nix-action-and-setup-nix.md)
> - 外部 action の固定 → [CI の外部 action は actions.lock で固定し、uses には semver タグを書く](ci-actions-pinned-by-actions-lock.md)

- ステータス: deprecated
- 日付: 2026-11-01
- 起点: 参照: https://github.com/yasunori0418/cryoflow の CI 構成（同一メンテナの実績構成）。ADR-0012 §1 を書き直した

## 背景

CI で nix を使うには、ランナーへ nix を入れる手段が要る。同じメンテナの cryoflow はすでに
Nix ベースの CI を運用しており、インストーラとキャッシュの設定を composite action にまとめている。

## 決定

CI の nix は `cachix/install-nix-action` で入れる。インストールと `cachix/cachix-action`
（キャッシュ `yasunori0418`）の設定を composite action `.github/actions/setup-nix` にまとめ、
nix を使う全ての workflow から再利用する。外部 action はコミット SHA で固定する。

CI のジョブ構成は [CI パイプライン（flake check の os×system マトリクスと E2E ジョブ）](../infrastructure/ci-pipeline.md) に置く。

## 根拠

同じメンテナの実績ある構成にそろえると、CI の保守のコストと再現性がそろう。非 NixOS の
ランナーに nix を入れるこの方法が、そのまま「非 NixOS でも nix があれば動く」を確かめる E2E の
実行環境になる。

## 棄却した代替案

### DeterminateSystems/nix-installer を使う

動作はするが、cryoflow が `cachix/install-nix-action` でそろっており、メンテナのツールチェインを
統一する方が保守しやすい。

### `pkgs.dockerTools` で nix 入りのテスト用イメージを自作する

イメージの中で nix を動かすには store・daemon・sandbox の配管が要り、`cachix/install-nix-action` が
提供する機能の作り直しになる。結局 CI のジョブで動かす必要もあり、検証したいのはイメージ構築の
再現性ではない。
