---
id: "ADR-fb5e9bee-8132-431b-9948-6ad9995c9a9a"
type: adr
name: "キャッシュは 3 system の matrix ビルドで投入する"
section: ci-release-docs
status: stable
origin: "参照: https://github.com/yasunori0418/cryoflow の CI 構成（同一メンテナの実績構成）。ADR-0012 §4 を書き直した"
justifies:
  - "INF-d1230e1f-8ba8-49d8-8386-409bfbb7dd27"
  - "INF-af33c5a1-f846-4f09-8e43-c6ea042d6e06"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:00:56+09:00"
---
# キャッシュは 3 system の matrix ビルドで投入する

- ステータス: stable
- 日付: 2026-11-01
- 起点: 参照: https://github.com/yasunori0418/cryoflow の CI 構成（同一メンテナの実績構成）。ADR-0012 §4 を書き直した

## 背景

layat を使う側（activation や CI）は layat のバイナリを必要とする。キャッシュが無いと、利用者の
手元で Go のビルドを実行することになる。どの system のビルドを、どうやってキャッシュへ投入するかを決める必要がある。

## 決定

キャッシュへの投入は、x86_64-linux・aarch64-linux・aarch64-darwin の matrix で `nix build .#layat` を
実行し、cachix のキャッシュ `yasunori0418` へ送って行う。投入の契機は [キャッシュへの投入を main への push で起動する](cachix-push-on-main.md)、ジョブの組み方は [キャッシュへの投入を cachix-action の自動 push に任せる](cachix-push-via-cachix-action.md) が
決めている。

検証の matrix は [CI パイプライン（flake check の os×system マトリクスと E2E ジョブ）](../infrastructure/ci-pipeline.md)、キャッシュ投入の構成は
[バイナリキャッシュ（main push 時の cachix 投入）](../infrastructure/binary-cache.md) に置く。

## 根拠

キャッシュを置く system を、CI がテストしている system にそろえる。ネイティブのランナーで
ビルドするので、クロスビルドの設定が要らない。

## 棄却した代替案

旧 ADR-0012 §4 は、この決定について代替案を記録していない。
