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
references:
  - "ADR-0066"
---
# ADR-eee039fc-40af-42ac-bc65-979d6eaa0d06: nix-darwin モジュールは postActivation に起動を追記する

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-0066, GitHub Issue #519, #493
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
