---
id: "REQ-b74a118a-1272-44eb-944c-7725163211c6"
type: requirement
name: "engine は第三者依存ゼロの stdlib-only とし内部層に閉じる"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The engine SHALL be stdlib-only with zero third-party dependencies, and SHALL be an
  internal layer rather than a public module, so that what it depends on is confined to the
  Go standard library and the runtime commands it invokes. File locking, recursive copying,
  the manifest contract and error wrapping SHALL all be realised with the standard library.
  Which external commands it may invoke at run time is stated by REQ-6c4e174a-4d16-477a-96ff-17cb4eb5b564 and SHALL NOT
  be restated here.
specification_ja: |
  engine は第三者依存ゼロの stdlib-only とし、公開モジュールではなく内部層としなければ
  ならない（依存を Go 標準ライブラリと実行時に叩くコマンドだけに閉じるため）。ファイル
  ロック・再帰コピー・manifest 契約・エラーラップはいずれも標準ライブラリで実現しなければ
  ならない。
  実行時に叩いてよい外部コマンドは REQ-6c4e174a-4d16-477a-96ff-17cb4eb5b564 の規範であり、ここでは再掲しない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:32:25+09:00"
---
# REQ-b74a118a-1272-44eb-944c-7725163211c6: engine は第三者依存ゼロの stdlib-only とし内部層に閉じる

## 仕様

| コンポーネント | 依存 |
|---|---|
| `internal/`（配置エンジン = 内部層分離。公開モジュールではない）| **stdlib-only 厳守**（第三者依存ゼロ）。`syscall.Flock`・`filepath.WalkDir`+`io.Copy`+`os.Chmod`・`encoding/json`・`fmt.Errorf`+`%w`。`manifest.json` を入力に取り runtime に `nix`（profile）/ `git`（toplevel）をサブプロセスで要求 |

> **上は原文の写しで、規範は frontmatter が正**。実行時に叩く外部コマンドを `nix` と
> `git` に限ることは [engine が叩く外部コマンドは git のみに限る](engine-external-command-constraint.md)、engine が `manifest.json` を入力に取る 2 層構成は
> [layat は CLI とエンジンの 2 層で構成する](two-layer-architecture.md)、その `manifest.json` が唯一の安定契約であることは [manifest.json が Nix と engine の唯一の安定契約であり schemaVersion は 1 に固定する](manifest-single-stable-contract.md) の担当。
> 表が挙げる具体的な標準ライブラリ API 名は stdlib-only を満たす実装手段の例示であり、
> 規範文では「標準ライブラリで実現する」に留めた。

## 出典

`docs/spec.md`「依存関係」節の表の `internal/` 行。

決定の実体は [engine は JSON とエラーを標準ライブラリで扱う](../adr/json-and-errors-via-stdlib.md)・[engine は cmd/ と internal/ に置きライブラリ化しない](../adr/engine-internal-cmd-layout.md)・[flock は syscall.Flock で実装する](../adr/flock-via-syscall-flock.md) で、engine を stdlib-only の
内部層とすることを定めている。runtime に `nix` / `git` をサブプロセスで要求する形は
旧決定が定めていた（現在は [engine が直接叩く外部コマンドを git に限る](../adr/engine-external-command-git-only.md) が `git` だけに限る）が、旧決定は
stdlib-only 自体を決めていないため、側面の根拠として `justifies` は張らない（外部コマンドの
限定そのものの帰属は [engine が叩く外部コマンドは git のみに限る](engine-external-command-constraint.md) が担当する）。
