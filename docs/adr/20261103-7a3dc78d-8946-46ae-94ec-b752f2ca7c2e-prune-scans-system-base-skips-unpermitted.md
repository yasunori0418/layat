---
id: "ADR-7a3dc78d-8946-46ae-94ec-b752f2ca7c2e"
type: adr
name: "prune は system 基底も走査し、権限の無い系列を見送る"
status: 採用
origin: "次期マイルストーン計画の grilling（2026-07-04）。ADR-0034 §1b と ADR-0036 §3c を書き直した"
justifies:
  - "REQ-c44433a1-7ee7-459a-9aae-7cc42166876f"
  - "DSG-096dc893-21f4-45e3-9347-986e9275b4d1"
supersedes:
  - "ADR-0034"
  - "ADR-0036"
references:
  - "ADR-0034"
  - "ADR-0036"
---
# ADR-7a3dc78d-8946-46ae-94ec-b752f2ca7c2e: prune は system 基底も走査し、権限の無い系列を見送る

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-0034, ADR-0036
- 起点: 次期マイルストーン計画の grilling（2026-07-04）。ADR-0034 §1b と ADR-0036 §3c を書き直した

## 背景

`layat prune` は、root が実在しなくなった `<roothash>` 系列を削除するコマンドである。profile の基底は
ユーザー state 基底だけではなく、system mode の config では system profile 基底になる。system mode でも
`--root` で root を上書きした系列は `<roothash>` でキーされるので、上書き先のディレクトリが消えれば
ユーザー state 基底の系列と同じように孤児になる。

一方、system 基底の下にある系列の削除には root 権限が要る。prune を非 root で実行すると、
system 基底の孤児系列は削除しようとした時点で権限エラーになる。

## 決定

`layat prune` はユーザー state 基底に加えて system profile 基底も走査する。非 root で実行して権限が
無いために削除へ着手できなかった系列は、warning を出して見送り、他の系列の処理を続ける。

走査の対象と見送りの条件は REQ-c44433a1-7ee7-459a-9aae-7cc42166876f と
DSG-096dc893-21f4-45e3-9347-986e9275b4d1 に置く。

## 根拠

system mode の `--root` 系列も、ユーザー state 基底の系列と同じ理由で孤児になる。基底が違うことを
理由に走査から外すと、その孤児を消す手段が手作業しか残らない。

system 基底の系列を 1 つ消せないことを理由に全体を止めると、消せるはずの系列まで残る。見送った系列は warning で
示されるので、root で実行し直せば消せることが分かる。

## 棄却した代替案

### system 基底を走査の対象から外す

root = `/` の `<name>` 直キー系列は孤児にならないが、`--root` で上書きした系列は孤児になる。
これを消す手段が無くなる。

### 権限エラーで prune 全体を失敗させる

非 root の実行では system 基底に孤児が 1 つあるだけで毎回失敗し、ユーザー state 基底の掃除もできなくなる。
