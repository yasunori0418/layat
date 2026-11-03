---
id: "ADR-eee039fc-40af-42ac-bc65-979d6eaa0d06"
type: adr
name: "nix-darwin モジュールは postActivation に起動を追記する"
status: 採用
issues:
  - "#519"
  - "#493"
origin: "Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）。ADR-0066 §2b を書き直した"
justifies:
  - "REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9"
supersedes:
  - "ADR-0066"
references:
  - "ADR-c2ea1648-5e7a-47a1-a2c6-30169f93e22c"
  - "ADR-68f453b3-e6c7-4e0d-9978-32b5a00f627e"
  - "ADR-f58e9981-bacc-4937-98b8-bddce0a057db"
  - "ADR-a0d0cf15-8234-42be-b83e-9b8271850475"
  - "ADR-1338a293-3f69-48f5-9164-efb4ec6994d6"
  - "ADR-1743c957-a732-4a44-a89a-4c03ab277f0c"
  - "ADR-0da9d817-c870-4f48-8a6d-8fb9b55e2a70"
  - "ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75"
  - "ADR-d01223d3-47a2-4bdd-ac5d-26370d553f7c"
  - "ADR-abfe1b12-1a7d-42e1-8520-c1e9fcf5846c"
---
# ADR-eee039fc-40af-42ac-bc65-979d6eaa0d06: nix-darwin モジュールは postActivation に起動を追記する

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-c2ea1648-5e7a-47a1-a2c6-30169f93e22c, ADR-68f453b3-e6c7-4e0d-9978-32b5a00f627e, ADR-f58e9981-bacc-4937-98b8-bddce0a057db, ADR-a0d0cf15-8234-42be-b83e-9b8271850475, ADR-1338a293-3f69-48f5-9164-efb4ec6994d6, ADR-1743c957-a732-4a44-a89a-4c03ab277f0c, ADR-0da9d817-c870-4f48-8a6d-8fb9b55e2a70, ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75, ADR-d01223d3-47a2-4bdd-ac5d-26370d553f7c, ADR-abfe1b12-1a7d-42e1-8520-c1e9fcf5846c, GitHub Issue #519, #493
- 起点: Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）。ADR-0066 §2b を書き直した

## 背景

nix-darwin モジュールも host の activation から engine を起動する。nix-darwin の
`system.activationScripts` は NixOS と同じ名前のオプションだが、activation で実行されるのは
nix-darwin が決めた固定名のスクリプトだけである。独自の名前で定義したスクリプトは、定義できても
実行されない。

## 決定

nix-darwin モジュールは `system.activationScripts.postActivation.text` へ engine の起動を追記する。

フックの規範は REQ-c1b3ca5f-d2f7-443c-bc4b-b18413ca97b9 に置く。

## 根拠

`postActivation` は nix-darwin が activation の最後に実行する固定名のスクリプトで、利用者の追記を
受けるための場所である。`text` は `types.lines` なので、利用者や他のモジュールの追記と共存できる。

## 棄却した代替案

### NixOS と同じく `system.activationScripts.layat` を定義する

定義はできるが実行されないので、engine が起動しない。
