---
id: "DSG-dae4180f-ffe0-4cbd-8bee-0bda331c8043"
type: design
name: "stderr へ書く行は cmd の 1 箇所で `layat:` を付け、単発の失敗・`--all` の失敗行・status の行・警告が同じ整形を通る"
description: "stderr へ書く失敗・警告・通知の行は cmd/layat の 1 箇所で先頭に layat: を付け、internal/ のエラーと警告の文字列は layat: を持たず、--json の message にも付けない"
section: core-logic
status: stable
satisfies:
  - "REQ-fea038de-55eb-45ac-87fc-ec3a7287592a"
  - "REQ-8ef34101-8150-4124-92d5-94fabe6b5d90"
  - "REQ-9341fa5d-836e-4023-af53-cc7d273438d1"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-15T00:19:28+09:00"
---

# stderr へ書く行は cmd の 1 箇所で `layat:` を付け、単発の失敗・`--all` の失敗行・status の行・警告が同じ整形を通る

## 設計

stderr へ書く行の先頭の `layat:` は、`cmd/layat/` の 1 箇所に置いた整形の関数だけが付ける。単発の
コマンドの失敗、`--all` の失敗行、status の行、警告、通知は、どれもこの関数を通して書く
（→ [stdout は機械可読出力を専有しレポートと warning は stderr へ出す](../requirements/stream-discipline.md)、
[成功時はデフォルト沈黙とし warning と error は常時 stderr へ出す](../requirements/silent-on-success.md)）。

| 行の種類 | 形 |
|---|---|
| 失敗 | `layat: <エラー>` |
| 警告 | `layat: warning: <文>` |
| 通知（処理を止めない知らせ） | `layat: <文>` |

- **エラーと警告の文字列は `layat:` を持たない**: `internal/` は `layat:` を書かない。エラーの
  文字列の書式は [エラーの文字列は失敗の種類ごとの書式で書き、括弧の対象は原因が対象を持たないときだけ添える](error-message-format-by-failure-kind.md) に従う
- **`--json` の message には付けない**: エンベロープの `errors[].message` はエラーの文字列を
  そのまま持ち、`layat:` を付けない
  （→ [エンベロープのエラーは主体の有無で層を分けコードを分類する](../requirements/json-error-layering.md)）
- **複数行のエラー**: `errors.Join` で束ねたエラーは要素ごとに先頭へ `layat:` を付ける。要素の
  中の改行には付けない
- **cobra が返すエラー**: flag の解析の失敗など、cobra が返すエラーにも `layat:` を付ける
- **engine の警告と失敗の詳細**: cmd は engine へ警告を出す関数を常に渡し、警告は
  `layat: warning: <文>` で出す。engine は失敗の詳細（conflict の列挙と undo の報告）を警告とは
  別の出口へ流し、cmd はそれを `layat: <文>` で出す
- **`--all` での config 名の前置**: `apply --all` では、engine の警告と失敗の詳細を config ごとの
  出力に入れ、行の先頭に config 名を付けて `[<name>] layat: warning: <文>` と `[<name>] layat: <文>` で
  出す。config の失敗行は単発の失敗と同じ整形を通り、schemaVersion の食い違いとリポジトリの外の
  案内は全部の行の後に 1 回だけ出す
  （→ [apply --all では engine の警告と失敗の詳細も config ごとの出力にまとめ、config 名を前置する](../adr/apply-all-engine-warnings-in-config-output.md)）
- **nixgen の開示の行**: nixgen が `--debug` で出す開示の行も、cmd が渡した関数が書く。nixgen は
  `layat:` を書かない

## 出典

[`layat:` はエラーの文字列に持たせず、stderr へ書く場所で付ける](../adr/layat-prefix-added-where-stderr-is-written.md) の決定を、
cmd の出力の整形として書き下した。
