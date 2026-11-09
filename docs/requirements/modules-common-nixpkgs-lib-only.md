---
id: "REQ-2bd0d35f-dc29-4771-9cfe-6998247afa0f"
type: requirement
name: "modules/common.nix は nixpkgs.lib のみに依存する"
status: stable
derives_from:
  - "UC-d39c1994-f9a5-4860-80ba-f6e584adaf14"
specification: |
  `modules/common.nix` SHALL depend only on nixpkgs.lib. It SHALL NOT introduce a
  dependency on the module system of home-manager, NixOS or nix-darwin, so that the common
  option set it holds can be imported from every integration layer without one layer's
  module system leaking into another. Which options it defines is stated by [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](module-common-options.md),
  and the same constraint on `lib` by [lib は nixpkgs.lib のみに依存する純データ生成器である](lib-pure-data-generator.md); neither is restated here. What the
  per-layer files (`modules/home-manager.nix` and its counterparts) may depend on is
  outside the scope of this item and is not stated here; that they are wiring which kicks
  the engine is stated by [全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない](modules-are-engine-wiring.md).
specification_ja: |
  `modules/common.nix` は nixpkgs.lib のみに依存しなければならない。home-manager /
  NixOS / nix-darwin の module system への依存を持ち込んではならない。同ファイルが持つ
  共通オプション集合を、ある統合層の module system を別の統合層へ漏らすことなく全統合層
  から import できるようにするためである。同ファイルがどのオプションを定義するかは
  [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](module-common-options.md)、`lib` に対する同種の制約は [lib は nixpkgs.lib のみに依存する純データ生成器である](lib-pure-data-generator.md) の担当で、いずれも本 item では
  規定しない。統合層ごとのファイル（`modules/home-manager.nix` 等）が何に依存してよいかは
  本 item の射程外であり、ここでは規定しない。それらが engine を起動する配線であることは
  [全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない](modules-are-engine-wiring.md) の担当である。
risk_exempt: "`modules/common.nix` の依存の制約であり、評価テストが nixpkgs だけで通ることが傍証になる。"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# modules/common.nix は nixpkgs.lib のみに依存する

## 仕様

| コンポーネント | 依存 |
|---|---|
| `modules/common.nix` | nixpkgs.lib のみ |

> **上は原文の写しで、規範は frontmatter が正**。同ファイルが定義する共通オプション集合
> （`enable` / `configs`（`entries` を含む）/ `backup.*` と `layat.entries` の糖衣）は
> [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](module-common-options.md)、entry submodule の型定義を `lib/types.nix` と共有することは [mkManifest 自身が evalModules で入力を確認する単一の箇所になる](mkmanifest-input-validation.md)、
> `lib` が nixpkgs.lib のみに依存する純データ生成器であることは [lib は nixpkgs.lib のみに依存する純データ生成器である](lib-pure-data-generator.md)、engine が
> stdlib-only であることは [engine は第三者依存ゼロの stdlib-only とし内部層に閉じる](engine-stdlib-only.md) の担当。統合層ごとのファイルが engine を起動する
> 配線に徹することは [全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない](modules-are-engine-wiring.md) が定めるが、**それらが何に依存してよいかの規範は現状
> どの item も持たない**（原文の `modules/home-manager.nix` 行以下も `modules/common.nix`
> 行と同じく縮退で落ちたまま未回収）。層の積み方と依存の向きを一方向に限る設計は
> [層を CLI / engine / lib / common.nix / 統合層の 5 段に積み、依存を呼ぶ側から呼ばれる側への一方向に限る](../design/layer-stack-dependency.md) が持つ。
>
> **新規 item にした理由（[lib は nixpkgs.lib のみに依存する純データ生成器である](lib-pure-data-generator.md) との粒度合わせ）**: 依存制約は層ごとに 1 item
> 立てるのが既存の粒度で、`lib` は [lib は nixpkgs.lib のみに依存する純データ生成器である](lib-pure-data-generator.md)、engine は [engine は第三者依存ゼロの stdlib-only とし内部層に閉じる](engine-stdlib-only.md) が各層 1 item
> として持つ。`modules/common.nix` は [層を CLI / engine / lib / common.nix / 統合層の 5 段に積み、依存を呼ぶ側から呼ばれる側への一方向に限る](../design/layer-stack-dependency.md) の 5 段のうち `lib` とも統合層とも
> 別の 1 段であり、同じ粒度に揃えるなら独立した item になる。既存 item への編入も
> 検討したが、[全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](module-common-options.md) は「共通オプションとして何を公開するか」の item で軸が別
> （公開する集合 vs 依存の範囲）であり、[lib は nixpkgs.lib のみに依存する純データ生成器である](lib-pure-data-generator.md) へ相乗りさせると `lib` と
> `common.nix` という別の層に対する 2 つの依存制約が 1 item に同居して 1 item 1 主張が
> 崩れるため、いずれも採らなかった。[層を CLI / engine / lib / common.nix / 統合層の 5 段に積み、依存を呼ぶ側から呼ばれる側への一方向に限る](../design/layer-stack-dependency.md) の層構成図は `common.nix` を「options 型定義
> （nixpkgs.lib のみ依存）」と記載するが、これは設計の図であって規範文（specification）
> ではないため、本 item が規範を持つ。

## 出典

`docs/spec.md`「依存関係」節の表の `modules/common.nix` 行「nixpkgs.lib のみ」。#209 の
分割では同表の `internal/` 行が [engine は第三者依存ゼロの stdlib-only とし内部層に閉じる](engine-stdlib-only.md) へ入った一方、`modules/common.nix` 行は
どの item の規範にも入らないまま縮退で落ちていた（epic #203 / issue #228 で回収）。
`lib` に対する同種の依存制約は、同表の `lib/` 行ではなく別出典（「lib API」節）の
[lib は nixpkgs.lib のみに依存する純データ生成器である](lib-pure-data-generator.md) が持つ。

同表の他行と異なり原文の当該行は ADR 参照を伴わず、`modules/common.nix` の依存制約
そのものを決定した ADR も無いため、`justifies` は張らない（側面としては、lib を
データ生成に徹させる [lib は純データ生成に徹する](../adr/lib-produces-data-only.md)、entry submodule の型定義を `lib/types.nix` と
`modules/common.nix` で共有すると定めた [モジュールの entries に lib と同じ型を使う](../adr/module-entries-share-lib-type.md)、モジュールを配線に徹させる [モジュールは root と起動の時機だけを供給して engine を起動する](../adr/modules-wire-engine-kick-only.md) が
この制約と整合する）。
