---
id: "ADR-7b9e4a7b-56d3-4356-9f67-bedbde8f0a52"
type: adr
name: "モジュールの起動修飾オプションを config ごとに持たない"
section: modules
status: stable
sources:
  - id: "issue-585"
    resource: "https://github.com/yasunori0418/layat/issues/585"
    title: "#585"
origin: "ADR-0061 §3 を書き直した"
justifies:
  - "REQ-d37d8433-a0c0-4387-aee9-9526f5cef70c"
  - "REQ-e1e1114b-ba07-4d57-8e04-6e30e39a5da3"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:30+09:00"
---
# モジュールの起動修飾オプションを config ごとに持たない

- ステータス: stable
- 日付: 2026-11-03
- 関連: GitHub Issue #585, `docs/spec.md`
- 起点: ADR-0061 §3 を書き直した

## 背景

`layat.backup` / `layat.recopy` / `layat.jobs` は、activation の起動に付ける修飾フラグを決める
オプションである。これらを `configs.<name>` ごとに上書きできる形も考えられる。

## 決定

3 つとも global のオプションとし、`configs.<name>` ごとの上書きは持たない。1 回の起動に載る
全 config へ一律に適用される。

`--recopy` の配線は [layat.recopy は engine 起動の配線レイヤーのオプションで、true のときだけ activation の起動へ --recopy を付ける](../requirements/recopy-wiring-layer.md)、`--backup` の配線は
[layat.backup は engine 起動の配線レイヤーのオプションで manifest には影響しない](../requirements/backup-wiring-layer.md) に置く。

## 根拠

activation は 1 回の `apply --all` の起動で、修飾フラグは起動全体に掛かる。config ごとに変えるには、
CLI の側に config ごとの修飾を足す必要がある。その要求は出ていない。

## 棄却した代替案

### `configs.<name>.recopy` などで config ごとに切り替える

CLI に config ごとの修飾を足す拡張が要る。要求が無いままインターフェイスを広げることになる。

### config ごとの修飾のために activation を config ごとの起動に戻す

集合の扱いを CLI に寄せた 1 回起動をやめ、ループと集約を bash で持つことになる。
