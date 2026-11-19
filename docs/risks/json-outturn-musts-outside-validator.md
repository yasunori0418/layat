---
id: "RISK-7b73d79d-9ac0-4f2a-88e8-9e03a565d46f"
type: risk
name: "outturn-validate が確かめない MUST から外れても適合の確認では見つからない"
description: "outturn の MUST のうち outturn-validate が確かめないもの（workingDirectory が実行時の作業ディレクトリの絶対パスであること、暗黙の入力の解決結果を info に記録すること）から外れた文書は、適合の確認を通ったまま消費側へ届く"
section: cli-json
status: stable
threatens:
  - "REQ-ac859e73-4fe6-448e-b7d2-e1c08296249b"
  - "REQ-50af28ba-ddf0-4488-abd7-244f1a5597cb"
likelihood: low
impact: medium
level: low
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-18T23:55:09+09:00"
---
# outturn-validate が確かめない MUST から外れても適合の確認では見つからない

## リスク

outturn の MUST には、schema と lint で確かめられないものがある。outturn-validate と Go の適合チェッカは、
`workingDirectory` が空でない文字列であることだけを確かめる。絶対パスであることと、実行時の作業ディレクトリと
同じ値であることは確かめない。layat の出力がこの MUST から外れても、適合の確認は成功する。

外れ方は 2 つある。

- 観測した値の代わりに `.` のような相対パスを出す
- 観測した値と違う値を出す。観測した値を run へ写し忘れて別の値になるほか、`$PWD` のような代わりの値や、
  symlink を解決した値、テスト用の固定値を出す

暗黙の入力の記録も、outturn-validate は確かめない。`info` のキーがあってもなくても、文書は適合する。
外れ方は 2 つある。

- 暗黙の入力を記録し漏らす。`-f`・`--generator`・`--root` を省いた実行で、`entrypoint`・`generator`・
  `root` のキーを出さない。コマンドや実行の形（`--all`・`--dryrun`・失敗した実行）によっては記録を書き忘れる
- 明示した値を記録してしまう。明示引数で与えた値のキーを出したり、`--manifest` の実行で `entrypoint` や
  `generator` を出したりする。出すキーが無い実行で、空の `info` を出す

## 影響

消費側は、`workingDirectory` を起点にして文書の中の相対パスを解決したり、どの作業ディレクトリでの実行かを
記録したりする。暗黙の入力の記録からは、どの値を layat が自分で決めたかを読み取り、明示引数に渡して
同じ配置を再現する。値やキーの有無が違っていても、文書は適合したまま正常終了で渡るので、消費側は誤った値を使う。impact を
medium とするのは、損なわれるのが出力の内容であって FS の配置ではなく、layat を直せば再実行で正しい文書が
返るためである。likelihood を low とするのは、値を観測する箇所と出す箇所がそれぞれ 1 つで、Go テストと E2E が
値を確かめているためである。暗黙の入力の記録は、記録する箇所が対象のコマンドごとにあるが、Go テストが
コマンドごとにキーの有無と値を確かめている（→ [outturn-validate が確かめない MUST を layat のテストで確かめる](../test/cli-json/outturn-musts-outside-validator.md)）。
