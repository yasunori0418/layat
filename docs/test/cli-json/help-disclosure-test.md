---
id: "CASE-ad7299c1-e8af-4298-88a4-e1c61062ba30"
type: test_case
name: "help_disclosure_test.go — --help の開示ブロックと nix 呼び出しの突合・nix を実行しないサブコマンド"
status: stable
resource: "cmd/layat/help_disclosure_test.go"
covers:
  - "TC-b4262641-776c-427d-8331-1a94b29fa568"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:34+09:00"
---
# help_disclosure_test.go — --help の開示ブロックと nix 呼び出しの突合・nix を実行しないサブコマンド

## 対象

`cmd/layat/help_disclosure_test.go`

## 検証内容

- **開示ブロックと呼び出しの比較** — 呼び出しを記録する fake nix を PATH の先頭に置いて各
  サブコマンドを実行し、記録した引数を placeholder へ正規化する。`rootCmdLong` から切り出した
  ブロックとラベル単位で比べ、行の過不足と引数の形の違いを検出する
- **nix を実行しないサブコマンド** — last-applied 記録と root の消えた系列を置いた状態で、prune
  （`--dryrun` と `--yes`）と status を同じ fake nix の下で実行する。成功することと呼び出しが
  0 件であること、`--yes` だけが系列を削除すること、`--help` にその旨の 1 行があることを確かめる
