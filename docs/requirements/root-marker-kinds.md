---
id: "REQ-37b56673-6e40-4a1b-a2a7-5d3c084e3e66"
type: requirement
name: "root は 3 マーカーと絶対パス文字列の union を取る"
section: lib-mkmanifest
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The type of `root` SHALL be a union of `string` (fixed at evaluation time) and marker
  (resolved at runtime). `layat.lib.projectRoot` SHALL select project mode (git toplevel
  at runtime, overridable with `--root`), `layat.lib.homeRoot` SHALL select home mode
  (`$HOME` at runtime), `layat.lib.systemRoot` SHALL select system mode (`/` at runtime),
  and an absolute path string SHALL select a fixed root determined at evaluation time.
specification_ja: |
  `root` の型は `string`（評価時固定）と marker（実行時解決）の union でなければ
  ならない。`layat.lib.projectRoot` は project mode（実行時に git toplevel・`--root` で
  上書き可）、`layat.lib.homeRoot` は home mode（実行時の `$HOME`）、
  `layat.lib.systemRoot` は system mode（実行時の `/`）、絶対パス文字列は固定 root
  （評価時に確定）へそれぞれ対応しなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:29:58+09:00"
---
# root は 3 マーカーと絶対パス文字列の union を取る

## 仕様

`root` の型は `string（評価時固定）| marker（実行時解決）` の union。

| `root` の値 | モード | root の解決 |
|---|---|---|
| `layat.lib.projectRoot` | project mode | 実行時に git toplevel（`--root` で上書き可）|
| `layat.lib.homeRoot` | home mode | 実行時の `$HOME`（standalone / HM 共通）|
| `layat.lib.systemRoot` | system mode | 実行時の `/` |
| 絶対パス文字列 | 固定 root | 評価時に確定する絶対パス（任意固定 root の seam）|

規範は frontmatter が正。

home mode と project mode は世代の扱いが異なる。この差異そのものは [project mode は世代を非公開にし、project / system mode は derivation 同一なら世代を積まず lstat ドリフト修復だけ行う](project-mode-generation-skip.md) の担当で、
本 item は root の値と対応するモードまでを規定する。

## 出典

`docs/spec.md`「lib API」→「`lib.mkManifest`」→「`root` の値」、および「root の解決」→
「system mode（`root = systemRoot`・将来）」節。後者は「root = `/`」「distro 構想の system
配置 seam」「今回の実装スコープ外」の 3 点からなり、規範として残る「root = `/`」が本 item の
表と同一のため、独立 item を立てず本 item に含めた。残る 2 点は、[rootKind に system を足しても schemaVersion を上げない](../adr/system-rootkind-without-version-bump.md)・[root は 3 つの marker で表す](../adr/three-root-markers.md) が system mode の
実装を決定したため規範に採らない。

> **marker の関数適用形（`subdir`）**: 適用形 `homeRoot { subdir = ".config"; }` は本 item の
> 担当範囲外で、その規範は [root マーカーの subdir 適用形は全 target への前置と等価な評価時の糖衣で、root を動かさない](root-marker-subdir-sugar.md) が持つ。適用形は root の値と
> `rootKind` を変えないので、本 item は原文の範囲（root の値と対応するモード）に留まる。
