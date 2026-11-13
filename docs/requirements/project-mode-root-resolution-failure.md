---
id: "REQ-6506bc82-d1e1-4dbf-8c57-d5d1babf218a"
type: requirement
name: "project mode で git から root を解決できないときは engine 実行時に停止する"
description: "project mode で git から root を解決できないときは、git リポジトリの外でも git が PATH に無いときも、別の root で続行せず engine 実行時にエラーで停止しなければならない"
section: errors
status: stable
derives_from:
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  In project mode, when the engine cannot resolve the root from git, it SHALL stop with an
  error at run time, rather than continuing with some other root, since placing into a root
  other than the one intended would scatter files outside the project. This SHALL apply
  both when the current directory is outside a git repository and `--root` is not given, so
  that resolving the toplevel fails, and when `git` is not on PATH.
specification_ja: |
  project mode において engine が git から root を解決できないとき、別の root で続行せず
  実行時にエラーで停止しなければならない（意図と異なる root へ配置すればプロジェクト外へ
  ファイルを撒くため）。git リポジトリ外かつ `--root` 未指定で toplevel の解決に失敗する
  場合と、`git` が PATH に無い場合の双方に適用しなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:29:58+09:00"
---
# project mode で git から root を解決できないときは engine 実行時に停止する

## 仕様

| 条件 | 動作 |
|---|---|
| project mode で git リポジトリ外かつ `--root` 未指定（git toplevel 解決失敗）| engine 実行時にエラーで停止 |
| project mode で `git` が PATH に無い | engine 実行時にエラーで停止 |

> **上は原文の写しで、規範は frontmatter が正**。project mode の root を git toplevel から
> 解決し CWD 相対も config 相対も採らないことは [project mode の root は git toplevel から解決し、config 相対も CWD 相対も採らない](project-mode-root-resolution.md)、`--root` が全モード共通の
> 上書きであることは [--root は全モード共通で解決 root を明示上書きする](root-override-flag.md)、engine が外部コマンドとして `git` を叩くことは
> [engine が叩く外部コマンドは git のみに限る](engine-external-command-constraint.md) の担当。

## 出典

`docs/spec.md`「エラー仕様」節の表の project mode の 2 行。

この 2 行が挙げる解決失敗時の停止そのものに対応する決定を持つ ADR は無く、
`docs/spec.md` が一次記述にあたる。よって本 item に `justifies` は張られないが、これは
張り忘れではない。前提となる「project root を git toplevel に置く」ことは
[root は 3 つの marker で表す](../adr/three-root-markers.md)が定めるが、同 ADR は
解決に失敗したときの扱いを決めていないため、側面の根拠として `justifies` は張らない
（前提そのものの帰属は [project mode の root は git toplevel から解決し、config 相対も CWD 相対も採らない](project-mode-root-resolution.md) が担当する）。
