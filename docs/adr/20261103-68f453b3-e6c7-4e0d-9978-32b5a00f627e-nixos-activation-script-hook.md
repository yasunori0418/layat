---
id: "ADR-68f453b3-e6c7-4e0d-9978-32b5a00f627e"
type: adr
name: "NixOS モジュールは system.activationScripts.layat から起動する"
status: stable
issues:
  - "#519"
  - "#493"
origin: "Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）。ADR-0066 §2a を書き直した"
justifies:
  - "REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9"
references:
  - "ADR-f58e9981-bacc-4937-98b8-bddce0a057db"
  - "ADR-1338a293-3f69-48f5-9164-efb4ec6994d6"
  - "ADR-1743c957-a732-4a44-a89a-4c03ab277f0c"
  - "ADR-0da9d817-c870-4f48-8a6d-8fb9b55e2a70"
  - "ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75"
  - "ADR-abfe1b12-1a7d-42e1-8520-c1e9fcf5846c"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-68f453b3-e6c7-4e0d-9978-32b5a00f627e: NixOS モジュールは system.activationScripts.layat から起動する

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-f58e9981-bacc-4937-98b8-bddce0a057db, ADR-1338a293-3f69-48f5-9164-efb4ec6994d6, ADR-1743c957-a732-4a44-a89a-4c03ab277f0c, ADR-0da9d817-c870-4f48-8a6d-8fb9b55e2a70, ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75, ADR-abfe1b12-1a7d-42e1-8520-c1e9fcf5846c, GitHub Issue #519, #493
- 起点: Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）。ADR-0066 §2a を書き直した

## 背景

NixOS モジュールは host の activation から engine を起動する。NixOS は activation の処理を名前付きの
スクリプトとして並べ、`deps` で順序を付ける。engine は system mode で配置するので、`/etc` やユーザーが
用意されてから動く必要がある。`nixos-rebuild dry-activate` でも、変更の予定を示せることが望ましい。

## 決定

NixOS モジュールは `system.activationScripts.layat` を定義し、そこから engine を起動する。
`deps` には `"etc"` と `"users"` を入れ、`supportsDryActivation = true` とする。

フック名と起動の規範は REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9 に置く。

## 根拠

`deps` に `etc` と `users` を入れるので、engine は `/etc` とユーザーが用意されたあとに動く。

`supportsDryActivation = true` が無いと、スクリプトは dry-activate 用の activate に含まれず、
dry-activate のときに何も示せない。

## 棄却した代替案

旧 ADR-0066 §2a は、この決定について代替案を記録していない。
