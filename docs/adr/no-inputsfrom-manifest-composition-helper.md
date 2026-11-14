---
id: "ADR-c973a6e3-b6d9-45d3-b237-83fb1b361b22"
type: adr
name: "複数の manifest を inputsFrom で合成するヘルパを持たない"
section: lib-entries
status: stable
sources:
  - id: "issue-112"
    resource: "https://github.com/yasunori0418/layat/issues/112"
    title: "#112"
  - id: "issue-107"
    resource: "https://github.com/yasunori0418/layat/issues/107"
    title: "#107"
origin: "ADR-0032 §3b を書き直した"
justifies:
  - "REQ-c890ce4a-6528-4ab3-ac86-23d7aebff7da"
references:
  - "ADR-28605860-770e-41df-95dd-7dc05b20497a"
  - "ADR-6b7becf0-f029-41e3-895e-b672dce94d3d"
  - "ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3"
  - "ADR-e683ac05-c0e0-41c1-8b6a-38b605972fe1"
  - "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# 複数の manifest を inputsFrom で合成するヘルパを持たない

- ステータス: stable
- 日付: 2026-11-03
- 関連: [apply --all は config ごとの情報を 1 回の eval でまとめて取る](apply-all-single-eval.md), [CLI は cobra で組む](cobra-cli-framework.md), [root の必須性を別の確認を書かずに標準のエラーで担保する](root-required.md), [named manifest を専用の `layat.<name>` 名前空間に公開する](layat-name-namespace.md), [home mode の CLI と lib の版ずれを manifest の確認に任せて許容する](home-mode-skew-guarded-by-schemaversion.md), [1 config = 1 profile を全モードで自前に持つ](one-config-one-profile-all-modes.md), GitHub Issue #112, #107
- 起点: ADR-0032 §3b を書き直した

## 背景

legacy entrypoint の正準形は `mkShell` の derivation である。`mkShell` には他の derivation の入力を
取り込む `inputsFrom` があり、これで複数の manifest を 1 つに合成するヘルパを layat が用意する案が
あった。

## 決定

複数の manifest を `inputsFrom` などで 1 つに合成するヘルパを持たない。役割ごとの分離は、
`layat.<name>` を複数持つ config の分割で行う。

legacy entrypoint で named manifest を公開する形は [legacy entrypoint は mkShell passthru 形を canonical とし CLI の attr path を分岐させない](../requirements/legacy-entrypoint-passthru-canonical.md) に置く。

## 根拠

世代の単位は 1 config に 1 profile である。複数の manifest を 1 つの profile に混ぜると、古くなった配置物が
どの manifest に由来するかが曖昧になり、rollback と古くなった配置の除去で何を戻し何を除くかが決まらなくなる。

複数の config をまとめて処理したいときは `apply --all` で足り、合成は要らない。

## 棄却した代替案

### `inputsFrom` で合成するヘルパを提供する

1 つの devShell に manifest をまとめて載せられるが、世代の単位を壊す。役割の分離を config の
分割で行う既存の方針とも食い違う。
