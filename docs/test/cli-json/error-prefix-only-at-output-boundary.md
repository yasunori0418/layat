---
id: "TC-cccd086a-5fb7-48d7-96ca-f4126e268396"
type: test_condition
name: "layat: を含む文字列リテラルが stderr へ書く出力のファイルの外に無い"
description: "internal/ と cmd/layat/ の非テストの Go で、layat: を含む文字列リテラルが stderr へ書く出力のファイルだけにあることを確かめる"
status: stable
mitigates:
  - "RISK-5a82656f-d1fa-4f71-8037-1daf307bfc95"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-14T23:41:04+09:00"
---
# layat: を含む文字列リテラルが stderr へ書く出力のファイルの外に無い

## 条件

`internal/` と `cmd/layat/` の非テストの Go を構文木として読み、文字列リテラルのうち `layat: ` を
含むものが、stderr へ書く行を整形する出力のファイルにだけあることを確かめる。エラーの文字列・
警告の書式・`generator.Error` の Message・sentinel のどれに紛れ込んでも、出力の関数を通ると
接頭辞が二重になるので、種類を問わず文字列リテラルの全部を見る。行の先頭だけでなく、文の途中に
`layat: ` を含むリテラルも拾う。

検出が実際に働くことも確かめる。`internal/` に `layat: ` を含むエラーを 1 つ足した状態で検査が
落ちることを、テストの外で一度確かめておく。

あわせて、出力の関数が書く行の形を確かめる。失敗・警告・通知・`--debug` の開示の行の形、
`errors.Join` で束ねたエラーを要素ごとに `layat: ` で始まる行へ分けること、engine の
conflict の列挙と undo の報告が警告の形ではなく `layat: <文>` で出ることを、apply と rollback の
両方の経路で見る。

`os.Is*` による判定の取りこぼしは、この条件では扱わない。lint が検出する。
