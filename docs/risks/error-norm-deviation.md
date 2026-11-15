---
id: "RISK-5a82656f-d1fa-4f71-8037-1daf307bfc95"
type: risk
name: "エラーの扱いが規約から外れ、stderr の行に layat: が二重に付くか、存在しない対象の判定を取りこぼす"
description: "エラーの文字列に layat: が混ざると stderr の行で接頭辞が二重になり、os.Is* で判定すると wrap されたエラーで存在しない対象を見逃す"
likelihood: medium
impact: medium
level: medium
section: cli-json
status: stable
threatens:
  - "DSG-dae4180f-ffe0-4cbd-8bee-0bda331c8043"
  - "DSG-9f5f949f-8b0e-4679-8f70-e0901a3d63b8"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-14T23:43:17+09:00"
---
# エラーの扱いが規約から外れ、stderr の行に layat: が二重に付くか、存在しない対象の判定を取りこぼす

## リスク

stderr へ書く行の先頭の `layat:` は、cmd の出力の関数だけが付ける
（→ [stderr へ書く行は cmd の 1 箇所で `layat:` を付け、単発の失敗・`--all` の失敗行・status の行・警告が同じ整形を通る](../design/stderr-lines-prefixed-in-one-place.md)）。
エラーや警告の文字列を作る側が `layat:` を書き足すと、その文字列が出力の関数を通ったときに
接頭辞が二重になる。wrap された連鎖の途中に `layat:` が残る形も同じで、`--all` の失敗行や
status の行では、見出しの `layat: apply <name> failed:` に続くエラーの頭へもう一度 `layat:` が現れる。`--json` の
`errors[].message` にも、付けないはずの `layat:` がそのまま載る。エラーを作る箇所は
`internal/` と `cmd/layat/` の全体に散らばっており、新しい失敗を足すときに既存の書き方を
写すと、`layat:` も一緒に写りやすい。

もう 1 つの向きは判定の取りこぼしである。エラーは発生源で 1 回だけ wrap して伝える
（→ [エラーは op と対象が分かるいちばん下の関数で 1 回だけ wrap し、伝搬では素通しする](../design/error-wrap-once-at-origin.md)）。
wrap されたエラーに `os.IsNotExist` などの `os.Is*` を当てると、連鎖をたどらないので、存在しない
対象を存在しないと判定できない。存在しないときに続行するはずの処理が、失敗として止まる。
この向きは lint が検出する
（→ [golangci-lint でエラー処理の規約を検出する](../adr/golangci-lint-enforces-error-conventions.md)）。

## 影響

二重の接頭辞は出力の形が崩れるだけで、配置の結果は変わらない。`--json` の message を照合する
側は、接頭辞を含めた文字列で比べていれば食い違う。判定の取りこぼしは、存在しない対象を前にした
apply や reset を失敗として止めうる。どちらも利用者の環境を壊さず、修正後の layat で再実行すれば
回復するので impact を medium とする。どちらの向きも、出力を読むまで気付かれにくい。

likelihood を medium とするのは、エラーを作る箇所が多く、新しい失敗を足すたびに入り込む
余地がある一方で、`layat:` を含む文字列リテラルは
[layat: を含む文字列リテラルが stderr へ書く出力のファイルの外に無い](../test/cli-json/error-prefix-only-at-output-boundary.md) を確かめるテストが、
`os.Is*` は lint が、どちらも機械的に止めるため。文字列リテラルを組み立てて `layat:` を作る
書き方は、この検出の外に残る。
