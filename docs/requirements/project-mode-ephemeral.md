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
# project mode の配置物は ephemeral とし、activation は git 状態に干渉しない

## 仕様

project mode の配置物は **ephemeral**（コミット対象外）。activation は `.gitignore` に触れず
git 状態に干渉しない。

> **上は原文の写しで、規範は frontmatter が正**。`gitignore` サブコマンドが `.gitignore` を
> 書き込まず stdout へ列挙するだけであることは [gitignore は配置 target を stdout へ列挙するだけでファイルを書き込まない](gitignore-anchored-output.md)、copy target も ephemeral 扱いで
> 全 target を列挙することは [gitignore は method を区別せず copy target も含めて全 target を列挙する](gitignore-includes-copy-targets.md)、`gitignore` が project mode 限定であることは
> [gitignore は project mode 限定で非 project config を指定したらエラーで停止する](gitignore-project-mode-only.md) の担当。本 item は「配置物が ephemeral であり activation が git 状態へ
> 干渉しない」という原則そのものを規定する。

## 出典

`docs/spec.md`「root の解決」→「project mode（`root = projectRoot`）」節の箇条書き最終項。

決定の実体は [ephemeral 配置は git に干渉しない](../adr/ephemeral-no-git-interference.md)。
copy target まで ephemeral 扱いとする範囲は [project mode では copy と out-of-store も ephemeral として扱う](../adr/copy-and-out-of-store-ephemeral-in-project.md) が定めているが、本 item の規範に
copy 固有の記述は無く、その範囲は [gitignore は method を区別せず copy target も含めて全 target を列挙する](gitignore-includes-copy-targets.md) の担当のため [project mode では copy と out-of-store も ephemeral として扱う](../adr/copy-and-out-of-store-ephemeral-in-project.md) からの `justifies` は
張らない。
