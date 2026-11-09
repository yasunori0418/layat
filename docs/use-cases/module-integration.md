---
id: "UC-d39c1994-f9a5-4860-80ba-f6e584adaf14"
type: use_case
name: "既に home-manager / NixOS / nix-darwin を使っている環境へ layat をモジュールとして組み込む"
status: stable
refines:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T22:44:41+09:00"
---
# 既に home-manager / NixOS / nix-darwin を使っている環境へ layat をモジュールとして組み込む

## 使われ方

既に home-manager / NixOS / nix-darwin で環境を管理しているユーザーが、layat を CLI 単体
としてではなくモジュールとして組み込み、host の activation（`home-manager switch` /
`nixos-rebuild switch` 等）に layat の配置を載せる。flake-parts を使っているプロジェクト向け
には `perSystem.layat` を `flake.layat.<system>` へ transpose する flakeModule も提供する
（→ [layat output の flake-parts module を flakeModules.default で公開する](../adr/flake-parts-module-published.md)）。

統合はあくまで**オプション**である。コアロジックはモジュールシステムに依存しない純粋関数
として実装され、統合層はコアの薄いラッパーに過ぎない。配置の実体は全層で layat 自身の固定
Go エンジンが実行し、`home.file` / `systemd.tmpfiles` などプラットフォームのネイティブ機構
へは委譲しない（→ [配置は全層で engine が行い、ネイティブ機構へ翻訳しない](../adr/engine-owns-placement-no-native-translation.md) / [engine を単一の固定 Go バイナリにする](../adr/engine-fixed-go-binary.md)）。

```
layat コア（配置エンジン）= 振る舞いの単一の源
        ↑ 起動するだけ
モジュール層 = root と activation hook を供給する薄い配線
```

モジュール経由でも役割ごとに名前つき config を分けて独立した profile を取れる。ただし
profile は**内部機構**（古くなった配置の追跡）に留め、ユーザー向け rollback は host に一本化する
（host rollback は旧 config の再 activate で layat が自動追従する）。

## この使われ方が要求すること

- モジュールが root と activation タイミングを供給して engine をキックするだけの配線であり、
  ネイティブ機構へ翻訳しないこと
- 全モジュールが共通オプションの同一集合を公開すること
- モジュールで使うときは entrypoint 発見・eval を経ず、ビルド済み link-farm を engine へ直接
  適用すること
- 配置に失敗したとき host の switch が止まること
- モジュール経由でも名前つき config ごとに独立した profile を取れること
- ユーザー向け rollback を host へ一本化し、layat profile は前進のみで追従すること
- flake-parts を使う場面でも直書きと同一の derivation が生まれ、CLI のアドレッシングが変わらないこと

## 出典

`docs/concept.md`「設計の哲学」の「配置ロジックはコアが所有し、モジュールは配線に徹する」
「統合は『オプション』」、および「世代管理（standalone）」のモジュール時の profile の扱いに
関する記述。

同じ「設計の哲学」の「home-manager に依存しない」節は本 item の出典ではない。同節が述べるのは
HM 不在の環境でも同じ設定定義で動くことであり、それを体現するのは統合ではなく CLI 単体の
使われ方（[home mode で外部リポジトリの中身をバージョン固定して $HOME 配下の任意パスへ配置する](home-mode-pinned-repo-placement.md) / [プロジェクト repo 内へ nix store の物を devShell 入室のたびに配置してチームで共有する](project-mode-in-repo-placement.md)）。本 item が引き受けるのは、そこから派生する
「コアロジックはモジュールシステムに依存しない純粋関数で、統合層は薄いラッパーに過ぎない」
という統合層側の含意に留まる。

Issue #211 の素材リストにはこの項目が無いが、モジュール統合は concept.md が繰り返し述べる
使われ方であり、統合層の `requirement` 群がこの use_case を持たないと接続漏れになるため
item 化した（2026-08-02 確定）。
