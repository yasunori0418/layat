---
id: "ADR-d01223d3-47a2-4bdd-ac5d-26370d553f7c"
type: adr
name: "モジュールの検証を flake check と VM テストに置く"
status: 採用
issues:
  - "#519"
  - "#493"
origin: "Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）。ADR-0066 §5b を書き直した"
justifies:
  - "DSG-901351ea-f01b-4bb0-a470-b890d92c980b"
  - "TP-b7f1dc79-0222-4b6e-9e91-0545046e34f2"
references:
  - "ADR-c2ea1648-5e7a-47a1-a2c6-30169f93e22c"
  - "ADR-68f453b3-e6c7-4e0d-9978-32b5a00f627e"
  - "ADR-f58e9981-bacc-4937-98b8-bddce0a057db"
  - "ADR-a0d0cf15-8234-42be-b83e-9b8271850475"
  - "ADR-1338a293-3f69-48f5-9164-efb4ec6994d6"
  - "ADR-1743c957-a732-4a44-a89a-4c03ab277f0c"
  - "ADR-0da9d817-c870-4f48-8a6d-8fb9b55e2a70"
  - "ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75"
  - "ADR-abfe1b12-1a7d-42e1-8520-c1e9fcf5846c"
---
# ADR-d01223d3-47a2-4bdd-ac5d-26370d553f7c: モジュールの検証を flake check と VM テストに置く

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-c2ea1648-5e7a-47a1-a2c6-30169f93e22c, ADR-68f453b3-e6c7-4e0d-9978-32b5a00f627e, ADR-f58e9981-bacc-4937-98b8-bddce0a057db, ADR-a0d0cf15-8234-42be-b83e-9b8271850475, ADR-1338a293-3f69-48f5-9164-efb4ec6994d6, ADR-1743c957-a732-4a44-a89a-4c03ab277f0c, ADR-0da9d817-c870-4f48-8a6d-8fb9b55e2a70, ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75, ADR-abfe1b12-1a7d-42e1-8520-c1e9fcf5846c, GitHub Issue #519, #493
- 起点: Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）。ADR-0066 §5b を書き直した

## 背景

NixOS / nix-darwin モジュールの振る舞いには、モジュールを評価した結果（どのフックにどの起動文字列が入るか）と、
実際に activate したときの配置の 2 つがある。後者を確かめるには NixOS の VM が要る。`tests/e2e/` の
ハーネスは NixOS ではない環境の単一ジョブで動き、VM を前提にしていない。

## 決定

NixOS / nix-darwin モジュールの評価結果は、eval でアサートする flake check（`checks.nixos-module` /
`checks.darwin-module`）で検証する。NixOS の実際の activate は、`runNixOSTest` の VM テスト
（`checks.nixos-vm`。x86_64-linux のみ）で検証する。どちらも `tests/e2e/` のハーネスとは別の系統に置く。
nix-darwin の flake input は check 専用とし、`lib/` とモジュール本体はそれに依存しない。

検証の配置と範囲は DSG-901351ea-f01b-4bb0-a470-b890d92c980b と TP-b7f1dc79-0222-4b6e-9e91-0545046e34f2 に置く。

## 根拠

VM を要するテストを e2e ハーネスに入れると、ハーネス全体が VM を前提にすることになり、NixOS ではない環境で
回らなくなる。別の flake check にすれば、ハーネスはこれまでどおり単一ジョブで動き、VM テストは動かせる
環境でだけ動く。

評価結果のアサートは VM を立ち上げずに済むので軽く、nix-darwin のように VM で activate できないモジュールも
同じ方法で検証できる。

nix-darwin の input を check 専用にするのは、`lib/` が nixpkgs のみに依存するという制約を保つためである。

## 棄却した代替案

### e2e ハーネスに VM のシナリオを足す

ハーネスが VM と NixOS を前提にすることになり、NixOS ではない環境の単一ジョブで回せなくなる。

### 評価結果のアサートだけで済ませる

フックに起動文字列が入ることは確かめられるが、実際に activate しないと分からない失敗を見逃す。
