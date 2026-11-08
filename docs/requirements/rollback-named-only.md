---
id: "REQ-89c7baf9-9be0-417b-bd2d-b2e4edabe796"
type: requirement
name: "rollback は名指し必須で --all に対応しない"
status: stable
derives_from:
  - "UC-0b6f60cb-3e98-4ee7-8929-4d94a29f0af6"
specification: |
  `layat rollback` SHALL require an explicitly named config and SHALL NOT support `--all`,
  because rolling every config back at once is destructive and a footgun, and a failure
  partway through can leave the state inconsistent.
specification_ja: |
  `layat rollback` は名指し必須とし、`--all` に対応してはならない。全 config を一斉に
  戻すのは破壊的で footgun であり、途中失敗で状態が不揃いになり得るため。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:31:01+09:00"
---
# REQ-89c7baf9-9be0-417b-bd2d-b2e4edabe796: rollback は名指し必須で --all に対応しない

## 仕様

`rollback` は **名指し必須**（`--all` 非対応）。全 config を一斉に戻すのは破壊的で
誤用の危険が大きく、途中失敗で状態が不揃いになり得るため。

同じ理由による `reset` の名指し必須は [reset は名指し必須で profileDir 単位の blocking flock を取る](reset-named-only-and-flock.md) の担当。

## 出典

`docs/spec.md`「CLI 仕様」→「サブコマンド体系」の `rollback` 名指し必須の箇条書き。

決定の実体は [rollback に config の名指しを必須にする](../adr/rollback-requires-name-no-all.md)「rollback に config の名指しを必須にする」。
