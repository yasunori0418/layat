---
id: "ADR-1908099c-87f7-4df1-bda5-7731df7d9ed7"
type: adr
name: "engine が直接叩く外部コマンドを git に限る"
status: 採用
issues:
  - "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §4 を書き直した"
justifies:
  - "REQ-6c4e174a-4d16-477a-96ff-17cb4eb5b564"
  - "REQ-4ffda99a-7062-4c00-915f-70b525cb215b"
references:
  - "ADR-0006"
---
# ADR-1908099c-87f7-4df1-bda5-7731df7d9ed7: engine が直接叩く外部コマンドを git に限る

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-0006
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §4 を書き直した

## 背景

engine は git のほか、世代操作のために nix-env をサブプロセスで直接叩いていた。世代操作を世代
バックエンドの契約越しに行うようになると、engine が nix-env を呼ぶ理由は無くなる。

## 決定

engine が直接叩く外部コマンドは `git`（toplevel の解決）だけにする。`nix-env` は世代バックエンドの
nix 実装（nixprofile）の内側にだけ現れる。

外部コマンドの制約は REQ-6c4e174a-4d16-477a-96ff-17cb4eb5b564、内部で実行する nix コマンドの開示と
委譲は REQ-4ffda99a-7062-4c00-915f-70b525cb215b に置く。

## 根拠

engine が nix を直接呼ばなければ、nix への依存は世代バックエンドと manifest 生成器の内側に閉じる。
engine から契約を経ずに世代を操作する経路も残らない。

## 棄却した代替案

### engine が nix-env を引き続き直接叩く

世代操作を契約越しにしても engine に nix 固有の呼び出しが残り、契約を経ない経路ができる。
