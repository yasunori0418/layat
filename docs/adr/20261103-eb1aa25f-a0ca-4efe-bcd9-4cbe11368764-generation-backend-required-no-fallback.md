---
id: "ADR-eb1aa25f-a0ca-4efe-bcd9-4cbe11368764"
type: adr
name: "世代バックエンドの注入を必須にし nix-env への暗黙のフォールバックを持たない"
status: 採用
issues:
  - "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §2d を書き直した"
justifies:
  - "REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b"
---
# ADR-eb1aa25f-a0ca-4efe-bcd9-4cbe11368764: 世代バックエンドの注入を必須にし nix-env への暗黙のフォールバックを持たない

- ステータス: 採用
- 日付: 2026-11-03
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §2d を書き直した

## 背景

engine の世代操作の差し替え点は、何も注入されていない（nil の）とき nix-env を暗黙に呼んでいた。
cmd 層が注入を忘れても本番では動いてしまい、テストも nil のまま nix-env を呼ぶ。

## 決定

profileDir を解決する engine の全入口（apply・rollback・reset・prune・profile の解決）は、世代バックエンドの
注入を必須とし、注入が無ければ error にする。nix-env への暗黙のフォールバックは持たない。cmd 層は常に
nixprofile を注入する。

規範は REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b に置く。

## 根拠

注入し忘れを error にすれば、全入口に注入されていることをテストで確かめられる。フォールバックがあると、
注入し忘れが nix-env の直叩きとして本番の挙動に紛れ込む。

## 棄却した代替案

### nil のとき nix-env を呼ぶ差し替え点を残す

注入し忘れが error にならずに通り、世代操作が契約越しであることを検証できない。
