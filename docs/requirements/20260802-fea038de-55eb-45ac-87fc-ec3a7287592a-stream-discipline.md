---
id: "REQ-fea038de-55eb-45ac-87fc-ec3a7287592a"
type: requirement
name: "stdout は機械可読出力を専有しレポートと warning は stderr へ出す"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  stdout SHALL be exclusively occupied by machine-readable output (the enumeration of
  `gitignore` and the plan of `apply --dryrun`). The placement report (placed / replaced /
  removed / skipped), warnings and the shellHook skip notice SHALL all go to stderr, so
  that `layat gitignore <name> >> .gitignore` and `layat apply <name> --dryrun | ...` can be
  piped safely. The scope of silencing on success SHALL be the placement report on stderr
  only, and the machine-readable output that occupies stdout SHALL always be emitted, both
  by default and under `-v`.
specification_ja: |
  stdout は機械可読出力（`gitignore` の列挙・`apply --dryrun` のプラン）を専有しなければ
  ならない。配置レポート（placed / replaced / removed / skipped）・warning・shellHook の
  skip 通知はすべて stderr へ出さなければならず、`layat gitignore <name> >> .gitignore` や
  `layat apply <name> --dryrun | ...` が安全にパイプできるようにする。成功時の沈黙化の
  対象は stderr の配置レポートのみとし、stdout 専有の機械可読出力は指定が無いときも `-v` のときも
  常に出さなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:31:01+09:00"
---
# REQ-fea038de-55eb-45ac-87fc-ec3a7287592a: stdout は機械可読出力を専有しレポートと warning は stderr へ出す

## 仕様

**ストリーム規律**: **stdout は機械可読出力を専有**する（`gitignore` の列挙・
`apply --dryrun` のプラン）。**配置レポート（placed / replaced / removed / skipped）・
warning・shellHook の skip 通知はすべて stderr**。これにより
`layat gitignore <name> >> .gitignore` や `layat apply <name> --dryrun | ...` が安全に
パイプできる。

**沈黙化の対象は stderr の配置レポートのみ**で、**stdout 専有の機械可読出力
（`apply --dryrun` の plan・`gitignore` の列挙）は指定が無いときも `-v` のときも常に出す**。
stdout 専有原則を貫き、`layat apply <name> --dryrun | ...` や
`layat gitignore <name> >> .gitignore` のパイプを壊さない。

`--json` 指定時に行指向 stdout を出さない（エンベロープが stdout を専有する）ことは
REQ-2353259f-5878-452a-8e11-3445de69abc2 の担当。

## 出典

`docs/spec.md`「CLI 仕様」→「出力ストリームと終了コード」の「ストリーム規律」箇条書きと、
同節末尾の「沈黙化の対象は stderr の配置レポートのみ」箇条書き。

決定の実体は ADR-9e328362-2e90-441a-8eca-40081ce71b1b「エラーはエンベロープに構造化し stderr のテキストと併存させる」（ストリーム規律）と ADR-36bfbfbd-f1dd-4581-a75a-943d80f6ccf5「指定が無いときは成功時に何も出力しない」
（沈黙化の対象を stderr のレポートに限り stdout 専有出力は常に出す）。
