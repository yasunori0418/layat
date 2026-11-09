---
id: "ADR-997c2459-f2ab-4025-ac92-e9ce836fc9ed"
type: adr
name: "rollback に config の名指しを必須にする"
status: stable
origin: "ADR-0018 §4c を書き直した"
justifies:
  - "REQ-89c7baf9-9be0-417b-bd2d-b2e4edabe796"
references:
  - "ADR-baf6346d-783b-4607-ab1c-c2c93c09e064"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# rollback に config の名指しを必須にする

- ステータス: stable
- 日付: 2026-11-03
- 関連: [--all は config ごとの SubjectResult を積む](all-as-multiple-subject-results.md)
- 起点: ADR-0018 §4c を書き直した

## 背景

`--all` を `apply` 以外のサブコマンドにも広げるかを決める際、config の世代を 1 つ前へ戻して配置物を
書き換える `rollback` も候補になった。

## 決定

`rollback` は `--all` を受けない。戻す config の名指しを必須にする。

引数の形は [rollback は config の名指しを必須にする](../requirements/rollback-named-only.md) に置く。

## 根拠

全 config を一斉に前の世代へ戻すのは、配置物を書き換える一括操作である。戻したいのは通常、問題が
出た役割の config だけだ。一括は誤操作の被害を広げる。途中で失敗すると、config ごとに世代が揃わない
状態が残る。

名指しを必須にすれば、どの config を戻すかを利用者が毎回明示する。配置を利用者が明示的に握るという
layat の方針にも合う。

## 棄却した代替案

### `rollback --all` に対応する

全体を戻す手間は省けるが、上に挙げた誤操作と途中失敗の危険を既定の操作に持ち込む。

### `--all` を全サブコマンドで一律に受ける、または一律に受けない

一貫はする。しかしコマンドごとに安全性と実用性が違う。読み取りだけの `gitignore` や
`list-generations` と、配置を書き換える `rollback` を同じ扱いにする理由が無い。
