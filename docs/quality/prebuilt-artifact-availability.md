---
id: "QA-58522afb-31d5-4a1f-a7df-0858efa9e44b"
type: quality
name: "最新 main のビルド成果物を再ビルドなしに消費できる状態を保つ"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
specification: |
  The build outputs of the latest `main` SHALL be published to a public binary cache for
  every supported platform, so that a consumer that takes layat as a flake input can obtain
  them without building from source. Publication SHALL be driven by the changes that
  determine those outputs, and SHALL NOT be tied to release tags, so that the cache tracks
  `main` rather than the release cadence.
specification_ja: |
  最新 main のビルド成果物は、サポート対象の全プラットフォームについて公開バイナリ
  キャッシュへ投入されなければならない（layat を flake input として取り込む消費側が、
  ソースからビルドせずに取得できるようにするため）。投入は成果物の内容を決める変更に
  よって駆動されなければならず、リリースタグに紐づけてはならない（キャッシュがリリースの
  頻度ではなく main に追従するようにするため）。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-11T00:30:18+09:00"
---
# 最新 main のビルド成果物を再ビルドなしに消費できる状態を保つ

## 仕様

layat は flake input として消費されるツールであり、消費側の CI やローカル環境が layat を使う
たびにソースからビルドするのは実用上のコストになる。最新 main の成果物を公開キャッシュから
取得できる状態を保つ。

投入の契機を**リリースタグではなく main の変更**に置く点が要点になる。タグに紐づけると
リリース間の main 変更がキャッシュされず、「flake input として最新 main を追う」という
消費のされ方と噛み合わない。リリース（[リリースはバージョンの一次情報の変更で駆動し、手作業の工程を挟まない](release-driven-by-source-of-truth.md)）とキャッシュは互いに独立に駆動する。

> **受益者は消費側だが、型は quality に置いた。** この norm を落として困るのはビルド時間を
> 払う消費側であり、[規範の型は、それが誰を縛るかで決める](norm-type-by-who-it-binds.md) の quality の判定の問い（落ちて気づくのは
> contributor だけ）とは張力がある。それでも quality に置くのは、落としても **shipped artifact も
> 検証も変わらない**（同一 derivation をキャッシュから取得するか実ビルドするかの差でしかない）ため。
> `requirement` の判定の問い「layat の使い方が変わるか」は成立せず、対応する use_case も無い。

## 出典

[キャッシュへの投入を main への push で起動する](../adr/cachix-push-on-main.md)、[キャッシュは 3 system の matrix ビルドで投入する](../adr/ci-cache-via-matrix-build.md)、
[キャッシュへの投入を cachix-action の自動 push に任せる](../adr/cachix-push-via-cachix-action.md) が置いた方針を、
基盤（INF）から
分離して quality item として立てたもの。投入先のキャッシュ名・マトリクス構成・トリガの
`paths` は [バイナリキャッシュ（main push 時の cachix 投入）](../infrastructure/binary-cache.md) が持つ。
