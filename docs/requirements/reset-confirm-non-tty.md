---
id: "REQ-31dae599-f3a3-4bbe-b367-c955535265da"
type: requirement
name: "reset の確認プロンプトは stdin が TTY のときだけ出し、非 TTY で同意が無ければ即エラー停止する"
status: stable
derives_from:
  - "UC-403fbe32-b146-401b-8b53-fe67c1e169c5"
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The confirmation prompt of `layat reset` SHALL be shown only while stdin is a TTY. When
  stdin is not a TTY — under CI, direnv or a pipe — and neither `-y` nor `--yes` has been
  given, the command SHALL NOT show the prompt and SHALL stop immediately with an error
  and exit code 1, so as to prevent both hanging and an accidental deletion caused by
  empty input.
specification_ja: |
  `layat reset` の確認プロンプトは stdin が TTY のときのみ出さなければならない。非 TTY
  （CI / direnv / パイプ）かつ `-y` / `--yes` が未指定なら、プロンプトを出さず即エラー停止
  （exit 1）しなければならない（ハングと空入力による誤削除を防ぐため）。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# reset の確認プロンプトは stdin が TTY のときだけ出し、非 TTY で同意が無ければ即エラー停止する

## 仕様

確認プロンプトは **stdin が TTY のときのみ**出す。**非 TTY（CI / direnv / パイプ）かつ
`-y/--yes` 未指定なら、プロンプトを出さず即エラー停止（exit 1）**する（ハングと空入力誤削除を
防ぐ）。

> **上は原文の写しで、規範は frontmatter が正**。データ損失リスクのため確認プロンプトか
> `-y` / `--yes` で同意を要求すること自体は [reset は profile を触らない FS-only teardown で配置物を無い状態へ戻す](reset-fs-only-teardown.md) の担当で、本 item はその同意取得が
> TTY 有無でどう振る舞うかだけを規定する。`--json` 指定時に TTY でもプロンプトを出さない
> ことは [reset --json は --yes を必須とし無ければ fail fast する](json-reset-requires-yes.md) の担当。

## 出典

`docs/spec.md`「配置動作仕様」→「recopy・reset」節の reset 箇条書き第 1 項。

決定の実体は [reset は非対話の実行で --yes を必須にする](../adr/reset-confirmation-and-yes.md)。
