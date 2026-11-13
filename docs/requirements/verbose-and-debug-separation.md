---
id: "REQ-0a123b89-0399-4f76-b988-56a5f7e0becf"
type: requirement
name: "冗長度は -v、デバッグは --debug に分離し --json と直交させる"
section: cli-output
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The placement report SHALL be written to stderr only when `-v` / `--verbose` is given
  (an opt-in on top of the default silence), including the skip notice and the
  `apply --all` completion summary. Disclosure of the nix commands run internally SHALL be
  separated into `--debug`, so that verbosity (`-v`) and debugging (`--debug`) are
  orthogonal. `--json` (stdout, machine-facing) SHALL be orthogonal to `-v` (stderr,
  human-facing) and usable together with it. `--quiet` SHALL NOT be provided, having been
  abolished along with the move to silence by default.
specification_ja: |
  配置レポートは `-v` / `--verbose` を付けたときだけ stderr に出さなければならない
  （既定沈黙の opt-in・skip 通知・`apply --all` 完了サマリも含む）。内部実行する nix
  コマンドの開示は `--debug` に分離しなければならず、冗長度（`-v`）とデバッグ
  （`--debug`）を直交させなければならない。`--json`（stdout・機械向け）は `-v`
  （stderr・人間向け）と直交し併用できなければならない。
  `--quiet` は既定沈黙化に伴い廃止し、提供してはならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# 冗長度は -v、デバッグは --debug に分離し --json と直交させる

## 仕様

**`-v` / `--verbose` を付けたときだけ配置レポートを stderr に出す**（既定沈黙の opt-in・
skip 通知・`apply --all` 完了サマリも含む）。**内部実行する nix コマンドの開示は
`--debug`** に分離する（冗長度＝`-v` と、デバッグ＝`--debug` を直交させる）。
`--json`（stdout・機械向け）は `-v`（stderr・人間向け）と直交・併用可。`--quiet` は
既定沈黙化に伴い廃止した。

既定沈黙そのものは [成功時はデフォルト沈黙とし warning と error は常時 stderr へ出す](silent-on-success.md)、`--help` 等での nix コマンド開示（透明性）は
[内部実行する nix コマンドを開示し世代の切替と GC は標準の nix コマンドへ委譲する](nix-command-transparency-and-delegation.md) の担当。

## 出典

`docs/spec.md`「CLI 仕様」→「出力ストリームと終了コード」の `-v` / `--verbose` の
箇条書き、およびグローバルフラグ表の `-v` / `--debug`。

決定の実体は [-v で配置レポートを opt-in で出す](../adr/verbose-opts-into-placement-report.md)・[内部コマンドの開示を --debug に分ける](../adr/debug-discloses-internal-commands.md)・
[指定が無いときは成功時に何も出力しない](../adr/silent-on-success.md)。`--json` との直交は [エラーはエンベロープに構造化し stderr のテキストと併存させる](../adr/structured-errors-with-stderr.md)。
