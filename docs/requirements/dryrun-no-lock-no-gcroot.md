---
id: "REQ-7a71a049-5876-4cfc-a65e-44e9a0349856"
type: requirement
name: "--dryrun は root を解決するが flock も pending gcroot も取らない"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  `--dryrun` SHALL evaluate the root kind and resolve the root, in order to display the
  plan, but because it is read-only and does not place anything even if it builds the
  link-farm, it SHALL take neither the flock nor the pending gcroot (the out-link).
specification_ja: |
  `--dryrun` はプラン表示のため root kind を eval し root を解決しなければならないが、
  link-farm を build しても配置しない読み取り専用であるため、flock も pending gcroot
  （out-link）も取ってはならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:29:58+09:00"
---
# REQ-7a71a049-5876-4cfc-a65e-44e9a0349856: --dryrun は root を解決するが flock も pending gcroot も取らない

## 仕様

`--dryrun` は root の種別を eval し root を解決するが（プラン表示のため）、link-farm を
build しても配置しない読み取り専用なので、flock も pending gcroot（out-link）も取らない。

`apply --dryrun` が読み取り専用であること自体は [apply --dryrun は読み取り専用で conflict 検出時に非ゼロ終了する](apply-dryrun-readonly.md)、通常 apply が
out-link で gcroot を張ることは [実行フローの順序は eval 先行 → flock → build とし build をロック内に閉じる](exec-flow-eval-lock-build.md) の担当。

## 出典

`docs/spec.md`「CLI 仕様」→「実行フロー」の `--dryrun` の箇条書き。

決定の実体は [store path は out-link で受け取る](../adr/store-path-via-out-link.md)と [eval を先に行い、flock を取ってから build する](../adr/eval-first-then-flock-then-build.md)。
