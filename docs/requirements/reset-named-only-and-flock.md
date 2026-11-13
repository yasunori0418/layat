---
id: "REQ-a8edc58f-4adc-4637-b888-ab8ccc7e73e4"
type: requirement
name: "reset は名指し必須で profileDir 単位の blocking flock を取る"
section: cli-subcommands
status: stable
derives_from:
  - "UC-403fbe32-b146-401b-8b53-fe67c1e169c5"
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  `layat reset` SHALL require an explicitly named config and SHALL NOT support `--all`,
  because removing everything at once is a destructive footgun, consistently with the
  rejection of `rollback --all`; removing several configs SHALL be done by naming each of
  them in turn. `reset` SHALL acquire a blocking flock keyed on the resolved `profileDir`
  so that it is serialized against a concurrent apply / reset. Beyond the eval-first
  preliminary stage common to the non-building commands, `reset` SHALL additionally
  evaluate the entrypoint in order to read the entries.
specification_ja: |
  `layat reset` は名指し必須とし、`--all` に対応してはならない（一斉撤去は破壊的な
  footgun であり、`rollback --all` の却下と一貫させるため）。複数撤去は名指しを複数回
  行わなければならない。`reset` は解決後 `profileDir` 単位の blocking flock を取得して、
  並行する apply / reset と直列化しなければならない。非 build コマンド共通の eval 先行
  前段に加えて、`reset` は entries 読みのため entrypoint eval も行わなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# reset は名指し必須で profileDir 単位の blocking flock を取る

## 仕様

`reset` は **名指し必須（`--all` 非対応）**（一斉撤去は破壊的 footgun・`rollback --all`
却下と一貫）。複数撤去は名指しを複数回。**解決後 `profileDir` 単位の blocking flock を
取得**して並行 apply / reset と直列化する。profileDir 確定のため **rootKind 先取り eval →
root 解決**を build しないコマンドでも先行する（apply と共通の前段・`--root` 時は同じ
roothash キー）。`reset` はさらに entries 読みのため entrypoint eval も行う。

> **上は原文の写しで、規範は frontmatter が正**。このうち次の 2 点は本 item では
> 規範化しない。
>
> - 非 build コマンド一般の eval 先行（build しないコマンドでも rootKind 先取り eval →
>   root 解決を先行する）→ [非 build コマンドも eval 先行を共通前段に持つ](non-build-commands-eval-first.md)
> - `--root` 時に profileDir を roothash でキーすること → [profileDir は home と system のみ name 直キーとし、fixed root と --root 上書きは roothash でキーする](profiledir-keying.md)（`--root` が全モードの
>   解決 root を上書きすること自体は [--root は全モード共通で解決 root を明示上書きする](root-override-flag.md)）
>
> 本 item が規定するのは `reset` 固有の 3 点（名指し必須・profileDir 単位の blocking
> flock・共通前段に加えて行う entrypoint eval）に限る。

## 出典

`docs/spec.md`「CLI 仕様」→「サブコマンド体系」の `reset` 名指し必須の箇条書き。

決定の実体は [reset は config の名指しを必須にする](../adr/reset-requires-config-name.md)と、`--all` 非対応の一貫性を与える [rollback に config の名指しを必須にする](../adr/rollback-requires-name-no-all.md)。flock の直列化は [flock は明示実行で待ち、hook 実行でスキップする](../adr/flock-mode-explicit-wait-hook-skip.md)、
非 build コマンドの eval 先行は [eval を先に行い、flock を取ってから build する](../adr/eval-first-then-flock-then-build.md)。
