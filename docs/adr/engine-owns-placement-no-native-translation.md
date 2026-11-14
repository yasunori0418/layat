---
id: "ADR-3d450af1-1c79-48ce-8ee9-b989dcb41ad7"
type: adr
name: "配置は全層で engine が行い、ネイティブ機構へ翻訳しない"
section: concept
status: stable
origin: "ADR-0003 §a（決定節の第 1 項）と「layat は OS のファイル管理機構とは別の機構である」節を書き直した"
justifies:
  - "REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9"
  - "DSG-17db0831-d7da-446d-ba3e-404df64c582d"
references:
  - "ADR-2119433e-de69-4447-85a5-575d58246007"
  - "ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c"
  - "ADR-59f63d1f-25df-4290-913f-02804c5eeade"
  - "ADR-90964d0f-3c2e-4201-a04b-5821f2f4c799"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# 配置は全層で engine が行い、ネイティブ機構へ翻訳しない

- ステータス: stable
- 日付: 2026-11-03
- 関連: [engine を単一の固定 Go バイナリにする](engine-fixed-go-binary.md), [古くなった配置の除去は記録どおりの symlink だけに限る](conservative-removal-invariant.md), [root の型を mkManifest だけに持たせる](roottype-mkmanifest-only.md), [devShell の shellHook で配線する](devshell-shellhook-wiring.md), `docs/design.md`, `docs/spec.md`
- 起点: ADR-0003 §a（決定節の第 1 項）と「layat は OS のファイル管理機構とは別の機構である」節を書き直した

## 背景

初期の設計では、統合層がそれぞれのプラットフォームのネイティブ機構へ配置を翻訳する想定だった。
home-manager なら `home.file` と `mkOutOfStoreSymlink`、NixOS なら `systemd.tmpfiles` の `L` 型、
nix-darwin なら `system.activationScripts` と `ln` である。

この方式では、配置の振る舞いが層ごとに別々に実装され、各プラットフォームの機構の挙動に左右される。
layat は「テスト可能な純粋関数群で、利用者が配置を明示的に管理する」ためのもので、振る舞いの源は
1 つのコアであるべきだという方針と合わない。

## 決定

配置は、CLI 単体・devShell・各モジュールのどの層でも layat の engine
（[engine を単一の固定 Go バイナリにする](engine-fixed-go-binary.md)）が manifest に従って root の下へ行う。`home.file`・
`systemd.tmpfiles`・`environment.etc` などのネイティブ機構には翻訳も委譲もしない。layat は OS が
自身の宣言的なファイル管理に使う機構とは別の機構として動く。

各層が engine を起動する配線の細目は [全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない](../requirements/modules-are-engine-wiring.md) に、層の積み方は
[層を CLI / engine / lib / common.nix / 統合層の 5 段に積み、依存を呼ぶ側から呼ばれる側への一方向に限る](../design/layer-stack-dependency.md) に置く。

## 根拠

配置の振る舞いが 1 つのコアに集まるので、テストしやすくなり、プラットフォームをまたいで同じ
スキーマと同じ挙動を保証できる。NixOS のモジュール生態系から独立した配置機構という north-star
とも合う。

layat は NixOS の宣言性を担おうとしていないので、ネイティブ機構を使わないことで宣言性を失うという
懸念はそもそも当たらない。stow や git を NixOS で動かすのに tmpfiles が関係ないのと同じで、NixOS で
layat を使うことは layat を実行することにすぎない。

`systemd.tmpfiles` は作成しか宣言的に扱わず、規則が消えたときに古くなったリンクを消さない。削除はどのみち
layat が担う必要があり（[古くなった配置の除去は記録どおりの symlink だけに限る](conservative-removal-invariant.md)）、tmpfiles を併用しても機構が
二重になるだけである。

ネイティブ機構の恩恵（プラットフォーム標準の追跡と GC、`home.file` の宣言性）は全層で手放す。
古くなった配置の追跡は、世代に残す manifest で home-manager と同等の正確さを保つ。

## 棄却した代替案

### CLI 単体では engine を使い、モジュールはネイティブ機構へ翻訳する

ネイティブ統合の恩恵は得られるが、振る舞いが層ごとに二重になる。1 つのコアで利用者が管理する
という方針に逆行する。

### NixOS の system パスの symlink だけ tmpfiles を使う

tmpfiles は作成しかしないので、古くなった配置の除去に layat の manifest が別途要り、機構が二重になる。
