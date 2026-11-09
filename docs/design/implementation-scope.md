---
id: "DSG-4a84f282-76ea-47c9-aede-deac12ff5257"
type: design
name: "実装スコープを CLI 単体 + project mode + home mode + system mode と、home-manager / NixOS / nix-darwin モジュールとする"
status: stable
satisfies:
  - "REQ-14f0aec9-abae-4621-82f3-40536a1ad904"
  - "TP-b7f1dc79-0222-4b6e-9e91-0545046e34f2"
  - "REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:22:37+09:00"
---
# DSG-4a84f282-76ea-47c9-aede-deac12ff5257: 実装スコープを CLI 単体 + project mode + home mode + system mode と、home-manager / NixOS / nix-darwin モジュールとする

## 設計

| 対象 | スコープ |
|---|---|
| CLI 単体（起動形態。配置モードは root マーカーが決める）| 実装する |
| project mode（devShell / `projectRoot`）| 実装する（コア）|
| home-manager モジュール | 実装する |
| NixOS / nix-darwin モジュール | 実装する（system mode の配線・→ [NixOS / nix-darwin モジュールは modules/kick.nix を介した system mode の配線とする](module-stubs.md)）|
| system mode（`systemRoot`）| 実装する（変更系操作は root 権限が要る）|

[layat CLI は PATH 常駐の一次 UX で、project mode は devShell 同梱を canonical とする](../requirements/cli-primary-ux-installation.md) という要求がある以上、
**最初に完成させるべきは CLI と project mode**になる。home mode（`homeRoot`）を
同時に含めるのは、そこが rollback・世代管理という layat 固有機構を通る唯一の使い方であり、
ここを外すと世代機構が実装されないまま残るため。

system mode（`systemRoot`）は、root マーカーの union に最初から含めてあり（root マーカーの型は
`requirement` 側の担当）、engine の root 解決・profile 基底・記録基底に分岐を足して実装している。
変更系操作（`apply` / `reset` / `rollback`）は engine が root 権限を確認する。

NixOS / nix-darwin モジュールは、[全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない](../requirements/modules-are-engine-wiring.md) の通り配線に過ぎない。home-manager モジュールで
確立した engine kick の形（[engine の起動を entrypoint 駆動とビルド済み manifest の 2 クラスに束ね、統合層の差を link-farm の取得方法だけに閉じる](integration-layer-kick-classes.md) のビルド済み manifest クラス）をそのまま使い、
root を `systemRoot` に pin する（[NixOS / nix-darwin モジュールは modules/kick.nix を介した system mode の配線とする](module-stubs.md)）。engine 側で足したのは、boot のたびに実行される
再 activation で世代が増えないよう世代スキップを system mode へ広げたことだけである（[世代スキップを project mode と system mode で行う](../adr/generation-skip-project-and-system.md)）。
実 activate の扱いは [NixOS / nix-darwin モジュールで使うときの実 activate は E2E ハーネスの対象外とする](../test-plan/nixos-vm-test-future.md) の通り。

**モジュール対応の位置づけ**も同じスコープの決め方から来る。モジュール層は他のモジュールシステムの
switch と一括で動いてほしいユースケースを拾うために存在するのであって、各モジュールの
内部事情を layat の設計へ持ち込まない。持ち込むと [全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない](../requirements/modules-are-engine-wiring.md) の「ネイティブ機構へ
翻訳しない」が層ごとに崩れる。

## 出典

分割時点の原文（縮退前）の `docs/design.md`「プロジェクト構成」末尾の実装スコープ段落と、
「モジュール統合設計」→「各統合層の動作」末尾のモジュール対応の位置づけ。

なお同原文の「使用パターン」→「パターン 2」は CLI 単体を
「`standalone CLI`（home mode）」と呼んでいたが、本 item はこれを採らない。用語 item
[standalone](../ubiquitous-language/standalone.md) が、CLI 単体は配置モードではなく
起動形態であり、「`standalone = home mode`」と短絡しない（配置モードは `root` マーカーが決める）と
定めているため。
