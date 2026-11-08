---
id: "REQ-8ef34101-8150-4124-92d5-94fabe6b5d90"
type: requirement
name: "成功時はデフォルト沈黙とし warning と error は常時 stderr へ出す"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
  - "UC-0b6f60cb-3e98-4ee7-8929-4d94a29f0af6"
  - "UC-403fbe32-b146-401b-8b53-fe67c1e169c5"
specification: |
  The default SHALL be silence on success. The success of `apply` / `reset` / `rollback`
  SHALL be conveyed by exit code 0, and the placement report (summary plus per-target
  lines), the try-lock skip notice and the `apply --all` completion summary SHALL NOT be
  emitted by default. Warnings (such as a foreign symlink) and errors SHALL always be
  written to stderr and SHALL be outside the scope of the silencing. The confirmation
  prompt and the abort notice of `reset` SHALL also remain. When `apply --all` finds no
  matching config, the notice `layat: apply --all: no matching configs` SHALL NOT be
  emitted by default and SHALL be written to stderr only with `-v`; the run SHALL exit
  with code 0 in both cases.
specification_ja: |
  成功時はデフォルト沈黙でなければならない。`apply` / `reset` / `rollback` の成功は
  終了コード 0 が語らなければならず、配置レポート（サマリ + per-target 行）・try-lock
  skip 通知・`apply --all` 完了サマリは指定が無いときは出してはならない。warning（foreign
  symlink 等）と error は常時 stderr に出さなければならず、沈黙の対象外でなければ
  ならない。`reset` の確認プロンプト・中止通知も存続させなければならない。`apply --all` で
  該当する config が無いときの通知（`layat: apply --all: no matching configs`）は、指定が
  無いときは出してはならず、`-v` のときだけ stderr に出さなければならない。どちらの場合も
  終了コードは 0 でなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-05T00:26:18+09:00"
---
# REQ-8ef34101-8150-4124-92d5-94fabe6b5d90: 成功時はデフォルト沈黙とし warning と error は常時 stderr へ出す

## 仕様

**成功時はデフォルト沈黙**（UNIX 哲学「沈黙は金」）。`apply` / `reset` / `rollback` の
**成功は終了コード 0 が語り、配置レポート（サマリ + per-target 行）・try-lock skip 通知・
`apply --all` 完了サマリは指定が無いときは出さない**。**warning（foreign symlink 等）と error は
常時 stderr に出す**（沈黙対象外）。`reset` の確認プロンプト・中止通知も存続する。

`apply --all` で該当する config が無いときの通知（`layat: apply --all: no matching configs`）も
沈黙の対象に含む。root フィルタ（`--project-root` / `--home-root` / `--system-root`）に合う
config が無いときがこれにあたる。指定が無いときは出さず、`-v` のときだけ stderr に出す。
どちらの場合も終了コードは 0 である。

`-v` による opt-in と、stdout 専有出力が沈黙の対象外である点は REQ-0a123b89-0399-4f76-b988-56a5f7e0becf /
REQ-fea038de-55eb-45ac-87fc-ec3a7287592a の担当。

## 出典

`docs/spec.md`「CLI 仕様」→「出力ストリームと終了コード」の第 1 箇条書き。

決定の実体は ADR-36bfbfbd-f1dd-4581-a75a-943d80f6ccf5「指定が無いときは成功時に何も出力しない」。
