---
id: "TP-b7f1dc79-0222-4b6e-9e91-0545046e34f2"
type: test_plan
name: "NixOS / nix-darwin モジュールで使うときの実 activate は E2E ハーネスの対象外とする"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
specification: |
  Actually activating the NixOS and nix-darwin module paths SHALL be outside the scope of
  the E2E harness, since it requires a VM or a sandbox whereas the harness is a single
  non-NixOS job, and a VM test SHALL NOT be a prerequisite for that harness. Verification
  of those paths SHALL be carried by a separate series of flake checks: evaluation
  assertions on both modules, and a VM test of the NixOS module.
specification_ja: |
  NixOS / nix-darwin モジュールで使うときの実 activate は E2E ハーネスのスコープ外としなければ
  ならない（VM / sandbox を要するのに対しハーネスは非 NixOS の単一ジョブであるため）。
  VM テストを同ハーネスの前提条件にしてはならない。これらの使い方の検証は、両モジュールの
  評価アサートと NixOS モジュールの VM テストという別系統の flake check が担わなければ
  ならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T22:44:41+09:00"
---
# TP-b7f1dc79-0222-4b6e-9e91-0545046e34f2: NixOS / nix-darwin モジュールで使うときの実 activate は E2E ハーネスの対象外とする

## 仕様

NixOS / nix-darwin モジュールで使うときの実 activate は VM / sandbox を要し、本ハーネス（非 NixOS の
単一ジョブ）のスコープ外。検証は flake check の別系統（評価アサートと NixOS VM テスト
`runNixOSTest`）が担う。

> **上は原文の写しを [モジュールの検証を flake check と VM テストに置く](../adr/module-checks-separate-from-e2e.md) に合わせて改めたもので、規範は frontmatter が正**。E2E ハーネス本体の検証範囲は
> [非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する](e2e-harness-scope.md)、NixOS / nix-darwin モジュールが engine キックの配線であることは
> [全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない](../requirements/modules-are-engine-wiring.md)、別系統の check の分担は [NixOS モジュールの実 activate は runNixOSTest の checks.nixos-vm で検証し、E2E ハーネスとは別系統に置く](../design/nixos-vm-test-position.md) の担当。`runNixOSTest` という手段は
> 規範に採らない。

## 出典

`docs/spec.md`「E2E 検証範囲（非 NixOS）」節の末尾の注記。

決定の実体は [テストを sandbox 可否で flake check と E2E に分ける](../adr/test-strategy-layers.md)「テストを sandbox 可否で flake check と E2E に分ける」と
[モジュールの検証を flake check と VM テストに置く](../adr/module-checks-separate-from-e2e.md)「モジュールの検証を flake check と VM テストに置く」で、
NixOS VM テストを E2E のハーネスとは別系統の check で検証すると定めている。

> **本 item は `requirement` から `test_plan` へ移設した**（→ Issue #238。旧 ID は
> `REQ-690f2730-2628-420d-8e72-ed1ce747ac1e`）。E2E ハーネスの検証範囲からの除外を宣言するスコープ外規定であり、
> ユーザーの使われ方（use_case）から導かれるものではないため、use_case を親に持てず接続漏れに
> なっていた（当時の判断は Issue #211）。モジュールでの使い方そのものは [既に home-manager / NixOS / nix-darwin を使っている環境へ layat をモジュールとして組み込む](../use-cases/module-integration.md) が担い、
> [全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない](../requirements/modules-are-engine-wiring.md) 等がそちらへ紐づく。テスト計画の型を新設して solution 直下で受けることに
> したため、`derives_from` は [layat は nix store の物を任意パスへ置く配置プリミティブであり、設定を生成せずユーザーが配置を明示的に握る](../solution/layat-placement-primitive.md) を指す。[非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する](e2e-harness-scope.md)・[エンベロープの outturn 適合を Go テストと E2E の両方で検証する](json-conformance-verification.md) も同じ経緯で
> 移設した。[lib.__internal は private helper のテスト seam として公開する](lib-internal-test-seam.md) は #238 では「テスト計画そのものではない」と判断して見送ったが、
> Issue #239 でその判断を改めて移設した（経緯は同 item の注記）。
