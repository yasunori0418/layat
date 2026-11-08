---
id: "REQ-eaa8c0df-af44-4f52-9603-cd2bc22a67e9"
type: requirement
name: "gitignore は project mode 限定で非 project config を指定したらエラーで停止する"
status: stable
derives_from:
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  `layat gitignore` SHALL be restricted to project mode. A bare `gitignore <name>` SHALL
  accept only a project mode config, and SHALL stop with an error when a non-project
  config (home / fixed) is specified, because the anchored output form presupposes that
  root = git toplevel = the location of `.gitignore`, which is meaningless under home /
  fixed. This is symmetric with `rollback` / `list-generations` being restricted to home
  mode and system mode.
specification_ja: |
  `layat gitignore` は project mode 限定でなければならない。単体の `gitignore <name>` も
  project mode の config のみを受理し、非 project config（home / fixed）を指定したら
  エラーで停止しなければならない（出力のアンカー形式が root = git toplevel =
  `.gitignore` の置き場所を前提とし、home / fixed では意味を成さないため）。
  これは `rollback` / `list-generations` が home / system mode 限定であるのと対称である。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# REQ-eaa8c0df-af44-4f52-9603-cd2bc22a67e9: gitignore は project mode 限定で非 project config を指定したらエラーで停止する

## 仕様

`gitignore` は **project mode 限定**。単体 `gitignore <name>` も project mode の config
のみ受理し、**非 project config（home / fixed）を指定したらエラーで停止する**（出力の
アンカー形式が git toplevel = `.gitignore` 置き場所を前提とし、home / fixed では意味を
成さないため）。`rollback` / `list-generations` が home / system mode 限定なのと対称。

アンカー形式そのものは [gitignore は配置 target を stdout へ列挙するだけでファイルを書き込まない](gitignore-anchored-output.md)、home / system mode 限定の側は [rollback と list-generations は home / system mode 限定にする](rollback-list-generations-home-only.md) の担当。

## 出典

`docs/spec.md`「CLI 仕様」→「サブコマンド体系」の `gitignore` project mode 限定の
箇条書き。

決定の実体は [配置 target を gitignore コマンドで stdout へ列挙する](../adr/gitignore-command.md)。
