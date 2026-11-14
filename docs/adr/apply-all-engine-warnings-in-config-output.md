---
id: "ADR-a97c014b-efdd-4643-9247-004c4f268191"
type: adr
name: "apply --all では engine の警告と失敗の詳細も config ごとの出力にまとめ、config 名を前置する"
description: "apply --all で engine が出す警告と失敗の詳細を config ごとの出力に入れ、行の先頭に config 名を付けて、辞書順の集約に含める"
section: cli-commands
status: stable
sources:
  - id: "issue-677"
    resource: "https://github.com/yasunori0418/layat/issues/677"
    title: "#677 Go のエラー処理を規約に揃える"
origin: "Issue #677 の grilling（2026-11-14）"
justifies:
  - "REQ-4cbd9a0d-9f94-4747-8881-56020dc6d5af"
references:
  - "ADR-af37a636-a798-4c03-8fa4-e73ed6375a92"
  - "ADR-e7a57f51-41da-460c-8f51-c84e94643f3d"
  - "ADR-c109ac0c-d5df-4a15-acf0-d6d85ce5402e"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-15T00:15:03+09:00"
---

# apply --all では engine の警告と失敗の詳細も config ごとの出力にまとめ、config 名を前置する

- ステータス: stable
- 日付: 2026-11-15
- 関連: GitHub Issue #677, [apply --all の集約は config 名の辞書順に並べる](apply-all-lexicographic-order.md), [apply --all の配置を config 単位で並列に行う](apply-all-parallel-placement-per-config.md), [apply --all の build を worker pool で並列に実行する](apply-all-parallel-build-worker-pool.md)
- 起点: Issue #677 の grilling（2026-11-14）

## 背景

`apply --all` は config ごとの出力（`-v` のレポート・失敗と skip の行）を config ごとに溜め、
全部の config が終わってから辞書順に出す。engine が配置中に出す警告と失敗の詳細（conflict の
列挙と undo の報告）はこの対象から外れ、発生順に stderr へ直接流れていた。config は並列に
動くので、行がどの config から出たのか分からなかった。

## 決定

- `apply --all` では、engine の警告と失敗の詳細を config ごとの出力に入れ、辞書順の集約に含める
- 行の先頭に config 名を付ける。警告は `[<name>] layat: warning: <文>`、失敗の詳細は
  `[<name>] layat: <文>` とする
- nix の出力と `--debug` の開示行は今までどおり発生順に流す

## 根拠

警告と失敗の詳細は、その config の結果を読むための情報である。同じ config の失敗行と並べれば、
どの config で何が起きたかを 1 か所で読める。config 名の前置は build の段で nix の出力に付ける
`[<name>] ` と同じ形にした。nix の出力は子プロセスが stderr へ直接書くので、溜めると長い build の
進み具合が見えなくなる。そのため発生順のままにする。

## 棄却した代替案

### config 名を `layat:` の後ろに置く（`layat: warning: [<name>] <文>`）

build の段の `[<name>] layat: + nix …` と形が揃わず、行の先頭で config を見分けられない。

### 警告を発生順のまま流し、config 名だけを付ける

どの config の行かは分かるが、config ごとにまとまらない。並び順も実行ごとに変わる。
