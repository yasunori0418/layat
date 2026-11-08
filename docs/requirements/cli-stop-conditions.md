---
id: "REQ-774cef80-2872-4ea1-937b-a0fbabc305a9"
type: requirement
name: "要求された操作が成立しないときは CLI がエラーで停止し、暗黙のフォールバックを採らない"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
  - "UC-0b6f60cb-3e98-4ee7-8929-4d94a29f0af6"
specification: |
  When the operation the user asked for does not hold, the CLI SHALL stop with an error and
  SHALL NOT fall back implicitly. It SHALL stop when no entrypoint can be discovered — no
  entrypoint file in the current working directory and no explicit `-f` — when the
  discovered entrypoint does not expose the requested `layat.<name>`, and, for `rollback`,
  when no previous generation exists, in which case it SHALL print an error message before
  stopping. Guessing another config, treating the situation as a no-op, or continuing with
  an empty manifest SHALL NOT be adopted in any of these cases, since each would silently
  produce a placement the user did not request.
specification_ja: |
  ユーザーが要求した操作が成立しないとき、CLI はエラーで停止しなければならず、暗黙の
  フォールバックを採ってはならない。entrypoint が発見できないとき（CWD に entrypoint
  ファイルが無く `-f` の明示も無い）、発見した entrypoint に指定の `layat.<name>` が存在
  しないとき、および `rollback` で前世代が存在しないときに停止しなければならない
  （`rollback` は停止前にエラーメッセージを出力しなければならない）。いずれの場合も、
  別 config の推測・no-op 扱い・空 manifest
  での続行を採ってはならない（ユーザーが要求していない配置を利用者に知らせずに生むため）。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:29:58+09:00"
---
# REQ-774cef80-2872-4ea1-937b-a0fbabc305a9: 要求された操作が成立しないときは CLI がエラーで停止し、暗黙のフォールバックを採らない

## 仕様

| 条件 | 動作 |
|---|---|
| `rollback` で前世代が存在しない | エラーメッセージを出力して停止 |
| `layat.<name>` が entrypoint に存在しない | CLI がエラーで停止 |
| entrypoint が発見できない（CWD に flake.nix/shell.nix/default.nix なし・`-f` 未指定）| CLI がエラーで停止 |

> **上は原文の写しで、規範は frontmatter が正**。entrypoint の探索順と `-f` による上書きは
> [entrypoint は CWD で flake.nix → shell.nix → default.nix の順に探し -f で上書きする](entrypoint-discovery-order.md)、`layat.<name>` のアドレッシングは [entrypoint は layat.<name> に named manifest を公開し CLI は形ごとの attr path で build する](named-manifest-addressing.md)、`apply` で name を省略した
> ときに `default` へ解決すること（および `layat.default` 未定義でのエラー停止）は
> [apply の config 選択は name 省略で default・明示で単一・--all で全件](apply-config-selection.md) / [config 名 default を慣例の解決先とし専用 layat 名前空間で packages を汚さない](default-config-name-and-namespace.md)、`rollback` が名指し必須で home / system mode 限定であることは
> [rollback は名指し必須で --all に対応しない](rollback-named-only.md) / [rollback と list-generations は home / system mode 限定にする](rollback-list-generations-home-only.md)、`gitignore` を非 project config へ与えたときのエラー停止は
> [gitignore は project mode 限定で非 project config を指定したらエラーで停止する](gitignore-project-mode-only.md)、`--manifest` と `-f` の併用エラー、および `--all` と合成するときの形の制約は [apply --manifest はビルド済み link-farm を engine へ直接適用し、--all と合成するときは name=link-farm を繰り返す](apply-manifest-direct.md)、
> experimental-features 未有効の案内エラーは [nix experimental-features は nix 生成器の前提条件とし、自動付与せず案内エラーで停止する](nix-experimental-features-prerequisite.md) の担当。停止時の終了コードは
> [終了コードは 0 = 成功 / 1 = 一般エラー / 2 = dryrun の conflict とする](exit-codes.md) の担当。

## 出典

`docs/spec.md`「エラー仕様」節の表のうち、要求された操作が成立しないことによる CLI 停止の
3 行（`rollback` で前世代なし・`layat.<name>` が entrypoint に不在・entrypoint が発見できない）。
同じ述語を満たす他の行（`layat.default` 未定義・`gitignore` の非 project config・`--manifest`
の併用・experimental-features 未有効）は、下記のとおり各担当 item が規範を持つため除く。

この 3 行が挙げる停止そのものに対応する決定を持つ ADR は無く、`docs/spec.md` が一次記述に
あたる。よって本 item に `justifies` は張られないが、これは張り忘れではない。前提となる
「CLI が entrypoint と named manifest を解決する」ことは旧決定が定めていた（現在は [生成器の契約を Discover / Roots / Build / DryBuild の 4 操作にする](../adr/generator-four-operations.md) の生成器が担う）が、旧決定は解決できなかった
ときの扱いを決めていないため、側面の根拠として `justifies` は張らない（前提そのものの帰属は
[entrypoint は CWD で flake.nix → shell.nix → default.nix の順に探し -f で上書きする](entrypoint-discovery-order.md) / [entrypoint は layat.<name> に named manifest を公開し CLI は形ごとの attr path で build する](named-manifest-addressing.md) が担当する）。
