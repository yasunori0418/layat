---
id: "ADR-13c8b82f-f1a0-475e-a478-06467f5ec838"
type: adr
name: "layat をフレームワークではなく純粋関数群として提供する"
section: concept
status: stable
origin: "ADR-0004 §d を書き直した"
justifies:
  - "REQ-2b0c2bb8-964f-4e36-a121-c6ea0d4be1c4"
  - "REQ-d85f0cef-0f1e-4897-a841-41b61a8dae51"
references:
  - "ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71"
  - "ADR-eeeb9dc8-9089-43ec-946a-c3ed76e98137"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:26:59+09:00"
---
# layat をフレームワークではなく純粋関数群として提供する

- ステータス: stable
- 日付: 2026-11-03
- 関連: [--json は outturn 規約に常時準拠する](outturn-conformance-principle.md), [layat.lib.mkEnv で buildEnv 合成環境を提供する](mkenv-buildenv.md), `docs/concept.md`
- 起点: ADR-0004 §d を書き直した

## 背景

layat の north-star は、配置のプリミティブからミニマルな distro の基盤を組むことにある。同じ領域には
numtide/system-manager があり、NixOS 風のモジュールシステム（`lib.evalModules`）でパッケージ・
systemd の unit・`/etc` を宣言的に作る。ブートローダ・カーネル・init・FS には関与せず、既存の distro の
上に足す層である。

layat がこの領域にどういう形で関わるか、既存のツールとの違いをどこに置くかを決める必要があった。

## 決定

layat はフレームワークではなく、テスト可能な純粋関数群（配置のプリミティブ）として提供する。
distro は利用者が layat の関数を合成して組むものとし、振る舞いをモジュールのオプションの奥に隠す
抽象は意図して持たない。

関数群の性質は [mkManifest は配置データを生成する純粋関数である](../requirements/mkmanifest-pure-function.md) と
[lib は nixpkgs.lib のみに依存する純データ生成器である](../requirements/lib-pure-data-generator.md) に置く。

## 根拠

パッケージ・systemd・`/etc` をモジュールで宣言的に管理する部分は system-manager と完全に重なる。
これを layat がモジュールとして作り直すのは再発明になる。

一方で layat の主張は「モジュールで隠さず、テスト可能な純粋関数で利用者に握らせる」ことにある。
同じ領域でも思想の段階でアプローチが違うので、競合しない。既存のツールと比べる軸は機能の有無では
なく、モジュール抽象で隠すか（NixOS・home-manager・system-manager）、純粋関数で利用者が握るか
（layat）である。

## 棄却した代替案

### distro のフレームワークとして root = `/` の systemd や `/etc` の統合まで担う

system-manager と大きく重なり、再発明になる。
