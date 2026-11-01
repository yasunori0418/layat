---
id: "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
type: adr
name: "1 config = 1 profile を全モードで自前に持つ"
status: 採用
origin: "ADR-0002 §決定（全モードで自前 profile を持つ・粒度 = 1 profile）を書き直した"
justifies:
  - "REQ-1be4d678-959c-44d7-a346-44bfd95af56e"
  - "REQ-844ee375-919f-4341-81e1-a5f89fd32840"
references:
  - "ADR-0007"
---
# ADR-fe035038-4415-4a1e-aa5c-509fdfce9955: 1 config = 1 profile を全モードで自前に持つ

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-0007
- 起点: ADR-0002 §決定（全モードで自前 profile を持つ・粒度 = 1 profile）を書き直した

## 背景

stale 除去には「前回何を置いたか」の記録が要る。module 経由（home-manager / NixOS / nix-darwin）で
動くときは host が自前の世代を持つので、host の前世代（oldGenPath）から前回の配置を読む道もある。
あわせて、profile をどの粒度で分けるかを決める必要がある。

## 決定

layat は standalone と module のどちらでも、自前の profile を持つ。profile は config（entrypoint の
`layat.<name>`）1 つにつき 1 つで、適用は config 単位で atomic にする。host から layat へ求めるのは
switch のときに layat を起動することだけで、host の oldGenPath には依存しない。役割ごとに独立した
世代が欲しいときは config を分ける。

profile と前世代の読み方の細目は REQ-1be4d678-959c-44d7-a346-44bfd95af56e、module での host との
分担は REQ-844ee375-919f-4341-81e1-a5f89fd32840 に置く。

## 根拠

前世代の出所が全モードで「自 profile の前世代」にそろい、host ごとに違う oldGenPath の配管を
持たずに済む。他のモジュールの内部事情に依存しないという方針とも合う。home-manager が NixOS の
submodule として動くときにも自前の profile を持つのと同じ構図で、実績がある。

粒度を config にすると、世代が atomic に切り替わる単位と、独立して rollback できる単位が一致する。

## 棄却した代替案

### module では自前の profile を持たず host の世代に委ねる

前世代の出所が host ごとの oldGenPath に依存し、対応する host ごとに配管が増える。

### module では stale 除去をしない

前世代を持たないので dangling link が残り、standalone と振る舞いが食い違う。
