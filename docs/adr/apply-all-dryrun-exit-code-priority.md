---
id: "ADR-3c3aa4a5-083a-43c1-a99f-92b9e97a68b9"
type: adr
name: "apply --all --dryrun の終了コードは error を conflict より優先する"
status: stable
sources:
  - id: "issue-550"
    resource: "https://github.com/yasunori0418/layat/issues/550"
    title: "#550"
origin: "実装前残セマンティクス確定の第6巡（2026-06-14）で確定。ADR-0024 §7 を書き直した"
justifies:
  - "REQ-b7bb09d6-74c4-44d6-905f-cb5e8383ea32"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-01T23:01:19+09:00"
---
# ADR-3c3aa4a5-083a-43c1-a99f-92b9e97a68b9: apply --all --dryrun の終了コードは error を conflict より優先する

- ステータス: stable
- 日付: 2026-11-01
- 関連: GitHub Issue #550
- 起点: 実装前残セマンティクス確定の第6巡（2026-06-14）で確定。ADR-0024 §7 を書き直した

## 背景

`--dryrun` は conflict を見つけると 2 で、エラーなら 1 で終える。`apply --all --dryrun` では
config ごとに結果が分かれるので、error の config と conflict の config が両方あるときに、全体の
終了コードをどちらにするかを決める必要がある。

## 決定

`apply --all --dryrun` の終了コードは、error を conflict より優先して決める。error が 1 つでも
あれば 1、error が無く conflict があれば 2、どちらも無ければ 0 にする。

優先順の規範は REQ-b7bb09d6-74c4-44d6-905f-cb5e8383ea32 に置く。

## 根拠

eval や engine のエラーは conflict より深刻で、CI では先に気付く必要がある。終了コードの数値の
大小で決めると conflict（2）が error（1）を覆い、CI がエラーを conflict として扱ってしまう。

## 棄却した代替案

### config ごとの終了コードの最大値を取る

実装は最も単純だが、conflict（2）が error（1）を隠し、深刻なエラーに CI で気付けなくなる。
