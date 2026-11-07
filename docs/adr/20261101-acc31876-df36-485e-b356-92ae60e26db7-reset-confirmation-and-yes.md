---
id: "ADR-acc31876-df36-485e-b356-92ae60e26db7"
type: adr
name: "reset は非対話の実行で --yes を必須にする"
status: stable
sources:
  - id: "issue-550"
    resource: "https://github.com/yasunori0418/layat/issues/550"
    title: "#550"
  - id: "issue-outturn-1"
    resource: "https://github.com/yasunori0418/outturn/issues/1"
    title: "yasunori0418/outturn#1"
origin: "実装前残セマンティクス確定の第7巡（2026-06-14）で確定。ADR-0025 §5 と ADR-0043 §8 の reset --json を書き直した"
justifies:
  - "REQ-31dae599-f3a3-4bbe-b367-c955535265da"
  - "REQ-2a613337-7646-4ced-8807-e43bca18acf3"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:00:56+09:00"
---
# ADR-acc31876-df36-485e-b356-92ae60e26db7: reset は非対話の実行で --yes を必須にする

- ステータス: stable
- 日付: 2026-11-01
- 関連: GitHub Issue #550, yasunori0418/outturn#1
- 起点: 実装前残セマンティクス確定の第7巡（2026-06-14）で確定。ADR-0025 §5 と ADR-0043 §8 の reset --json を書き直した

## 背景

`reset` は配置物を消す破壊的な操作で（ADR-ddeafc2b-f5a9-4332-a0b9-b18bea6ce6f7）、実行前に確認のプロンプトを出す。CI・direnv・
パイプのように stdin が TTY でない環境では、プロンプトが入力を待ってハングするか、空の入力を同意と
取り違えて削除してしまうおそれがある。`--json` の出力を機械が読む場合も、プロンプトには誰も答えない。

## 決定

`reset` は、確認のプロンプトに答えられない実行では `-y` / `--yes` を必須にする。stdin が TTY で
ないときと、`--json` を付けたときがこれに当たる。`--yes` が無ければ、プロンプトを出さずにエラーで
止める。

TTY でない場合の規範は REQ-31dae599-f3a3-4bbe-b367-c955535265da に、`--json` の場合の規範は
REQ-2a613337-7646-4ced-8807-e43bca18acf3 に置く。

## 根拠

非対話の環境で破壊的な操作を行うなら、同意は `--yes` として明示されているべきである。止めれば、
ハングと誤削除の両方を防げる。`--json` は TTY であっても機械が出力を読む前提なので、同じ扱いにする。

## 棄却した代替案

### TTY でないときは `--yes` を付けたものとみなす

スクリプトからは楽になるが、非対話の環境で意図しない削除が起きやすくなる。
