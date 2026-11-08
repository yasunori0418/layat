---
id: "ADR-10c6dc33-d188-406b-beb4-d10ccbe3c645"
type: adr
name: "記録を書き直す操作ごとに linkFarm の値を定める"
status: 採用
issues:
  - "#580"
  - "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0058 §2b を書き直した"
justifies:
  - "REQ-9d622539-a1e5-4719-af51-e0cd85b66f35"
references:
  - "ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d"
  - "ADR-d97154cd-a2e5-4853-a682-800ff9435c89"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
  - "ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-10c6dc33-d188-406b-beb4-d10ccbe3c645: 記録を書き直す操作ごとに linkFarm の値を定める

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d, ADR-d97154cd-a2e5-4853-a682-800ff9435c89, ADR-fe035038-4415-4a1e-aa5c-509fdfce9955, ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93, GitHub Issue #580, #406
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0058 §2b を書き直した

## 背景

記録を書くのは apply だけではない。rollback と reset も FS を変えるので、記録を書き直す。apply なら
`linkFarm` は今回組んだ link-farm だが、rollback と reset は新しい link-farm を組まない。
書き直すときに `linkFarm` と `appliedAt` へ何を入れるかを決める必要がある。

## 決定

rollback で書き直すときの `linkFarm` は、戻り先の世代 N-1 の link-farm にする。reset
で書き直すときは記録済みの値を保ち、記録が無く profile から比較の基準を取ったときは、その current
世代の link-farm を入れる。`appliedAt` は、どの操作でもその操作の時刻に更新する。

## 根拠

`linkFarm` は、FS に残っている entry がどの link-farm から来たかを表す。rollback の後の FS は N-1
の link-farm の内容になるので N-1 を入れる。reset は entry を減らすだけで、残った entry
の出所は変わらないので値を保つ。こうしておけば、記録と profile の比較が操作の後も意味を保つ。

`appliedAt` は FS を最後に変えた時刻として読めるよう、その操作の時刻にする。

## 棄却した代替案

### reset で書き直すときに `linkFarm` を空にする

残った entry の出所が失われ、profile との比較ができなくなる。

### `appliedAt` を apply のときだけ更新する

rollback や reset の後に、FS を最後に変えた時刻として読めなくなる。
