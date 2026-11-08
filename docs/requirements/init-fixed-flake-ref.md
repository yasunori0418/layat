---
id: "REQ-cbd61281-64b0-4487-a4b7-ce76e70dc4f9"
type: requirement
name: "init のテンプレート参照はバイナリにハードコードした固定 flake ref とする"
status: stable
derives_from:
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The template reference of `layat init` SHALL be the fixed flake ref
  `github:yasunori0418/layat`, hardcoded into the binary, so that it works without
  depending on registry registration. Since `init` is for bootstrapping something new, a
  divergence between the CLI version and the template version (which always refers to the
  latest main) SHALL be accepted.
specification_ja: |
  `layat init` のテンプレート参照は、バイナリにハードコードした固定 flake ref
  `github:yasunori0418/layat` でなければならない（registry 登録に依存せず動くように）。
  `init` は新規 bootstrap 用途であるため、CLI 版と template 版のズレ（常に latest main
  参照）は許容しなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# REQ-cbd61281-64b0-4487-a4b7-ce76e70dc4f9: init のテンプレート参照はバイナリにハードコードした固定 flake ref とする

## 仕様

**テンプレート参照はバイナリにハードコードした固定 flake ref
`github:yasunori0418/layat`**。registry 登録に依存せず動く。`init` は新規 bootstrap 用途で、
CLI 版と template 版のズレ（常に latest main 参照）は許容する。apply 時の
`schemaVersion` 整合は project mode の devShell 同梱 pin が担う（[layat CLI は PATH 常駐の一次 UX で、project mode は devShell 同梱を canonical とする](cli-primary-ux-installation.md)）。

## 出典

`docs/spec.md`「CLI 仕様」→「`layat init`（テンプレート展開）」の箇条書き第 3 項。

決定の実体は [templates と layat init は固定 ref で導入する](../adr/templates-and-init-fixed-ref.md)「templates と layat init は固定 ref で導入する」。
