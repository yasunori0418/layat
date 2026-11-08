---
id: "ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75"
type: adr
name: "module と project では世代操作を公開しない"
status: stable
origin: "ADR-0002 §決定（module の profile は内部機構）・ADR-0005 §決定（project mode の世代はユーザー露出しない）を書き直した"
justifies:
  - "REQ-844ee375-919f-4341-81e1-a5f89fd32840"
  - "REQ-05abce3e-9797-432b-b93f-37c55d09afde"
  - "REQ-46fccb80-4bae-4d37-bc19-dded88e9a9c0"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:41+09:00"
---
# ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75: module と project では世代操作を公開しない

- ステータス: stable
- 日付: 2026-11-01
- 起点: ADR-0002 §決定（module の profile は内部機構）・ADR-0005 §決定（project mode の世代はユーザー露出しない）を書き直した

## 背景

layat は全モードで自前の profile を持つが、その世代を rollback や一覧としてユーザーに見せるべきかは
適用のしかたによって事情が違う。module 経由では host が自前の rollback を持つ。project mode の配置物は
リポジトリに残さない ephemeral なもので、devShell から起動されたときには戻し先になる host の世代も無い。

## 決定

`layat rollback` と `layat list-generations` を公開するのは、配置が永続する home mode と system mode
に限る。module 経由では layat の profile を内部機構に留め、ユーザー向けの rollback を host
（`home-manager --rollback` など）に一本化する。layat の profile は前進だけを行い、host の rollback が
旧 config で layat を再び起動することで追従する。project mode では profile を古くなった配置の除去と世代
スキップ判定のための内部機構に留める。

モードごとの公開範囲は REQ-05abce3e-9797-432b-b93f-37c55d09afde、module での host との分担は
REQ-844ee375-919f-4341-81e1-a5f89fd32840、project mode の世代の扱いは
REQ-46fccb80-4bae-4d37-bc19-dded88e9a9c0 に置く。

## 根拠

module で layat が独自に rollback できると、host の世代と layat の世代がずれる。host の rollback で
layat を再起動すれば FS は旧 config の状態へ収束するので、独自の rollback は要らない。
home-manager が NixOS の submodule として動くときと同じやり方になる。

project mode の配置物は ephemeral で、前の状態へ戻す意味が薄い。公開しなければ、利用者に見える
世代操作を home / system mode の永続的な配置だけに絞れる。

## 棄却した代替案

### module でも `layat rollback` を公開する

利用者が layat 側だけを戻すと host の世代とずれ、次の host の switch で意図しない状態へ戻る。

### project mode でも世代操作を公開する

配置物が ephemeral なので戻す意味が薄く、devShell からの起動では戻し先の host の世代も無い。
