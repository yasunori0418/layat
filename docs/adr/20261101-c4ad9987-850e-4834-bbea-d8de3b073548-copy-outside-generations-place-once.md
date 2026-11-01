---
id: "ADR-c4ad9987-850e-4834-bbea-d8de3b073548"
type: adr
name: "copy は世代外で place-once"
status: 採用
origin: "ADR-0002 の決定「copy は世代から除外する」と「copy の扱い」節を書き直した"
justifies:
  - "REQ-d2277c7a-7992-49af-a9dc-4cc73843a6f9"
  - "REQ-77689c68-953c-4cbb-ab31-1ac1e4f5f2fe"
  - "REQ-16aef46b-7bb8-4ca1-b962-e9f3ed1fd1d2"
references:
  - "ADR-0020"
---
# ADR-c4ad9987-850e-4834-bbea-d8de3b073548: copy は世代外で place-once

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-0020, `docs/concept.md`, `docs/spec.md`
- 起点: ADR-0002 の決定「copy は世代から除外する」と「copy の扱い」節を書き直した

## 背景

entry の `method` には symlink と copy がある。symlink は store を指すだけなので、世代に載せれば
rollback で配置ごと戻せる。copy は target に実ファイルを置き、利用者がそれを手で編集する
（テーマや設定の一時調整）。編集された実ファイルを世代でどう扱うかを決める必要があった。

## 決定

copy は世代管理の対象外とし、target が無いときだけ実体化する place-once にする。target が
既にあれば上書きしない。stale 除去は copy target を消さず、entry から外れた copy は orphan として
警告するにとどめる。

store の更新を copy target へ反映する経路と copy target を消す経路は、利用者が明示したときに限る
（ADR-0020 の `apply --recopy` と `layat reset`）。規範の細目は REQ-d2277c7a-7992-49af-a9dc-4cc73843a6f9・
REQ-77689c68-953c-4cbb-ab31-1ac1e4f5f2fe・REQ-16aef46b-7bb8-4ca1-b962-e9f3ed1fd1d2 に置く。

## 根拠

copy を選ぶのは、store から切り離して自分で編集したいときである。layat が適用のたびに
上書きすると、その編集が消える。copy target を利用者が管理する副作用として扱えば、layat は
初回に置くだけで済み、編集との衝突が起きない。

世代に載せるには、store の外にある実ファイルのスナップショットを layat が持つ必要がある。
その管理は重く、利用者が持ち主だという位置付けとも合わない。

## 棄却した代替案

### copy も世代で追跡する（再実体化やスナップショットを取る）

rollback で copy も戻せるが、store の外のスナップショット管理を抱え込む。利用者の編集を
layat が巻き戻すことにもなり、copy を選んだ意図と矛盾する。

### 適用のたびに copy を上書きする

store の更新には常に追従できるが、利用者の編集が毎回消える。
