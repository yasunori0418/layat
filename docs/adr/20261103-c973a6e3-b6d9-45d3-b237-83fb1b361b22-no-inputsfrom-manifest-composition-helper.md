---
id: "ADR-c973a6e3-b6d9-45d3-b237-83fb1b361b22"
type: adr
name: "複数の manifest を inputsFrom で合成するヘルパを持たない"
status: 採用
issues:
  - "#112"
  - "#107"
origin: "ADR-0032 §3b を書き直した"
justifies:
  - "REQ-c890ce4a-6528-4ab3-ac86-23d7aebff7da"
supersedes:
  - "ADR-0032"
---
# ADR-c973a6e3-b6d9-45d3-b237-83fb1b361b22: 複数の manifest を inputsFrom で合成するヘルパを持たない

- ステータス: 採用
- 日付: 2026-11-03
- 関連: GitHub Issue #112, #107
- 起点: ADR-0032 §3b を書き直した

## 背景

legacy entrypoint の正準形は `mkShell` の derivation である。`mkShell` には他の derivation の入力を
取り込む `inputsFrom` があり、これで複数の manifest を 1 つに合成するヘルパを layat が用意する案が
あった。

## 決定

複数の manifest を `inputsFrom` などで 1 つに合成するヘルパを持たない。役割ごとの分離は、
`layat.<name>` を複数持つ config の分割で行う。

legacy entrypoint で named manifest を公開する形は REQ-c890ce4a-6528-4ab3-ac86-23d7aebff7da に置く。

## 根拠

世代の単位は 1 config に 1 profile である。複数の manifest を 1 つの profile に混ぜると、古くなった配置物が
どの manifest に由来するかが曖昧になり、rollback と古くなった配置の除去で何を戻し何を除くかが決まらなくなる。

複数の config をまとめて処理したいときは `apply --all` で足り、合成は要らない。

## 棄却した代替案

### `inputsFrom` で合成するヘルパを提供する

1 つの devShell に manifest をまとめて載せられるが、世代の単位を壊す。役割の分離を config の
分割で行う既存の方針とも食い違う。
