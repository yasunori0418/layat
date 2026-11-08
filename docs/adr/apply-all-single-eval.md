---
id: "ADR-28605860-770e-41df-95dd-7dc05b20497a"
type: adr
name: "apply --all は config ごとの情報を 1 回の eval でまとめて取る"
status: stable
sources:
  - id: "issue-550"
    resource: "https://github.com/yasunori0418/layat/issues/550"
    title: "#550"
origin: "実装前残セマンティクス確定の第6巡（2026-06-14）で確定。ADR-0024 §8 を書き直した"
justifies:
  - "REQ-535b811d-dfc5-4eac-92db-737e70eb5415"
references:
  - "ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9"
  - "ADR-955c2601-88fb-4be8-85de-ec7d0a8c727d"
  - "ADR-9e328362-2e90-441a-8eca-40081ce71b1b"
  - "ADR-feb60906-1517-46b8-8fa5-e007df4d3e11"
  - "ADR-f607b605-c69e-4c7e-ac89-75b7bce5bee1"
  - "ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988"
  - "ADR-777fc3c8-4ca8-49da-aaf3-146b7d991ebe"
  - "ADR-b3f3b68e-5096-46d1-bafd-53a9ab4711f4"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:00:56+09:00"
---
# ADR-28605860-770e-41df-95dd-7dc05b20497a: apply --all は config ごとの情報を 1 回の eval でまとめて取る

- ステータス: stable
- 日付: 2026-11-01
- 関連: [eval を先に行い、flock を取ってから build する](eval-first-then-flock-then-build.md), [shell.nix / default.nix の再現性は利用者の責任にする](shell-default-nix-best-effort.md), [エラーはエンベロープに構造化し stderr のテキストと併存させる](structured-errors-with-stderr.md), [--json は常に results\[\] エンベロープを出す](results-envelope-always.md), [終了コードは 0 / 1 / 2 の 3 値にする](exit-code-scheme.md), [profile は解決済み root の roothash でキーする](profile-keyed-by-roothash-of-resolved-root.md), [配置 target を gitignore コマンドで stdout へ列挙する](gitignore-command.md), [配置の振動はユーザー責任とし、文書で注意するに留める](oscillation-is-user-responsibility.md), GitHub Issue #550
- 起点: 実装前残セマンティクス確定の第6巡（2026-06-14）で確定。ADR-0024 §8 を書き直した

## 背景

apply は build の前に eval で rootKind を先に読み、root と profileDir を決める（[Roots の先読みを全ての生成器の前提条件にする](roots-lookahead-required-of-every-generator.md)）。
`apply --all` で config ごとに eval すると、config の数だけ nix の評価プロセスを起動することになる。

## 決定

`apply --all` は、全 config の rootKind などを 1 回の eval でまとめて取り、root モードのフィルタと
profileDir の確定にその結果を使う。build だけは、config ごとの atomic 性を保つために config ごとに
行う。

一括 eval で取る項目と式の形は [apply --all は rootKind を 1 回の一括 eval で取る](../requirements/apply-all-batched-eval.md) に置く。

## 根拠

eval のプロセス起動は config の数に比例して重くなる。1 回にまとめれば、起動のコストは config の
数に依らず一定になる。build を分けるのは、1 つの config の build 失敗を他の config に波及させない
ためである。

同じ一括 eval で config ごとの target 一覧も取れるので、config 間の target 衝突を適用前に止める
処理（[apply --all は選択した config 間の target 重複を build 前に止める](apply-all-preflight-target-conflict.md)）も、eval の回数を増やさずに行える。

## 棄却した代替案

### config ごとに eval する

実装は単一の apply と同じ処理で済むが、eval の起動コストが config の数だけ掛かる。
