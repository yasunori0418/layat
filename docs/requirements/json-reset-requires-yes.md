---
id: "REQ-2a613337-7646-4ced-8807-e43bca18acf3"
type: requirement
name: "reset --json は --yes を必須とし無ければ fail fast する"
status: stable
derives_from:
  - "UC-403fbe32-b146-401b-8b53-fe67c1e169c5"
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  Because a confirmation prompt for a destructive operation cannot be handled by a machine
  consumer, `reset --json` SHALL NOT show the prompt even on a TTY, and SHALL fail fast
  immediately with `status:"error"` and a non-zero exit when `--yes` is absent.
specification_ja: |
  破壊的操作の確認プロンプトは機械消費で扱えないため、`reset --json` は TTY でも
  プロンプトを出してはならず、`--yes` が無ければ即 `status:"error"` + 非ゼロで
  fail fast しなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:04:55+09:00"
---
# reset --json は --yes を必須とし無ければ fail fast する

## 仕様

**`reset --json` は `--yes` 必須**: 破壊的操作の確認プロンプトは機械消費で扱えないため、
TTY でもプロンプトを出さず、`--yes` が無ければ即 `status:"error"` + 非ゼロで fail fast
する。

`--json` なしの `reset` が確認プロンプトか `--yes` で同意を要求することは [reset は profile を触らない FS-only teardown で配置物を無い状態へ戻す](reset-fs-only-teardown.md) の
担当。

## 出典

`docs/spec.md`「CLI 仕様」→「出力ストリームと終了コード」→「outturn 準拠の `--json`
出力」のサブ項目「`reset --json` は `--yes` 必須」。

決定の実体は [reset は非対話の実行で --yes を必須にする](../adr/reset-confirmation-and-yes.md)。
