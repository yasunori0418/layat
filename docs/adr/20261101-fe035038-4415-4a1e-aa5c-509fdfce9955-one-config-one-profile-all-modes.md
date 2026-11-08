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
  - "ADR-560852e1-3d07-46b9-ae1a-d15f82d28a5e"
  - "ADR-f85f1afd-9b87-4f7a-922e-97c62554f198"
  - "ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed"
  - "ADR-88231ffc-4836-4252-be34-894944fd28c3"
  - "ADR-86a77232-dc99-4c04-bd6d-2127410e8828"
  - "ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246"
  - "ADR-e683ac05-c0e0-41c1-8b6a-38b605972fe1"
  - "ADR-28605860-770e-41df-95dd-7dc05b20497a"
  - "ADR-955c2601-88fb-4be8-85de-ec7d0a8c727d"
  - "ADR-65b3439a-b90a-4056-b55a-522399bc71e6"
  - "ADR-d15a216f-b5d6-4df9-9825-b23a341bec1d"
  - "ADR-553eea60-6ab4-4da7-93a9-3acf44b2fae3"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# ADR-fe035038-4415-4a1e-aa5c-509fdfce9955: 1 config = 1 profile を全モードで自前に持つ

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-560852e1-3d07-46b9-ae1a-d15f82d28a5e, ADR-f85f1afd-9b87-4f7a-922e-97c62554f198, ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed, ADR-88231ffc-4836-4252-be34-894944fd28c3, ADR-86a77232-dc99-4c04-bd6d-2127410e8828, ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246, ADR-e683ac05-c0e0-41c1-8b6a-38b605972fe1, ADR-28605860-770e-41df-95dd-7dc05b20497a, ADR-955c2601-88fb-4be8-85de-ec7d0a8c727d, ADR-65b3439a-b90a-4056-b55a-522399bc71e6, ADR-d15a216f-b5d6-4df9-9825-b23a341bec1d, ADR-553eea60-6ab4-4da7-93a9-3acf44b2fae3
- 起点: ADR-0002 §決定（全モードで自前 profile を持つ・粒度 = 1 profile）を書き直した

## 背景

古くなった配置の除去には「前回何を置いたか」の記録が要る。last-applied 記録（ADR-df8ab6ff-7636-42ff-8b47-6ee1209dfd15）が無いときは、
前世代から読む。module 経由（home-manager / NixOS / nix-darwin）で動くときは host が自前の世代を
持つので、host の前世代（oldGenPath）から前回の配置を読む道もある。
あわせて、profile をどの粒度で分けるかを決める必要がある。

## 決定

layat は CLI 単体と module のどちらでも、自前の profile を持つ。profile は config（entrypoint の
`layat.<name>`）1 つにつき 1 つで、適用は config 単位で atomic にする。host から layat へ求めるのは
switch のときに layat を起動することだけで、host の oldGenPath には依存しない。役割ごとに独立した
世代が欲しいときは config を分ける。

profile と前世代の読み方の細目は REQ-1be4d678-959c-44d7-a346-44bfd95af56e、module での host との
分担は REQ-844ee375-919f-4341-81e1-a5f89fd32840 に置く。

## 根拠

last-applied 記録が無いときに読む前世代の出所が、全モードで「自 profile の前世代」にそろう。
host ごとに違う oldGenPath の配管を持たずに済む。他のモジュールの内部事情に依存しないという
方針とも合う。home-manager が NixOS の submodule として動くときにも自前の profile を持つのと
同じやり方で、実績がある。

粒度を config にすると、世代が atomic に切り替わる単位と、独立して rollback できる単位が一致する。

## 棄却した代替案

### module では自前の profile を持たず host の世代に委ねる

前世代の出所が host ごとの oldGenPath に依存し、対応する host ごとに配管が増える。

### module では古くなった配置の除去をしない

前世代を持たないので dangling link が残り、CLI 単体と振る舞いが食い違う。
