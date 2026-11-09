---
id: "REQ-05abce3e-9797-432b-b93f-37c55d09afde"
type: requirement
name: "rollback と list-generations は home / system mode 限定にする"
status: stable
derives_from:
  - "UC-0b6f60cb-3e98-4ee7-8929-4d94a29f0af6"
  - "UC-6b8d31ea-3d5a-4334-aac7-8706df190e7d"
specification: |
  `layat rollback` and `layat list-generations` SHALL be restricted to home mode and system
  mode. In project mode and fixed mode generations SHALL be kept as an internal mechanism and
  SHALL NOT be exposed to the user. `list-generations --all` SHALL list the generations of
  every home mode config and then of every system mode config, and SHALL be read-only. A
  system base that does not exist SHALL be listed as empty, and one that cannot be read SHALL
  be skipped with a warning.
specification_ja: |
  `layat rollback` と `layat list-generations` は home mode と system mode に限らなければならない。
  project mode と fixed mode では世代を内部機構に留め、ユーザーに公開してはならない。
  `list-generations --all` は home mode の全 config、続けて system mode の全 config の世代を
  一覧しなければならず、読み取り専用でなければならない。system 基底が存在しなければ空として
  扱わなければならず、読めなければ warning を出してその基底の走査を省かなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:29:58+09:00"
---
# rollback と list-generations は home / system mode 限定にする

## 仕様

```bash
layat rollback <name>           # 前世代へ戻す（home / system mode 限定・名指し必須）
layat list-generations <name>   # 世代一覧を表示（home / system mode 限定）
layat list-generations --all    # home mode、続けて system mode の全 config の世代を一覧
```

`rollback` / `list-generations` は **home / system mode 限定**。project / fixed mode は世代を
内部機構に留めユーザーに公開しない。`list-generations --all` は user 基底（home mode）の全
config、続けて system 基底（system mode）の全 config の世代を一覧する（読み取り専用）。
system 基底が無ければ空として扱い、読めなければ warning を出してその基底を走査しない。

`rollback` が名指し必須である点は [rollback は名指し必須で --all に対応しない](rollback-named-only.md) の担当。

> **原文「standalone（CLI）」節の残る規範の所在**: 同節の第 1 文（`layat apply <name>` の
> 明示実行と、CLI による entrypoint 発見・manifest ビルド・エンジン駆動）は [layat は CLI とエンジンの 2 層で構成する](two-layer-architecture.md) /
> [entrypoint は CWD で flake.nix → shell.nix → default.nix の順に探し -f で上書きする](entrypoint-discovery-order.md)、世代管理の機構そのものは [世代は link farm derivation を layat 自前 profile へコミットして積み、前世代 manifest から古くなった配置を除去する](generation-mechanism-link-farm-manifest.md) の担当。

## 出典

`docs/spec.md`「CLI 仕様」→「サブコマンド体系」の `rollback` / `list-generations` の
箇条書き、および「モジュール別動作仕様」→「standalone（CLI）」節の第 2 文（home mode で
nix profile による世代管理を行い `rollback` / `list-generations` を提供する）。後者は
独立 item を立てず本 item に含めた。

決定の実体は [module と project では世代操作を公開しない](../adr/module-and-project-hide-generation-ops.md)と、
`list-generations --all` を追加した [list-generations --all は home mode と system mode の全 config を対象にする](../adr/list-generations-all-home-and-system.md)。home / system への公開は [module と project では世代操作を公開しない](../adr/module-and-project-hide-generation-ops.md)。
