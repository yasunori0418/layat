---
id: "ADR-e630482a-bc11-4ba6-a9cc-1ef46f54104d"
type: adr
name: "prune は引数を取らない"
status: stable
sources:
  - id: "issue-580"
    resource: "https://github.com/yasunori0418/layat/issues/580"
    title: "#580"
origin: "次期マイルストーン計画の grilling（2026-07-04）。公開後の実運用で孤児系列が溜まり始めたことを消費側の要求と判断した。ADR-0034 §3 を書き直した"
justifies:
  - "REQ-d41b1d0a-c6d5-41cc-93f9-e5cc7f152da4"
references:
  - "ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-e630482a-bc11-4ba6-a9cc-1ef46f54104d: prune は引数を取らない

- ステータス: stable
- 日付: 2026-11-03
- 関連: [profile は解決済み root の roothash でキーする](profile-keyed-by-roothash-of-resolved-root.md), GitHub Issue #580
- 起点: 次期マイルストーン計画の grilling（2026-07-04）。公開後の実運用で孤児系列が溜まり始めたことを消費側の要求と判断した。ADR-0034 §3 を書き直した

## 背景

reset・rollback・list-generations は config を名指しし、entrypoint を発見して nix eval / build
を行う。prune の対象は名前ではなく「root が消えた系列」で、利用者はその config
の名前を覚えていないことが多い。

## 決定

`layat prune` は config 名を引数に取らず、走査基底の全体を見る。entrypoint の発見と nix eval /
build を行わず、profile の状態だけを見る。mode を絞るフラグも持たない。

## 根拠

孤児系列の entrypoint はクローンごと消えていることが多く、
名指しを求めると消したい系列を指定できない。eval / build をしないので、手元に eval できない entrypoint
があっても prune は動く。

対象は roothash 系列に構造的に限られるので、rollback や list-generations のように mode
で絞る必要が無い。

## 棄却した代替案

### config 名を引数に取り、その系列だけを判定する

孤児の config 名は利用者の手元に残っていないことが多い。

### rollback / list-generations と同じように mode を限定するフラグを持つ

対象が roothash 系列に限られるので、絞り込む軸が無い。
