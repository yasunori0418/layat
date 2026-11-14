---
id: "ADR-a41e979b-701d-45b5-b371-8dbbcab82833"
type: adr
name: "CI の外部 action は actions.lock で固定し、uses には semver タグを書く"
description: "CI が参照する外部 action は gh actions-lock が生成する actions.lock で推移的な依存までコミット SHA に固定し、uses には完全な semver タグを書き、lockfile の整合を verify-lockfile.yml で検証する。"
section: ci-release-docs
status: stable
origin: "[CI の nix インストーラは cryoflow 踏襲](ci-nix-installer-cryoflow.md) の「外部 action はコミット SHA で固定する」を書き直した"
justifies:
  - "INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27"
  - "INF-9878e9f5-1ec0-4ef2-879b-859ea538cc43"
supersedes:
  - "ADR-ab400e58-afbd-40a0-8eae-239d3cfda211"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-13T00:09:46+09:00"
---

# CI の外部 action は actions.lock で固定し、uses には semver タグを書く

- ステータス: stable
- 日付: 2026-11-13
- 起点: [CI の nix インストーラは cryoflow 踏襲](ci-nix-installer-cryoflow.md) の「外部 action はコミット SHA で固定する」を書き直した

## 背景

CI が参照する外部 action は、版タグのままだと同じ名前で中身を差し替えられる。これまでは
`uses:` にコミット SHA を直書きし、版名をコメントで併記して固定していた。ただし、SHA とコメントの
対応を検証する仕組みがなかった。実際に `dorny/paths-filter` は `# v4.0.1` の注記に反して
v3.0.3（node20）の SHA を指しており、Node.js 20 の非推奨警告が出ていた。`cachix/cachix-action`
も、`# v17` の注記より古いコミットを指していた。composite action の中から呼ぶ action の、さらに
その先の推移的な依存は固定できていなかった。

## 決定

外部 action の固定は `gh actions-lock` が生成する `.github/workflows/actions.lock` に任せる。
lockfile は推移的な依存まで含めてコミット SHA を記録し、runner はジョブの開始時にこれを検証する。

- `uses:` には完全な semver タグ（`actions/checkout@v7.0.1` 等）を書く。版タグしか持たない
  action（`cachix/cachix-action@v17`）はその版タグを書く
- 同じリポジトリの composite action は `uses: $/.github/actions/<name>` で参照する。lockfile が
  composite の中の remote action まで追跡できるようにするため
- `uses:` を変えたら `gh actions-lock` を実行し、lockfile を同じ変更でコミットする
- `verify-lockfile.yml` が pull_request と main への push で `gh actions-lock --verify` を実行し、
  lockfile とワークフローの食い違いや、上流で検証できなくなった固定を検出する

## 根拠

固定の実体を lockfile へ移すと、`uses:` に書いた版と実際に固定したコミットの対応を
`gh actions-lock` が機械的に保証する。SHA と注記のずれは構造的に起きなくなる。推移的な依存も
固定の対象に入り、更新はこれまでどおり lockfile の差分として PR に現れ、マージ前の検証を通る。

## 棄却した代替案

### `uses:` の SHA 直書きと版名コメントを残し、lockfile を併用する

固定の情報が `uses:` と lockfile の 2 か所に分かれ、二重管理になる。注記と SHA のずれを
検出する仕組みも依然として持たない。

### SHA 直書きのまま lockfile を導入しない

推移的な依存が固定の外に残り、注記と SHA のずれも検出できない。
