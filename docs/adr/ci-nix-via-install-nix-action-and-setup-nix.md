---
id: "ADR-d1180af4-49b6-4d46-9575-7a9790339d17"
type: adr
name: "CI の nix は install-nix-action で入れ、setup-nix composite action にまとめる"
description: "CI の nix を cryoflow にそろえて cachix/install-nix-action で入れ、cachix/cachix-action の設定とあわせて composite action .github/actions/setup-nix にまとめ、nix を使う全 workflow から再利用する。"
section: ci-release-docs
status: stable
origin: "[CI の nix インストーラは cryoflow 踏襲](ci-nix-installer-cryoflow.md) を書き直した（外部 action の固定は別 ADR へ分けた）"
justifies:
  - "INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27"
supersedes:
  - "ADR-ab400e58-afbd-40a0-8eae-239d3cfda211"
references:
  - "ADR-a41e979b-701d-45b5-b371-8dbbcab82833"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-13T00:09:54+09:00"
---

# CI の nix は install-nix-action で入れ、setup-nix composite action にまとめる

- ステータス: stable
- 日付: 2026-11-13
- 関連: [CI の外部 action は actions.lock で固定し、uses には semver タグを書く](ci-actions-pinned-by-actions-lock.md)
- 起点: [CI の nix インストーラは cryoflow 踏襲](ci-nix-installer-cryoflow.md) を書き直した（外部 action の固定は別 ADR へ分けた）。参照: https://github.com/yasunori0418/cryoflow の CI 構成（同一メンテナの実績構成）

## 背景

CI で nix を使うには、ランナーへ nix を入れる手段が要る。同じメンテナの cryoflow はすでに
Nix ベースの CI を運用しており、インストーラとキャッシュの設定を composite action にまとめている。

## 決定

CI の nix は `cachix/install-nix-action` で入れる。インストールと `cachix/cachix-action`
（キャッシュ `yasunori0418`）の設定を composite action `.github/actions/setup-nix` にまとめ、
nix を使う全ての workflow から再利用する。

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
