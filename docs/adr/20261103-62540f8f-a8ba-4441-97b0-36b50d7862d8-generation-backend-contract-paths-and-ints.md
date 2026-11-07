---
id: "ADR-62540f8f-a8ba-4441-97b0-36b50d7862d8"
type: adr
name: "世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う"
status: stable
issues:
  - "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §2a を書き直した"
justifies:
  - "REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b"
  - "REQ-706de717-4e47-471a-a1c0-448635be159c"
  - "REQ-1be4d678-959c-44d7-a346-44bfd95af56e"
  - "DSG-f8d9bba5-cfd5-4a19-b1fd-fd8ecc8a1814"
references:
  - "ADR-87a2cf34-9592-4c81-8980-246bca579ca0"
  - "ADR-bac005fb-fe4d-4a3b-b9e1-517a62fe46c2"
  - "ADR-02170bb7-d3b6-49ae-8426-de2113951637"
  - "ADR-efa7685c-6fa4-45df-b7a3-8f3f58e4118c"
  - "ADR-1ddf424a-db2c-4bcb-8823-bf7e4326b980"
  - "ADR-6687c0dc-957d-4913-99c6-168b0fe552ad"
  - "ADR-964e99a8-4aaa-46f6-b4ee-26270fc04649"
  - "ADR-cd139de9-06a4-49a2-8bea-4801c8acc518"
  - "ADR-494b6477-9d5b-4e29-879d-f7d16be5d466"
  - "ADR-5be36ff2-ed08-4e8e-8476-62ff2b08a9dd"
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
  at: "2026-11-03T23:00:56+09:00"
---
# ADR-62540f8f-a8ba-4441-97b0-36b50d7862d8: 世代操作はパスと整数だけを受け渡す 5 操作の契約越しに行う

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-87a2cf34-9592-4c81-8980-246bca579ca0, ADR-bac005fb-fe4d-4a3b-b9e1-517a62fe46c2, ADR-02170bb7-d3b6-49ae-8426-de2113951637, ADR-efa7685c-6fa4-45df-b7a3-8f3f58e4118c, ADR-1ddf424a-db2c-4bcb-8823-bf7e4326b980, ADR-6687c0dc-957d-4913-99c6-168b0fe552ad, ADR-964e99a8-4aaa-46f6-b4ee-26270fc04649, ADR-cd139de9-06a4-49a2-8bea-4801c8acc518, ADR-494b6477-9d5b-4e29-879d-f7d16be5d466, ADR-5be36ff2-ed08-4e8e-8476-62ff2b08a9dd, ADR-30f105c7-85a8-4fab-9dea-3277d932b524, ADR-419914ba-9896-4dd1-853b-fb514fcb9f38, ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1, ADR-58fdbad9-5ed7-469d-b6b2-6462b2173888, ADR-5fc58c8f-dc0e-4c1d-8c1e-17d7ac7f0385, ADR-74423bb8-1453-4689-a25e-30bf46536afc, ADR-b1e1dd5e-61a4-4d04-8389-137ef2347c56, ADR-d97154cd-a2e5-4853-a682-800ff9435c89, ADR-f90509f2-d75a-4162-baa5-a5cfdb8a7c30
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §2a を書き直した

## 背景

世代の操作（コミット・一覧・切り替え）と profile の基底は、nix profile の使い方に結び付いている。
manifest の取得側は生成器の契約越しになったが、世代側にも同じ境界が要る。

世代管理は将来、独立した CLI へ切り出す。
終着点では、パイプラインが `layat apply --json` の出力を世代管理 CLI へ渡し、世代管理 CLI は状態の識別子（link-farm のパス）の中身を解釈せずに受け渡す。
プロセス境界の契約は世代管理 CLI の側が定めるので、layat はそれまでの境界を決める必要がある。

## 決定

engine は世代操作を `internal/generation` の Go インターフェイス越しに行う。
契約は Commit・List・Switch・UserBase・SystemBase の 5 操作とする。
入出力はパスと整数だけにし、Go のコールバックを含めない。

世代管理 CLI の契約が確定したら、Commit・List・Switch の配線をその呼び出しへ置き換える。

5 操作の入出力は REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b、パッケージの構成は DSG-f8d9bba5-cfd5-4a19-b1fd-fd8ecc8a1814 に置く。

## 根拠

入出力がパスと整数だけなら、同じ境界をそのままプロセス境界に載せられ、世代管理を切り出すときも配線の置き換えで済む。
Go のコールバックを混ぜると、契約はプロセス内でしか成り立たない。

外部プログラムの規約ではなくプロセス内のインターフェイスに留めるのは、生成器の契約（ADR-87a2cf34-9592-4c81-8980-246bca579ca0）と同じ理由で、実装が nix 1 つの段階で外部の規約を決めると、使われない仕様が固まるため。

## 棄却した代替案

### 外部プログラムの規約をいま決める

実装は nix 1 つしかない。
規約を使う世代管理 CLI もまだ無い。
規約を定める主体は世代管理 CLI の側に置く。

### 差し替え点にコールバックを渡せる形にする

プロセス内でしか成り立たず、プロセス境界へ移すときに契約を作り直すことになる。
