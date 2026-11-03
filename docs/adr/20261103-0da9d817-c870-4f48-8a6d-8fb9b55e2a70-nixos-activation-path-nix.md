---
id: "ADR-0da9d817-c870-4f48-8a6d-8fb9b55e2a70"
type: adr
name: "NixOS の activation の PATH に nix を前置する"
status: 採用
issues:
  - "#519"
  - "#493"
origin: "Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）。ADR-0066 §4b を書き直した"
justifies:
  - "REQ-a220bb28-89de-42b0-a854-e7468755ecb1"
  - "DSG-0e186e89-c8c6-4bca-9daa-03855e8d5cda"
references:
  - "ADR-0066"
---
# ADR-0da9d817-c870-4f48-8a6d-8fb9b55e2a70: NixOS の activation の PATH に nix を前置する

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-0066, GitHub Issue #519, #493
- 起点: Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）。ADR-0066 §4b を書き直した

## 背景

engine は世代をコミットするとき、既定の世代バックエンドを通して `nix-env` を呼ぶ。NixOS の activation
スクリプトが実行されるときの `PATH` には `nix` が含まれていない。そのままでは、NixOS モジュールから起動した
engine は世代のコミットで失敗する。

## 決定

NixOS モジュールは、engine を起動する前に activation の `PATH` へ `config.nix.package` の `bin` を前置する。

起動環境の規範は REQ-a220bb28-89de-42b0-a854-e7468755ecb1 と DSG-0e186e89-c8c6-4bca-9daa-03855e8d5cda に置く。

## 根拠

`config.nix.package` は host の構成が使う nix である。これを前置すれば、世代のコミットに要る `nix-env` が
activation の中でも見つかる。

## 棄却した代替案

旧 ADR-0066 §4b は、この決定について代替案を記録していない。
