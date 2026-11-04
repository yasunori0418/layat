---
id: "ADR-c2131901-290d-4311-8f84-64edb33e8897"
type: adr
name: "status は config 名の完全一致で絞り込む"
status: 採用
issues:
  - "#580"
  - "#198"
origin: "Issue #198（`layat status [name]`）の実装前に、2026-07-19 の grilling 以後の実装の変化を差分 grilling で反映した（2026-10-23）。ADR-0065 §2 を書き直した"
justifies:
  - "REQ-e8d911ff-8ee2-4d68-87d0-26f6d6817173"
---
# ADR-c2131901-290d-4311-8f84-64edb33e8897: status は config 名の完全一致で絞り込む

- ステータス: 採用
- 日付: 2026-11-03
- 関連: GitHub Issue #580, #198
- 起点: Issue #198（`layat status [name]`）の実装前に、2026-07-19 の grilling 以後の実装の変化を差分 grilling で反映した（2026-10-23）。ADR-0065 §2 を書き直した

## 背景

同じ config 名が home と system に、また複数の root に現れることがある。status
の出力で各記録を見分ける名前と、`[name]` 引数で何を絞り込むかを決める必要がある。

## 決定

status は、各記録を mode・config 名・root を含む名前で示す。`[name]` は config
名の部分の完全一致で絞り込み、mode と root は問わない。一致が 0 件なら一覧を出さない。stderr
に通知して exit 0 で終える。名前の書式と、backref や記録が読めない系列の名付けは
REQ-e8d911ff-8ee2-4d68-87d0-26f6d6817173 に置く。

## 根拠

config 名は利用者が自分で書いた名前で、最初に思い浮かぶ手掛かりになる。列挙の件数は少ない。
名前の完全一致で絞れば足りる。

名前を指定して一致しないのは「その名前の記録が無い」という結果で、コマンドの失敗ではない。

## 棄却した代替案

### mode や root で絞るフラグを持つ

列挙の件数は少なく、config 名の完全一致で足りる。

### 一致が 0 件のときに exit 1 で終える

記録が無いという正常な結果を失敗として扱うことになる。
