---
id: "REQ-e79178f5-5865-4444-a05d-3ab06f33cd6d"
type: requirement
name: "project mode の配置物は ephemeral とし、activation は git 状態に干渉しない"
status: stable
derives_from:
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  What is placed in project mode SHALL be ephemeral, outside the subject of a commit.
  Activation SHALL NOT touch `.gitignore` and SHALL NOT interfere with the state of git.
specification_ja: |
  project mode の配置物は ephemeral（コミット対象外）でなければならない。
  activation は `.gitignore` に触れてはならず、git 状態に干渉してはならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# REQ-e79178f5-5865-4444-a05d-3ab06f33cd6d: project mode の配置物は ephemeral とし、activation は git 状態に干渉しない

## 仕様

project mode の配置物は **ephemeral**（コミット対象外）。activation は `.gitignore` に触れず
git 状態に干渉しない。

> **上は原文の写しで、規範は frontmatter が正**。`gitignore` サブコマンドが `.gitignore` を
> 書き込まず stdout へ列挙するだけであることは REQ-a480c183-40ce-4201-93b5-65a7a59c1b9e、copy target も ephemeral 扱いで
> 全 target を列挙することは REQ-60787ed2-4176-4bdd-800f-1600c0315551、`gitignore` が project mode 限定であることは
> REQ-eaa8c0df-af44-4f52-9603-cd2bc22a67e9 の担当。本 item は「配置物が ephemeral であり activation が git 状態へ
> 干渉しない」という原則そのものを規定する。

## 出典

`docs/spec.md`「root の解決」→「project mode（`root = projectRoot`）」節の箇条書き最終項。

決定の実体は ADR-4dca6ef4-2791-48d5-aa82-7037630968a6「ephemeral 配置は git に干渉しない」。
copy target まで ephemeral 扱いとする範囲は ADR-1ea2b93a-f83f-46c6-b925-cb0b38f4fd3c が定めているが、本 item の規範に
copy 固有の記述は無く、その範囲は REQ-60787ed2-4176-4bdd-800f-1600c0315551 の担当のため ADR-1ea2b93a-f83f-46c6-b925-cb0b38f4fd3c からの `justifies` は
張らない。
