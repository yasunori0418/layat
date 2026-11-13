---
id: "REQ-27b75fe6-6c36-44a8-8cd3-5cc98043022a"
type: requirement
name: "subpath は src 内の相対パスとし、リポジトリ全体は省略で表して専用トークンを設けない"
section: entries-schema
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
specification: |
  The `subpath` field SHALL be a relative path selecting which path inside `src` is taken,
  and SHALL accept both a file and a directory. Omitting `subpath` SHALL be the canonical
  way to express the whole repository, with `subpath = "."` remaining a legal explicit
  equivalent. No dedicated token or marker for "the whole repository" other than `"."`
  SHALL be introduced, because `subpath` is a selection fixed at evaluation time whereas
  a marker is a container carrying a kind resolved at runtime.
specification_ja: |
  `subpath` フィールドは `src` 内のどのパスを取り出すかを表す相対パスとし、ファイル・
  ディレクトリのいずれも指定できなければならない。リポジトリ全体は `subpath` の省略で
  表すのを canonical としなければならず、`subpath = "."` は同義の明示形として合法に
  残らなければならない。`"."` 以外に
  「リポジトリ全体」を表す専用トークン / marker を設けてはならない。`subpath` は評価時に
  確定する subpath 選択であり、実行時解決の種別を運ぶ入れ物である marker とは相反する
  ためである。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# subpath は src 内の相対パスとし、リポジトリ全体は省略で表して専用トークンを設けない

## 仕様

`subpath` は string・省略可・デフォルト `"."`（リポジトリルート全体）。`src` 内の
どのパスを取り出すかを表す相対パスで、ファイル・ディレクトリどちらも指定できる。

- **リポジトリ全体は `subpath` を省略する**のが canonical。`subpath = "."` は同義の
  明示形。
- 「`"."` 以外で全体を表す専用トークン / marker」は設けない。`subpath` は評価時に確定
  する subpath 選択であり、糖衣 marker は marker パターン（実行時解決の種別を運ぶ
  入れ物）と相反するため（→ [リポジトリ全体は subpath の省略で表す](../adr/whole-repo-by-omitting-subpath.md)）。

```nix
# 省略 = リポジトリ全体（canonical）
subpath = ".";                  # リポジトリ全体（明示形）
subpath = "skills/nix";         # サブディレクトリのみ取り出す
subpath = "themes/dark.json";   # 単一ファイル
```

旧称 `source` からの改名理由（`src` との命名衝突の解消・→ [配置元の中のパスを全層で subpath と呼ぶ](../adr/subpath-naming-across-layers.md)）は原文の注記で、
規範ではない。旧名が評価時エラーになることは [entry submodule は strict とし未知キーと旧名を評価時エラーにする](entry-submodule-strict.md) が持つ。

## 出典

`docs/spec.md`「entries スキーマ仕様」→「フィールド定義（entry submodule）」→
`#### subpath`。
