---
id: "REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9"
type: requirement
name: "全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない"
derives_from:
  - "UC-d39c1994-f9a5-4860-80ba-f6e584adaf14"
specification: |
  Basic use SHALL be conceived around project mode and the standalone CLI, support for
  modules existing only to pick up the use case of wanting placement to run together with
  the switch of another module system. Every module (home-manager / NixOS / nix-darwin)
  and the devShell SHALL alike be wiring that does no more than kick the layat engine, each
  layer supplying only the root and the timing of activation. They SHALL NOT translate
  entries into a native mechanism of the platform, such as `systemd.tmpfiles` or
  `home.file`. Placement and stale removal SHALL be performed, in every layer, by the same
  engine and the same store manifest, so that the behaviour is not duplicated per layer.
  The mechanisms of the operating system SHALL remain outside the concern of layat. The
  NixOS module SHALL kick the engine from `system.activationScripts.layat`, and the
  nix-darwin module SHALL kick it from what it appends to
  `system.activationScripts.postActivation`; the hook the home-manager module uses is
  stated by REQ-8085f194-c903-4ecb-abd8-c719fe7b3292.
specification_ja: |
  基本的な利用は project mode（プロジェクト内配置）と standalone CLI を中心に考えなければ
  ならず、モジュール対応は、他のモジュールシステムの switch と一括で動いてほしいユースケースを
  拾うためだけに存在する。全モジュール（home-manager / NixOS / nix-darwin）と devShell は
  一律、layat エンジンをキックするだけの配線でなければならず、各層は root と activation
  タイミングだけを供給する。`systemd.tmpfiles` / `home.file` などプラットフォームの
  ネイティブ機構へ翻訳してはならない。配置・stale 除去は全層で同一の engine + store
  マニフェストが行わなければならず、振る舞いが層ごとに二重化しないようにする。OS の機構は
  layat の関心外でなければならない。NixOS モジュールは `system.activationScripts.layat` から、
  nix-darwin モジュールは `system.activationScripts.postActivation` への追記から engine を
  起動しなければならない（home-manager モジュールが使うフックは
  REQ-8085f194-c903-4ecb-abd8-c719fe7b3292 の担当）。
---
# REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9: 全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない

## 仕様

基本的な利用は **project mode（プロジェクト内配置）と standalone CLI** を中心に考える。
モジュール対応は、他のモジュールシステムの switch と**一括で動いてほしいユースケース**を拾う
ためだけに存在する。全モジュール（HM / NixOS / nix-darwin）と devShell は
**一律「layat エンジンをキックするだけ」の配線**であり、各層は root と activation タイミングを
供給するだけ。`systemd.tmpfiles` / `home.file` へは翻訳しない。配置・stale 除去は全層で同一の
engine + store マニフェスト。

- NixOS モジュールは `system.activationScripts.layat` から engine を起動する
  **配線**に徹する。`systemd.tmpfiles` へは翻訳しない。OS の機構（tmpfiles 等）は layat の
  関心外
- nix-darwin モジュールは `system.activationScripts.postActivation` への追記から engine を起動する

> **上は原文の写しを ADR-0066 に合わせて改めたもので、規範は frontmatter が正**。各層が供給する
> root の具体は REQ-8d965ca2-f8fd-44a4-87f3-94e850e9f85b（HM の `$HOME`）と REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10（NixOS / nix-darwin の `systemRoot`）、HM の
> activation フックと起動契約は REQ-8085f194-c903-4ecb-abd8-c719fe7b3292、NixOS / nix-darwin の起動契約は REQ-a220bb28-89de-42b0-a854-e7468755ecb1、devShell の起動タイミングは REQ-a0bdf6db-6c0c-476c-916a-61ee4e4510d9、
> module 経路の rollback を host 世代へ一本化することは REQ-844ee375-919f-4341-81e1-a5f89fd32840、engine 自身が
> ネイティブ FS 操作で配置することは REQ-6c4e174a-4d16-477a-96ff-17cb4eb5b564 の担当。
>
> CLI が PATH 常駐の一次 UX であることと、その導入形（standalone のグローバル導入・
> project mode の devShell 同梱）は REQ-14f0aec9-abae-4621-82f3-40536a1ad904 の担当。本 item はモジュール対応の側から見た
> 位置づけ（一括 switch 用途を拾うためだけに存在する配線であること）を規定する。

## 出典

`docs/spec.md`「モジュール別動作仕様」の前文と、「NixOS モジュール（将来拡張）」
「nix-darwin モジュール（将来拡張）」節。将来拡張 2 節のうち root の供給元は REQ-8d965ca2-f8fd-44a4-87f3-94e850e9f85b、
rollback の一本化は REQ-844ee375-919f-4341-81e1-a5f89fd32840 が担当し、残る「activation フックから engine を起動する
配線に徹する」を本 item に畳んだ。

決定の実体は ADR-0003「配置ロジックは全層 layat エンジンが所有し、モジュールは配線に徹する」で、
モジュール対応を一括 switch 用途に限る positioning は ADR-0007「汎用 layat CLI を一次 UX に
昇格し、entrypoint 発見＋root 明示モデルへ移行する」が定めている。NixOS / nix-darwin の hook を
具体化したのは ADR-0066「NixOS / nix-darwin モジュールを system mode の配線として実装する」。
