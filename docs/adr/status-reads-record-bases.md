---
id: "ADR-7abe9cc7-30cd-49c9-b10e-bb5e1790ff0d"
type: adr
name: "status は 2 つの記録基底を走査して表示する"
status: stable
sources:
  - id: "issue-580"
    resource: "https://github.com/yasunori0418/layat/issues/580"
    title: "#580"
  - id: "issue-198"
    resource: "https://github.com/yasunori0418/layat/issues/198"
    title: "#198"
origin: "Issue #198（`layat status [name]`）の実装前に、2026-07-19 の grilling 以後の実装の変化を差分 grilling で反映した（2026-10-23）。ADR-0065 §1a を書き直した"
justifies:
  - "REQ-e8d911ff-8ee2-4d68-87d0-26f6d6817173"
  - "DSG-85439cb5-a149-422e-8596-fc4e860bc3a5"
references:
  - "ADR-52423842-bacf-4773-bbca-bcb34ae71013"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T17:33:20+09:00"
---
# status は 2 つの記録基底を走査して表示する

- ステータス: stable
- 日付: 2026-11-03
- 関連: [last-applied 記録は manifest を内容ごと持つ](record-embeds-manifest.md), GitHub Issue #580, #198
- 起点: Issue #198（`layat status [name]`）の実装前に、2026-07-19 の grilling 以後の実装の変化を差分 grilling で反映した（2026-10-23）。ADR-0065 §1a を書き直した

## 背景

「今なにが apply されているか」を確かめる手段が無かった。`list-generations` は世代の一覧しか出さず、
`apply --dryrun` は nix eval / build が要るうえ、示すのは新しい manifest との差分である。

last-applied 記録は配置した manifest を埋め込み、`appliedAt` と `linkFarm` を持つので、store
を辿らずに配置の内容を読める。記録の基底は、user の `<state>/layat` と system の `/var/lib/layat`
の 2 つがある。

## 決定

`layat status [name]` は last-applied 記録をデータ源にし、user と system の 2
つの記録基底をこの順に走査して、記録を持つ config を列挙する。profile 基底は読まず、世代番号は
`list-generations` に任せる。並び順と、読めない基底の扱いは
[status は記録基底を走査して last-applied 記録を読み取り専用で表示する](../requirements/status-command.md) に、engine と CLI の分担は
[status を engine の記録基底走査と CLI の絞り込み・出力に分ける](../design/status-scans-record-bases.md) に置く。

## 根拠

記録を読むなら、eval / build も store も要らない。profile の current は commit
の失敗で記録とずれることがあり、比較の基準として信頼されているのは記録の側である。

system mode の config も配置を持つ。基底ごとに別のコマンドへ分けると、
全体を一度に見るという目的を満たせない。

## 棄却した代替案

### profile 基底の current 世代を読む（Issue #198 の当初案）

link-farm の store パスを辿る必要があり、記録とずれた profile を表示しうる。

### 基底ごとに別のコマンドへ分ける

home・project と system の配置を一度に見られない。
