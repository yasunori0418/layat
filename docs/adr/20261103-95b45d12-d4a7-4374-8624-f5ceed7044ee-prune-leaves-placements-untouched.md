---
id: "ADR-95b45d12-d4a7-4374-8624-f5ceed7044ee"
type: adr
name: "prune は配置物に触れない"
status: stable
sources:
  - id: "issue-580"
    resource: "https://github.com/yasunori0418/layat/issues/580"
    title: "#580"
origin: "次期マイルストーン計画の grilling（2026-07-04）。公開後の実運用で孤児系列が溜まり始めたことを消費側の要求と判断した。ADR-0034 §1c を書き直した"
justifies:
  - "REQ-c44433a1-7ee7-459a-9aae-7cc42166876f"
references:
  - "ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-95b45d12-d4a7-4374-8624-f5ceed7044ee: prune は配置物に触れない

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988, GitHub Issue #580
- 起点: 次期マイルストーン計画の grilling（2026-07-04）。公開後の実運用で孤児系列が溜まり始めたことを消費側の要求と判断した。ADR-0034 §1c を書き直した

## 背景

prune が削除する系列は、かつて apply が配置した config のものである。
系列を消すときに配置物まで撤去するのか、profile の状態だけを消すのかを決める必要がある。store
の回収をどこが受け持つかも同じ問いに含まれる。

## 決定

prune は profile の系列だけを消し、配置物には一切触れない。store の回収はこれまでどおり
`nix-collect-garbage` に任せる。prune は世代 symlink のディレクトリを消して GC root
を外すところまでを受け持ち、次の GC で store が回収できるようにする。

## 根拠

prune が対象にするのは root が消えた系列で、配置物は root ごと消えている。
マウントの境界などで一部が残っていても、それを片付けるのは root が戻ったときの apply / reset
の仕事になる。配置物に触れなければ、古くなった配置の除去の不変条件（記録どおりの layat 管理 symlink
だけを消す）を prune へ広げずに済む。

store の回収を prune に持ち込むと、GC を nix に任せる方針と二重になる。

## 棄却した代替案

### 系列を消すときに、記録に残る配置物も撤去する

root が無いので撤去する先が無い。root がアンマウントで一時的に見えないだけなら、
配置物を消すのは誤削除になる。

### prune を `reset --orphans` のように reset の拡張にする

reset は名指しした config の配置物を無い状態へ戻すコマンドで、配置物に触れず profile
の状態だけを消す prune とは不変条件が違う。同じコマンドに載せると、
どちらの不変条件に従うのかを利用者が取り違える。
