---
id: "ADR-fd3353cf-00af-4343-bd3f-3369c21d7194"
type: adr
name: "reset は撤去できなかった entry を残して記録を書き直す"
status: stable
sources:
  - id: "issue-580"
    resource: "https://github.com/yasunori0418/layat/issues/580"
    title: "#580"
  - id: "issue-406"
    resource: "https://github.com/yasunori0418/layat/issues/406"
    title: "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0058 §4c を書き直した"
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
# ADR-fd3353cf-00af-4343-bd3f-3369c21d7194: reset は撤去できなかった entry を残して記録を書き直す

- ステータス: stable
- 日付: 2026-11-03
- 関連: [project mode では同じ link farm なら世代を積まず drift だけを直す](project-mode-generation-skip-and-drift-repair.md), [rollback は FS を収束させてからポインタを動かす](rollback-baseline-and-pointer-last.md), [1 config = 1 profile を全モードで自前に持つ](one-config-one-profile-all-modes.md), [古くなった配置の除去は manifest の差分で行う](stale-removal-by-manifest-diff.md), GitHub Issue #580, #406
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0058 §4c を書き直した

## 背景

reset は名指しした config の配置物を撤去する。foreign や drift として撤去せずに残す entry があり、
途中で失敗することもある。reset は undo journal を持たないので、失敗しても FS は reset
が残した状態のままになる。

記録だけが残り profile が無い環境もある。この場合、flock の鍵になる profileDir が存在しない。

## 決定

reset は、撤去対象のうち target が消えた entry だけを除いた manifest で記録を書き直す。FS
操作が途中で失敗しても同じ規則で書き直し、全 entry が除かれたら記録ファイルを削除する。profile
が無く記録だけがあるときは、apply と同じ手順で profileDir（roothash 系列では backref も）
を作ってから flock を取る。

書き直しの失敗の報告を含む細目は [apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](../requirements/last-applied-record.md) に置く。

## 根拠

記録は FS に何が残っているかを表す。撤去できた entry だけを除けば、残した entry を次の apply や
reset が引き続き追える。reset には undo journal が無いので、失敗したときも FS
に合わせて記録を書き直すしかない。

flock の鍵は profileDir なので、記録だけが残った環境でも鍵を作れば、
同時に動くほかの実行と排他できる。

## 棄却した代替案

### reset の後は記録ファイルを丸ごと削除する

foreign や drift として残した entry と、途中の失敗で手が届かなかった entry を追えなくなる。

### 途中で失敗したら記録に触れない

撤去済みの entry が記録に残り、記録と FS がずれる。

### profile が無いときは flock を取らずに撤去する

同時に動く apply と競合する。
