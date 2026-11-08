---
id: "ADR-cfd13fb6-d11e-4f3b-b0c6-40ee21b6b212"
type: adr
name: "list-generations --all は home mode と system mode の全 config を対象にする"
status: stable
origin: "ADR-0018 §4b を書き直した"
justifies:
  - "REQ-05abce3e-9797-432b-b93f-37c55d09afde"
references:
  - "ADR-baf6346d-783b-4607-ab1c-c2c93c09e064"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-cfd13fb6-d11e-4f3b-b0c6-40ee21b6b212: list-generations --all は home mode と system mode の全 config を対象にする

- ステータス: stable
- 日付: 2026-11-03
- 関連: [--all は config ごとの SubjectResult を積む](all-as-multiple-subject-results.md)
- 起点: ADR-0018 §4b を書き直した

## 背景

`list-generations` は世代を利用者に公開する mode でだけ使える。project mode は世代を内部の機構に
留め、公開しない。`--all` を `list-generations` でも受けるか、受けるならどの config を対象にするかは
決まっていなかった。

## 決定

`list-generations` は `--all` を受け、home mode の全 config、続けて system mode の全 config の
世代を一覧する。

基底が無いときと読めないときの扱いは [rollback と list-generations は home / system mode 限定にする](../requirements/rollback-list-generations-home-only.md) に置く。

## 根拠

`list-generations` は読み取り専用で、一括で実行しても状態を変えない。全 config の世代をまとめて
見られれば、どの config がどこまで進んでいるかを一度で掴める。

世代を公開する mode の全 config を対象にすれば、`--all` の範囲と世代の公開範囲が一致する。
利用者は「名指しで一覧できる config は `--all` にも出る」とだけ覚えればよい。

## 棄却した代替案

### home mode の config だけを対象にする

system mode の世代は名指しでは一覧できるのに `--all` には出ず、公開範囲と食い違う。

### 名指しだけを受ける

config ごとに実行すれば同じ情報は得られるが、読み取り専用の操作で一括を禁じる理由が無い。
