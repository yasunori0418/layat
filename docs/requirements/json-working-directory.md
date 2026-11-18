---
id: "REQ-ac859e73-4fe6-448e-b7d2-e1c08296249b"
type: requirement
name: "--json のエンベロープは workingDirectory に実行時の作業ディレクトリの絶対パスを出す"
description: "--json のエンベロープはトップレベルの workingDirectory に、フラグの解析の後に観測した作業ディレクトリの絶対パスを正規化せずに出す"
section: cli-json
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  When `--json` is given, the top level of the envelope SHALL carry `workingDirectory`,
  whose value SHALL be the absolute path of the working directory of the execution. The
  CLI SHALL observe the working directory for `workingDirectory` once, after flag parsing
  and before the subcommand runs, and SHALL write the observed value as is, without
  resolving symlinks or otherwise normalizing it. An execution without `--json` SHALL NOT
  observe the working directory for `workingDirectory`. When the working directory cannot be obtained in an execution with `--json`,
  the CLI SHALL NOT run the subcommand, SHALL NOT emit an envelope, SHALL write
  `layat: cannot get the working directory: …` to stderr, and SHALL exit with code 1. The
  CLI SHALL NOT substitute another value such as `$PWD` or an empty string.
specification_ja: |
  `--json` 指定時、エンベロープのトップレベルは `workingDirectory` を持たなければならず、
  値は実行時の作業ディレクトリの絶対パスでなければならない。CLI は `workingDirectory` に出す
  作業ディレクトリを、フラグの解析の後、サブコマンドの処理の前に 1 度だけ観測しなければならない。
  観測した値は symlink の解決などの正規化を掛けずに、そのまま出さなければならない。`--json` を
  付けない実行では、`workingDirectory` に出すための観測をしてはならない。`--json` を付けた実行で作業ディレクトリを
  取得できないとき、CLI はサブコマンドの処理を行ってはならず、エンベロープを出しては
  ならない。stderr に `layat: cannot get the working directory: …` を書き、終了コード 1 で
  終わらなければならない。`$PWD` や空文字列などの代わりの値を出してはならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-18T23:30:06+09:00"
---
# --json のエンベロープは workingDirectory に実行時の作業ディレクトリの絶対パスを出す

## 仕様

`--json` のエンベロープは、トップレベルの `workingDirectory` に実行時の作業ディレクトリの絶対パスを
出す。この値に使う作業ディレクトリは、フラグを解析した後、サブコマンドの処理に入る前に 1 度だけ
観測する。値は symlink を解決せず、観測したとおりの文字列で出す。`--json` を付けない実行では、
`workingDirectory` に出すための観測をしない。entrypoint の探索のように、サブコマンドの処理が
作業ディレクトリを読むことは、この観測とは別である。

`--json` を付けた実行で作業ディレクトリを取得できなければ、サブコマンドの処理に入らずに終える。
エンベロープは出さず、stderr に `layat: cannot get the working directory: …` を書き、終了コード 1 で
終わる。作業ディレクトリを使わないコマンド（status・prune など）も同じ扱いになる。`$PWD` や空文字列の
ような代わりの値は出さない。

outturn の schema が確かめるのは、`workingDirectory` が空でない文字列であることだけである。絶対パスで
あることと、観測した値と同じであることは、layat のテストで確かめる。

エンベロープのトップレベルのキーの全体は [--json は outturn 規約準拠のエンベロープを出す第 2 契約とする](json-outturn-envelope.md)、
エンベロープを出さない実行の全体は [エンベロープはコマンド完了時に 1 回だけ出し成立条件を満たさない実行では出さない](json-emit-timing.md) の担当。

## 出典

outturn の ADR-0028 で `workingDirectory` がエンベロープの必須のフィールドになった。

決定の実体は [workingDirectory はコマンドの処理の前に観測し、取得できない --json の実行はエンベロープを出さない](../adr/working-directory-observed-before-run.md)。
