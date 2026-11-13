---
id: "REQ-9cb26ffd-071e-4c68-a6fc-faac6373b75e"
type: requirement
name: "project mode の root は git toplevel から解決し、config 相対も CWD 相対も採らない"
description: "project mode の root は指定が無ければ git rev-parse --show-toplevel から解決し、config ファイル相対も CWD 相対も採ってはならない"
section: root-resolution
status: stable
derives_from:
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  In project mode (`root = projectRoot`) the root SHALL be resolved by default from
  `git rev-parse --show-toplevel`, so that running from any subdirectory resolves to the
  same root; `--root <path>` SHALL state it explicitly for use outside git or to point at
  another root. Resolution relative to the config file SHALL NOT be adopted, since a flake
  source is copied into the store by Nix and becomes a store path, so that it does not
  hold. Resolution relative to the current working directory (`$PWD`) SHALL NOT be adopted
  either, since the placement destination would shift with the place of execution and
  break idempotency.
specification_ja: |
  project mode（`root = projectRoot`）の root は、指定が無いときは
  `git rev-parse --show-toplevel` から解決しなければならない（どのサブディレクトリから
  実行しても同じ root に解決される）。git 外で使う場合や別ルートを指す場合は
  `--root <path>` で明示しなければならない。config ファイル相対は採ってはならない
  （Nix で flake source が store にコピーされ store path 化するため成立しないため）。
  CWD（`$PWD`）相対も採っては
  ならない（実行場所で配置先がズレ冪等性を壊すため）。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:23:46+09:00"
---
# project mode の root は git toplevel から解決し、config 相対も CWD 相対も採らない

## 仕様

| 解決方法 | 備考 |
|---|---|
| `git rev-parse --show-toplevel`（既定）| どのサブディレクトリから実行しても同じ root に解決される |
| `--root <path>`（上書き）| git 外で使う場合や別ルートを指す場合に明示 |

- config ファイル相対は採らない（Nix で flake source が store にコピーされ store path 化する
  ため成立しない）
- CWD（`$PWD`）相対は採らない（実行場所で配置先がズレ冪等性を壊すため）

> **上は原文の写しで、規範は frontmatter が正**。`--root` が project mode に限らず全モードの
> 解決 root を一律上書きすることは [--root は全モード共通で解決 root を明示上書きする](root-override-flag.md)、明示時の profileDir キーイングは [profileDir は home と system のみ name 直キーとし、fixed root と --root 上書きは roothash でキーする](profiledir-keying.md)、
> `projectRoot` が marker であることは [root マーカーは `kind` を運ぶ入れ物でパス解決は engine が行う](root-markers-runtime-resolution.md) / [root は 3 マーカーと絶対パス文字列の union を取る](root-marker-kinds.md)、engine が外部コマンドとして
> `git` を叩くことは [engine が叩く外部コマンドは git のみに限る](engine-external-command-constraint.md)、project mode の配置物が ephemeral であることは [project mode の配置物は ephemeral とし、activation は git 状態に干渉しない](project-mode-ephemeral.md)、
> 世代の扱い（非公開・世代スキップ）は [project mode は世代を非公開にし、project / system mode は derivation 同一なら世代を積まず lstat ドリフト修復だけ行う](project-mode-generation-skip.md) の担当。

## 出典

`docs/spec.md`「root の解決」→「project mode（`root = projectRoot`）」節の表と、
同節の箇条書き第 1〜2 項。

決定の実体は [root は 3 つの marker で表す](../adr/three-root-markers.md)
（git toplevel 解決。config 相対と CWD 相対を棄却したのは旧決定）。
