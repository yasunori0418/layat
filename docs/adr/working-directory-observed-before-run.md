---
id: "ADR-d0949b61-59b7-4a96-8a47-998438ee872a"
type: adr
name: "workingDirectory はコマンドの処理の前に観測し、取得できない --json の実行はエンベロープを出さない"
description: "--json の実行では作業ディレクトリをフラグの解析の後・サブコマンドの処理の前に 1 度だけ観測して正規化せずに workingDirectory へ出し、取得できなければエンベロープを出さずに終了コード 1 で終わる"
section: cli-output
status: stable
sources:
  - id: "issue-826"
    resource: "https://github.com/yasunori0418/layat/issues/826"
    title: "#826 outturn への追従と workingDirectory の出力"
  - id: "outturn-adr-0028"
    resource: "https://github.com/yasunori0418/outturn/blob/628a4704dbbe/docs/adr/0028-working-directory-required-and-implicit-input.md"
    title: "yasunori0418/outturn ADR-0028"
origin: "outturn の ADR-0028 で workingDirectory が必須になったことへの追従（Issue #826、2026-11-18）"
justifies:
  - "REQ-ac859e73-4fe6-448e-b7d2-e1c08296249b"
  - "REQ-5c2e64c3-09a7-4ae8-b60c-4f1ccabce4fd"
references:
  - "ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-18T23:39:43+09:00"
---
# workingDirectory はコマンドの処理の前に観測し、取得できない --json の実行はエンベロープを出さない

- ステータス: stable
- 日付: 2026-11-18
- 関連: [--json は outturn 規約に常時準拠する](outturn-conformance-principle.md), GitHub Issue #826, yasunori0418/outturn ADR-0028
- 起点: outturn の ADR-0028 で workingDirectory が必須になったことへの追従（Issue #826、2026-11-18）

## 背景

outturn のエンベロープは、トップレベルに `workingDirectory` を必ず持つ。値は実行時の作業ディレクトリの
絶対パスである。schema は空でない文字列であることだけを確かめ、絶対パスであることは spec の MUST として
ツールに任せる。

layat の `--json` は、サブコマンドの処理の中で run を始め、コマンドの完了時にエンベロープを 1 度だけ出す。
作業ディレクトリをいつ観測するかと、観測できなかった実行で何を出すかを決める必要があった。

## 決定

`workingDirectory` に出す作業ディレクトリは、フラグを解析した後、サブコマンドの処理に入る前に 1 度だけ
観測する。観測するのは `--json` を付けた実行だけで、付けない実行ではこの観測をしない。観測した値は
symlink の解決などの正規化を掛けずに、そのまま `workingDirectory` に出す。サブコマンドの処理が
entrypoint の探索などで作業ディレクトリを読むことは、この観測とは別に扱う。

`--json` を付けた実行で作業ディレクトリを取得できなければ、サブコマンドの処理に入らずに終える。
エンベロープは出さず、stderr に `layat: cannot get the working directory: …` を書き、終了コード 1 で
終わる。作業ディレクトリを使わないコマンドも同じ扱いにする。

cobra が自動生成する `help` と `completion`（シェルごとのサブコマンドを含む）はエンベロープを出さないので、
`--json` を付けても作業ディレクトリを観測しない。作業ディレクトリを取得できない状態でも、これらは今までどおり
本文を出して成功する。

layat には作業ディレクトリを切り替えるフラグが無い。そのため、観測した値がそのまま outturn の求める
「実行時の作業ディレクトリ」になる。

## 根拠

サブコマンドの処理に入る前に観測すれば、どのコマンドの実行でも同じ時点の値が出る。エンベロープを出すか
どうかも、サブコマンドの処理に入る前に決まる。取得に失敗したら、そこで処理を止められる。
コマンドごとに観測すると、観測より前で失敗する処理や、観測を書き忘れるコマンドが出る。

値を正規化しないのは、利用者がシェルで見ている作業ディレクトリと同じ文字列を出すためである。
`--json` を付けない実行では値を使わないので、`workingDirectory` に出すための観測をしない。観測に失敗しても、既定の出力の契約は変わらない。

取得できない実行で代わりの値を出すと、outturn の MUST に違反した文書になる。消費側は値が正しいかどうかを
見分けられない。エンベロープを出さなければ、消費側は終了コードと stderr から失敗に気づける。フラグの解析の
失敗と同じ扱いになる。

## 棄却した代替案

### `$PWD` などの代わりの値を出す

`$PWD` は親のプロセスが渡した値で、実際の作業ディレクトリと一致する保証が無い。一致しない値を出すと、
outturn が求める「実行時の作業ディレクトリ」ではなくなる。

### 空文字列で出す

schema の `minLength: 1` に違反し、エンベロープが outturn に適合しなくなる。

### symlink を解決した値を出す

利用者がシェルで見ているパスと違う文字列になる。symlink の解決に失敗する場合も加わる。outturn は正規化を
求めていない。
