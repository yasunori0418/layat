---
id: "DSG-901351ea-f01b-4bb0-a470-b890d92c980b"
type: design
name: "NixOS モジュールの実 activate は runNixOSTest の checks.nixos-vm で検証し、E2E ハーネスとは別系統に置く"
status: stable
satisfies:
  - "TP-b7f1dc79-0222-4b6e-9e91-0545046e34f2"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T13:53:52+09:00"
---
# DSG-901351ea-f01b-4bb0-a470-b890d92c980b: NixOS モジュールの実 activate は runNixOSTest の checks.nixos-vm で検証し、E2E ハーネスとは別系統に置く

## 設計

NixOS モジュールを使う場面の実 activate は、**`runNixOSTest`（nixpkgs の VM テストフレームワーク）の
flake check `checks.nixos-vm` で検証する**。`tests/e2e/` の bash ハーネス
（[E2E は tests/e2e の bash ハーネスに置き、シナリオの詳細は同ディレクトリの README へ委譲する](e2e-harness-shape.md)）のスコープ外に置く。

**E2E ハーネスへ入れずに別系統とする**理由は、要求する実行環境が違うため。
[NixOS / nix-darwin モジュールで使うときの実 activate は E2E ハーネスの対象外とする](../test-plan/nixos-vm-test-future.md) が「NixOS / nix-darwin モジュールで使うときの実 activate は E2E ハーネスの
対象外」と定めているのは、E2E が非 NixOS（ubuntu ランナー）で「nix さえあれば動く」を
検証する趣旨で組まれているのに対し、NixOS モジュールの activate には NixOS 自体が
要るからである。同じハーネスへ押し込むと、前提の異なるシナリオが 1 つの実行系に含まれる。

**`runNixOSTest` を選ぶ**のは、`system.activationScripts.layat` からの engine 起動を
検証するには実際に system activation を実行する必要があり、nixpkgs が提供する
VM テストがその唯一の現実的な手段だからである。同一構成で activate を再実行しても世代が
増えないこと（[世代スキップを project mode と system mode で行う](../adr/generation-skip-project-and-system.md)）も、同じ VM 上で確かめられる。

**検証の分担**は次の通り。

| check | 対象 | 実行する system |
|---|---|---|
| `checks.nixos-module` | NixOS モジュールの評価結果（activation スクリプトの起動文字列・manifest）| Linux（他は trivial な派生）|
| `checks.darwin-module` | nix-darwin モジュールの評価結果（`postActivation.text`）| darwin（他は trivial な派生）|
| `checks.nixos-vm` | NixOS の実 activate（配置・世代・古くなった配置の除去・再 switch での収束）| x86_64-linux（他は trivial な派生）|

非対象 system でも check 名は存在させ、`touch $out` の派生に差し替える。nix-darwin の実
activate は VM で回せないため、評価結果のアサートに留める。

## 出典

`docs/design.md`「テスト戦略」の NixOS VM テストの段落と、
[module と project では世代操作を公開しない](../adr/module-and-project-hide-generation-ops.md)、
[モジュールの検証を flake check と VM テストに置く](../adr/module-checks-separate-from-e2e.md)。
