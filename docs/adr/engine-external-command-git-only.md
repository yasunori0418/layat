---
id: "ADR-1908099c-87f7-4df1-bda5-7731df7d9ed7"
type: adr
name: "engine が直接叩く外部コマンドを git に限る"
status: stable
sources:
  - id: "issue-406"
    resource: "https://github.com/yasunori0418/layat/issues/406"
    title: "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §4 を書き直した"
justifies:
  - "REQ-6c4e174a-4d16-477a-96ff-17cb4eb5b564"
  - "REQ-4ffda99a-7062-4c00-915f-70b525cb215b"
references:
  - "ADR-2119433e-de69-4447-85a5-575d58246007"
  - "ADR-521765d7-7e31-42f4-b548-979185a7586a"
  - "ADR-86a77232-dc99-4c04-bd6d-2127410e8828"
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
  - "ADR-62540f8f-a8ba-4441-97b0-36b50d7862d8"
  - "ADR-1cbacca1-1930-46ca-bf82-30310f82216e"
  - "ADR-4c8d459a-c371-4f0d-b009-16e2ca096e49"
  - "ADR-eb1aa25f-a0ca-4efe-bcd9-4cbe11368764"
  - "ADR-ffe1afcb-0b51-4b87-b698-0a6dea307bdd"
  - "ADR-8b8c9109-a479-4571-8b10-5a6079127441"
  - "ADR-efa7685c-6fa4-45df-b7a3-8f3f58e4118c"
  - "ADR-764ffff1-6b5a-4ed7-89d0-5700eb3ee883"
  - "ADR-9723dbb3-6659-4a69-ac4f-411c98bfa818"
  - "ADR-45380022-d959-4b22-aa05-5993ac1f3c87"
  - "ADR-a51897c5-f709-4b7a-8db2-61e6ebb0040c"
  - "ADR-f607b605-c69e-4c7e-ac89-75b7bce5bee1"
  - "ADR-88231ffc-4836-4252-be34-894944fd28c3"
  - "ADR-d15a216f-b5d6-4df9-9825-b23a341bec1d"
  - "ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9"
  - "ADR-dfdcc588-2e23-4d3b-ae17-b508a8d85ff0"
  - "ADR-553eea60-6ab4-4da7-93a9-3acf44b2fae3"
  - "ADR-30f105c7-85a8-4fab-9dea-3277d932b524"
  - "ADR-419914ba-9896-4dd1-853b-fb514fcb9f38"
  - "ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1"
  - "ADR-58fdbad9-5ed7-469d-b6b2-6462b2173888"
  - "ADR-5fc58c8f-dc0e-4c1d-8c1e-17d7ac7f0385"
  - "ADR-74423bb8-1453-4689-a25e-30bf46536afc"
  - "ADR-b1e1dd5e-61a4-4d04-8389-137ef2347c56"
  - "ADR-d97154cd-a2e5-4853-a682-800ff9435c89"
  - "ADR-f90509f2-d75a-4162-baa5-a5cfdb8a7c30"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:41+09:00"
---
# ADR-1908099c-87f7-4df1-bda5-7731df7d9ed7: engine が直接叩く外部コマンドを git に限る

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-2119433e-de69-4447-85a5-575d58246007, ADR-521765d7-7e31-42f4-b548-979185a7586a, ADR-86a77232-dc99-4c04-bd6d-2127410e8828, ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd, ADR-62540f8f-a8ba-4441-97b0-36b50d7862d8, ADR-1cbacca1-1930-46ca-bf82-30310f82216e, ADR-4c8d459a-c371-4f0d-b009-16e2ca096e49, ADR-eb1aa25f-a0ca-4efe-bcd9-4cbe11368764, ADR-ffe1afcb-0b51-4b87-b698-0a6dea307bdd, ADR-8b8c9109-a479-4571-8b10-5a6079127441, ADR-efa7685c-6fa4-45df-b7a3-8f3f58e4118c, ADR-764ffff1-6b5a-4ed7-89d0-5700eb3ee883, ADR-9723dbb3-6659-4a69-ac4f-411c98bfa818, ADR-45380022-d959-4b22-aa05-5993ac1f3c87, ADR-a51897c5-f709-4b7a-8db2-61e6ebb0040c, ADR-f607b605-c69e-4c7e-ac89-75b7bce5bee1, ADR-88231ffc-4836-4252-be34-894944fd28c3, ADR-d15a216f-b5d6-4df9-9825-b23a341bec1d, ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9, ADR-dfdcc588-2e23-4d3b-ae17-b508a8d85ff0, ADR-553eea60-6ab4-4da7-93a9-3acf44b2fae3, ADR-30f105c7-85a8-4fab-9dea-3277d932b524, ADR-419914ba-9896-4dd1-853b-fb514fcb9f38, ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1, ADR-58fdbad9-5ed7-469d-b6b2-6462b2173888, ADR-5fc58c8f-dc0e-4c1d-8c1e-17d7ac7f0385, ADR-74423bb8-1453-4689-a25e-30bf46536afc, ADR-b1e1dd5e-61a4-4d04-8389-137ef2347c56, ADR-d97154cd-a2e5-4853-a682-800ff9435c89, ADR-f90509f2-d75a-4162-baa5-a5cfdb8a7c30
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §4 を書き直した

## 背景

engine は git のほか、世代操作のために nix-env をサブプロセスで直接実行していた。世代操作を世代
バックエンドの契約越しに行うようになると、engine が nix-env を呼ぶ理由は無くなる。

## 決定

engine が直接実行する外部コマンドは `git` だけにする。用途は toplevel の解決に限る。`nix-env` は世代バックエンドの
nix 実装（nixprofile）の内側にだけ現れる。

外部コマンドの制約は REQ-6c4e174a-4d16-477a-96ff-17cb4eb5b564、内部で実行する nix コマンドの開示と
委譲は REQ-4ffda99a-7062-4c00-915f-70b525cb215b に置く。

## 根拠

engine が nix を直接呼ばなければ、nix への依存は世代バックエンドと manifest 生成器の内側に閉じ、
engine が契約を経ずに世代を操作するコードも残らない。

## 棄却した代替案

### engine が nix-env を引き続き直接実行する

世代操作を契約越しにしても、engine に nix 固有の呼び出しが残る。契約を経ずに世代を操作するコードも残る。
