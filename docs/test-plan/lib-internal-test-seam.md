---
id: "TP-403c55c7-d996-4951-8e6b-c3a7dddd387c"
type: test_plan
name: "lib.__internal で private helper をテストへ公開する"
status: stable
derives_from:
  - "SOL-9fcd1d6e-6204-42e6-92bb-1faf966f0b3e"
specification: |
  `lib.__internal` SHALL expose the private helpers (`escapesBase` / `pathChecks` /
  `anchorName` / `resolveEntry` / `farmEntries` / `anchorLines`) as a test seam for
  nix-unit and namaka. It SHALL NOT be a stable API and SHALL NOT be covered by
  backward-compatibility guarantees.
specification_ja: |
  `lib.__internal` は private helper（`escapesBase` / `pathChecks` / `anchorName` /
  `resolveEntry` / `farmEntries` / `anchorLines`）を nix-unit / namaka のテスト seam
  として公開する内部 API でなければならない。これは安定 API であってはならず、後方互換の
  保証対象に含めてはならない。
issues:
  - "#71"
  - "#289"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T22:03:12+09:00"
---
# TP-403c55c7-d996-4951-8e6b-c3a7dddd387c: lib.__internal で private helper をテストへ公開する

## 仕様

`lib.__internal` は private helper（`escapesBase` / `pathChecks` / `anchorName` /
`resolveEntry` / `farmEntries` / `anchorLines`）を nix-unit / namaka のテスト seam として
公開する内部 API で、安定 API ではない。

- 目的は評価テストから内部関数を直接呼び出せるようにすること。
- 安定 API ではないため、後方互換の保証対象に含めない。
- 列挙は seam の実内容を追跡する。唯一の SSOT は `lib/__internal.nix` の export で、helper を
  足したらそれを写している 5 箇所——本 item の 3 箇所（`specification` / `specification_ja` /
  本節）と `lib/default.nix` / `lib/manifest.nix` のコメント——を揃えて更新する。機械的な比較は
  無いので、追随し忘れはレビューで見つける（→ Issue #289 で `anchorLines` を追加）。

## 出典

`docs/spec.md`「lib API」。

> **本 item の根拠は ADR ではなく Issue #71**: 原文が根拠として挙げるのは `→ #71` のみで、
> `lib.__internal` を決めた ADR は存在しない。frontmatter の `issues` がこれを受けている。
> `justifies` を持たないのは張り忘れではなく、根拠が ADR 以外にあることによる。

> **本 item は `requirement` から `test_plan` へ移設した**（→ Issue #239）。
> `lib.__internal` を公開する唯一の動機は評価テストから内部関数を呼び出せる
> ようにすること、すなわちテスト対象へのアクセス手段の確保であり、ユーザーの使われ方
> （use_case）から導かれるプロダクトの振る舞いではないため、use_case を親に持てず接続漏れに
> なっていた（当時の判断は Issue #211）。`lib.__internal` を「利用者への API 契約の宣言」と
> 読んで `requirement` に残す立場もあり得たが、specification が SHALL NOT で安定性・後方互換の
> 保証除外を宣言している以上、契約として規定しているものは実質何も無く、テスト容易性の確保が
> 支配的であると判断した。
>
> **#238 の判断を #239 で改めた**。[非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する](e2e-harness-scope.md)・[エンベロープの outturn 適合を Go テストと E2E の両方で検証する](json-conformance-verification.md)・[NixOS / nix-darwin モジュールで使うときの実 activate は E2E ハーネスの対象外とする](nixos-vm-test-future.md) を移設した
> Issue #238 の時点では、本 item は「テスト容易性のための実装上の seam であってテスト計画
> そのものではない」と判断して見送り、接続先を別途決めることにしていた。しかし
> `docs/model.yaml` の test_plan 型は収容範囲に**テスト容易性の要求**を明示して含んでおり、
> 「テスト計画そのものではないから対象外」という当時の切り分けは型定義と食い違う。
> Issue #239 でこの判断を改め、[非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する](e2e-harness-scope.md)・[エンベロープの outturn 適合を Go テストと E2E の両方で検証する](json-conformance-verification.md)・[NixOS / nix-darwin モジュールで使うときの実 activate は E2E ハーネスの対象外とする](nixos-vm-test-future.md) と同類として
> test_plan へ移した。solution 直下で受けるため `derives_from` は [layat は nix store の物を任意パスへ置く配置プリミティブであり、設定を生成せずユーザーが配置を明示的に握る](../solution/layat-placement-primitive.md) を指す。
