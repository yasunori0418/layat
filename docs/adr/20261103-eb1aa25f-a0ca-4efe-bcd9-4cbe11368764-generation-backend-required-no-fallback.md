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
  - "DSG-f8d9bba5-cfd5-4a19-b1fd-fd8ecc8a1814"
supersedes:
  - "ADR-0057"
references:
  - "ADR-30f105c7-85a8-4fab-9dea-3277d932b524"
  - "ADR-419914ba-9896-4dd1-853b-fb514fcb9f38"
  - "ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1"
  - "ADR-58fdbad9-5ed7-469d-b6b2-6462b2173888"
  - "ADR-5fc58c8f-dc0e-4c1d-8c1e-17d7ac7f0385"
  - "ADR-74423bb8-1453-4689-a25e-30bf46536afc"
  - "ADR-b1e1dd5e-61a4-4d04-8389-137ef2347c56"
  - "ADR-d97154cd-a2e5-4853-a682-800ff9435c89"
  - "ADR-f90509f2-d75a-4162-baa5-a5cfdb8a7c30"
---
# ADR-eb1aa25f-a0ca-4efe-bcd9-4cbe11368764: 世代バックエンドの注入を必須にし nix-env への暗黙のフォールバックを持たない

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-30f105c7-85a8-4fab-9dea-3277d932b524, ADR-419914ba-9896-4dd1-853b-fb514fcb9f38, ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1, ADR-58fdbad9-5ed7-469d-b6b2-6462b2173888, ADR-5fc58c8f-dc0e-4c1d-8c1e-17d7ac7f0385, ADR-74423bb8-1453-4689-a25e-30bf46536afc, ADR-b1e1dd5e-61a4-4d04-8389-137ef2347c56, ADR-d97154cd-a2e5-4853-a682-800ff9435c89, ADR-f90509f2-d75a-4162-baa5-a5cfdb8a7c30
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §2d を書き直した

## 背景

engine の世代操作の差し替え点は、何も注入されていない（nil の）とき nix-env を暗黙に呼んでいた。
cmd 層が注入を忘れても、本番では nix-env を直接呼ぶ挙動として動いてしまう。

## 決定

profileDir を解決する engine の処理（apply・rollback・reset・prune・profile の解決）は、どれも世代バックエンドの注入を必須とし、注入が無ければ error にする。
nix-env への暗黙のフォールバックは持たない。
cmd 層は常に nixprofile を注入する。

規範は REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b、engine と cmd 層での受け渡しは DSG-f8d9bba5-cfd5-4a19-b1fd-fd8ecc8a1814 に置く。

## 根拠

注入し忘れを error にすれば、これらの処理のすべてに注入されていることをテストで確かめられる。
フォールバックがあると、注入し忘れが nix-env の直叩きとして本番の挙動に紛れ込む。

## 棄却した代替案

### nil のとき nix-env を呼ぶ差し替え点を残す

注入し忘れが error にならずに通り、世代操作が契約越しであることを検証できない。
