---
id: "ADR-54f09f52-1c2c-4b84-8f18-717fd93e4342"
type: adr
name: "status を読み取り専用にする"
status: stable
sources:
  - id: "issue-580"
    resource: "https://github.com/yasunori0418/layat/issues/580"
    title: "#580"
  - id: "issue-198"
    resource: "https://github.com/yasunori0418/layat/issues/198"
    title: "#198"
origin: "Issue #198（`layat status [name]`）の実装前に、2026-07-19 の grilling 以後の実装の変化を差分 grilling で反映した（2026-10-23）。ADR-0065 §4 を書き直した"
justifies:
  - "REQ-e8d911ff-8ee2-4d68-87d0-26f6d6817173"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T17:33:20+09:00"
---
# ADR-54f09f52-1c2c-4b84-8f18-717fd93e4342: status を読み取り専用にする

- ステータス: stable
- 日付: 2026-11-03
- 関連: GitHub Issue #580, #198
- 起点: Issue #198（`layat status [name]`）の実装前に、2026-07-19 の grilling 以後の実装の変化を差分 grilling で反映した（2026-10-23）。ADR-0065 §4 を書き直した

## 背景

apply や reset は flock を取り、profileDir や記録を作り、system mode では euid を確かめる。
記録を読むだけの status に、これらが要るかを決める必要がある。

## 決定

status は flock を取らず、ディレクトリもファイルも作らず、記録を書き換えない。euid も確かめない。

## 根拠

status は書き込まないので、apply と排他する必要が無い。記録は rename で atomic に書かれるので、lock
を取らなくても書きかけの記録を読むことは無い。

記録は mode 0644 で書かれるので、root 以外の利用者でも system 基底の記録を読める。euid
を確かめなければ、一般の利用者が system mode の配置を確かめられる。

## 棄却した代替案

### flock を取って読む

実行中の apply を待たされ、lock ファイルを作る副作用も出る。atomic な書き込みのおかげで、lock
が無くても書きかけの記録を読むことは無い。

### system 基底を読むときに root 権限を求める

記録は機密ではなく、配置を確かめるだけのために root 権限を求めるのは過剰である。
