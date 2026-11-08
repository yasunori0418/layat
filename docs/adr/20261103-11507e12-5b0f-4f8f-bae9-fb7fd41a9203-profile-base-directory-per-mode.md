---
id: "ADR-11507e12-5b0f-4f8f-bae9-fb7fd41a9203"
type: adr
name: "profile の基底ディレクトリをモードで切り替える"
status: 採用
origin: "実装前の残セマンティクス確定の第 4 巡（2026-06-14）と次期マイルストーン計画の grilling（2026-07-04）。ADR-0022 §3a と ADR-0036 §3a を書き直した"
justifies:
  - "REQ-2aa3abbc-90b2-486e-92de-d785554bdeb3"
  - "REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b"
  - "REQ-0d1f557b-7ae7-41ea-b17b-cdb5c80935d4"
references:
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
  - "ADR-41d506b4-2c5c-4a17-9954-b0de37e97de5"
  - "ADR-9755ba76-4c02-4b4e-a8c0-031a24a3a5b9"
  - "ADR-62540f8f-a8ba-4441-97b0-36b50d7862d8"
  - "ADR-1cbacca1-1930-46ca-bf82-30310f82216e"
  - "ADR-4c8d459a-c371-4f0d-b009-16e2ca096e49"
  - "ADR-eb1aa25f-a0ca-4efe-bcd9-4cbe11368764"
  - "ADR-ffe1afcb-0b51-4b87-b698-0a6dea307bdd"
  - "ADR-8b8c9109-a479-4571-8b10-5a6079127441"
  - "ADR-1908099c-87f7-4df1-bda5-7731df7d9ed7"
  - "ADR-efa7685c-6fa4-45df-b7a3-8f3f58e4118c"
  - "ADR-764ffff1-6b5a-4ed7-89d0-5700eb3ee883"
  - "ADR-9723dbb3-6659-4a69-ac4f-411c98bfa818"
  - "ADR-419914ba-9896-4dd1-853b-fb514fcb9f38"
  - "ADR-04cc2a03-78dd-4bf3-9cc7-3b8ed7faf429"
  - "ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75"
  - "ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed"
  - "ADR-45380022-d959-4b22-aa05-5993ac1f3c87"
  - "ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:41+09:00"
---
# ADR-11507e12-5b0f-4f8f-bae9-fb7fd41a9203: profile の基底ディレクトリをモードで切り替える

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd, ADR-41d506b4-2c5c-4a17-9954-b0de37e97de5, ADR-9755ba76-4c02-4b4e-a8c0-031a24a3a5b9, ADR-62540f8f-a8ba-4441-97b0-36b50d7862d8, ADR-1cbacca1-1930-46ca-bf82-30310f82216e, ADR-4c8d459a-c371-4f0d-b009-16e2ca096e49, ADR-eb1aa25f-a0ca-4efe-bcd9-4cbe11368764, ADR-ffe1afcb-0b51-4b87-b698-0a6dea307bdd, ADR-8b8c9109-a479-4571-8b10-5a6079127441, ADR-1908099c-87f7-4df1-bda5-7731df7d9ed7, ADR-efa7685c-6fa4-45df-b7a3-8f3f58e4118c, ADR-764ffff1-6b5a-4ed7-89d0-5700eb3ee883, ADR-9723dbb3-6659-4a69-ac4f-411c98bfa818, ADR-419914ba-9896-4dd1-853b-fb514fcb9f38, ADR-04cc2a03-78dd-4bf3-9cc7-3b8ed7faf429, ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75, ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed, ADR-45380022-d959-4b22-aa05-5993ac1f3c87, ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988
- 起点: 実装前の残セマンティクス確定の第 4 巡（2026-06-14）と次期マイルストーン計画の grilling（2026-07-04）。ADR-0022 §3a と ADR-0036 §3a を書き直した

## 背景

layat は config ごとに profileDir を持ち、その中に profile リンクと世代を置く。profileDir を並べる基底の
ディレクトリをどこにするかは、モードによって事情が違う。home / project / fixed の config はユーザーの
状態で、system mode の config はシステムの状態である。system mode の apply は root で実行されるので、
ユーザーの state ディレクトリをそのまま使うと root の home の下に置かれる。

## 決定

profile の基底は、system mode の config では `/nix/var/nix/profiles/layat`、それ以外の config では
ユーザー state 基底 `<state>/nix/profiles/layat` とする。`<state>` は `$XDG_STATE_HOME` があればそれ、
無ければ `~/.local/state` とする。基底は世代バックエンドの契約が与える。

基底の値とその下の並びは REQ-2aa3abbc-90b2-486e-92de-d785554bdeb3 と
REQ-0d1f557b-7ae7-41ea-b17b-cdb5c80935d4、基底を返す契約の操作は REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b に置く。

## 根拠

ユーザー state 基底は nix 本体の profile の既定（`~/.local/state/nix/profiles`）に揃う。`$XDG_STATE_HOME` を
変えている利用者でも、nix の profile と layat の profile が別の場所に散らばらない。

system mode の基底は、NixOS / nix-darwin が system の世代に使う `/nix/var/nix/profiles` に揃う。
システムの状態が root の個人領域に置かれることがなく、NixOS / nix-darwin モジュールも同じ場所をそのまま使える。

## 棄却した代替案

### system mode でもユーザー state 基底を使う

実装の差分は無いが、root で実行すると `/root/.local/state` の下、つまり root の home にシステムの
状態が置かれる。後で場所を移すと、既存の世代が孤児になる。

### system mode の基底を `/var/lib/layat` に置く

FHS には沿うが、nix の profile の慣習から外れ、nix の周辺で探したときに見つけにくい。

### ユーザー state 基底を `~/.local/state` に決め打ちする

`$XDG_STATE_HOME` を変えている環境で、nix 本体の profile と layat の profile の置き場所がずれる。
