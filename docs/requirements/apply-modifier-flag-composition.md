---
id: "REQ-687e225f-5046-46db-88fb-f9e527a1e97a"
type: requirement
name: "apply 修飾フラグは --all と合成できる"
status: stable
derives_from:
  - "UC-1c280dce-7c72-44c0-95ea-d06344f62a47"
  - "UC-403fbe32-b146-401b-8b53-fe67c1e169c5"
specification: |
  `apply --all --recopy` SHALL be composable: `--recopy` SHALL be applied to each config
  selected by `--all` (with a filter such as `--project-root` when given), since
  `--recopy` is a modifier of apply and orthogonal to `--all`. `--all --backup[=<suffix>]`
  SHALL likewise be composable, `--backup` also being a modifier of apply.
specification_ja: |
  `apply --all --recopy` は合成可能でなければならない。`--recopy` は apply の修飾であり
  `--all` と直交するため、`--all`（必要なら `--project-root` 等のフィルタ）が選んだ各
  config に `--recopy` を適用しなければならない。`--all --backup[=<suffix>]` も同様に
  合成可能でなければならない（`--backup` も apply の修飾であるため）。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# REQ-687e225f-5046-46db-88fb-f9e527a1e97a: apply 修飾フラグは --all と合成できる

## 仕様

`apply --all --recopy` は**合成可**。`--all`（必要なら `--project-root` 等フィルタ）が
選んだ各 config に `--recopy` を適用する（`--recopy` は apply 修飾で `--all` と直交）。
`--all --backup[=<suffix>]` も同様に合成可（`--backup` も apply 修飾）。

各フラグ自体の規範は [apply --recopy は config 内の全 copy target を src から無条件に上書き再コピーする](apply-recopy.md)（`--recopy`）・[apply --backup は配置を塞ぐ記録外実体を rename 退避してから配置する](apply-backup-rename-aside.md)（`--backup`）の担当で、
本 item は `--all` との直交性のみを規定する。

## 出典

`docs/spec.md`「CLI 仕様」→「サブコマンド体系」の `apply --all --recopy` の箇条書き。

決定の実体は [apply の修飾フラグを --all と合成できるようにする](../adr/apply-modifiers-compose-with-all.md)「apply の修飾フラグを --all と合成できるようにする」と、`--backup` について同じ直交性を定めた [--backup を = 区切りで suffix を取る省略可能な値のフラグにする](../adr/backup-optional-value-flag.md)。
