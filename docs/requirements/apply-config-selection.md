---
id: "REQ-c2d44626-d8f4-446a-a80a-319a500129b4"
type: requirement
name: "apply の config 選択は name 省略で default・明示で単一・--all で全件"
status: stable
derives_from:
  - "UC-1c280dce-7c72-44c0-95ea-d06344f62a47"
specification: |
  When the name is omitted, `layat apply` SHALL apply `layat.default`, following the flake
  `default` convention, and SHALL fail with an error when `default` is not defined.
  Specifying `<name>` explicitly SHALL apply that config, and `--all` SHALL apply all of
  `layat.*`. That a profile is atomic per config is stated by REQ-1be4d678-959c-44d7-a346-44bfd95af56e and is not
  restated here.
specification_ja: |
  `layat apply` は name 省略時に `layat.default` を適用しなければならない（flake の
  `default` 慣例に倣う）。`default` が未定義ならエラーとしなければならない。`<name>` を
  明示すればその config を、`--all` で `layat.*` 全てを適用しなければならない。profile が
  config 単位で atomic で
  あることは REQ-1be4d678-959c-44d7-a346-44bfd95af56e の担当で、本 item では規定しない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# REQ-c2d44626-d8f4-446a-a80a-319a500129b4: apply の config 選択は name 省略で default・明示で単一・--all で全件

## 仕様

```bash
layat apply                     # name 省略時は layat.default を適用（flake の default 慣例。無ければエラー）
layat apply <name>              # layat.<name> をビルドし新世代を作って適用
```

`apply` の **name 省略時は `layat.default` を適用**する（flake の `default` 慣例に倣う。
`default` が未定義ならエラー）。`<name>` を明示すればその config を、`--all` で
`layat.*` 全てを適用する。profile は config 単位で atomic。

> **上は原文の写しで、規範は frontmatter が正**。`--all` の適用順・部分失敗時の挙動・
> root モードフィルタは本 item の規範ではなく、[apply --all は選択された config を並列に適用し、集約表示と results\[\] を辞書順にし、部分失敗でも続行して最後に集約する](apply-all-lexical-order-continue.md) / [--all は root モードフィルタで対象 config を絞れる](apply-all-root-mode-filter.md) の担当。
> profile の atomic 性そのもの（→ [1 config = 1 profile を全モードで自前に持つ](../adr/one-config-one-profile-all-modes.md)）は [世代は link farm derivation を layat 自前 profile へコミットして積み、前世代 manifest から古くなった配置を除去する](generation-mechanism-link-farm-manifest.md) の担当。

## 出典

`docs/spec.md`「CLI 仕様」→「サブコマンド体系」のコードブロックと、直後の箇条書き
第 1 項。

決定の実体は [apply の config は名前省略で default を選び、--all で全件を選ぶ](../adr/config-selection-default-and-all.md)（name 省略時は `layat.default`・未定義ならエラー・
一括適用は `--all`）。
