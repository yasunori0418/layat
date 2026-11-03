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
  - "ADR-0022"
  - "ADR-0036"
  - "ADR-0057"
  - "ADR-419914ba-9896-4dd1-853b-fb514fcb9f38"
---
# ADR-11507e12-5b0f-4f8f-bae9-fb7fd41a9203: profile の基底ディレクトリをモードで切り替える

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-0022, ADR-0036, ADR-0057, ADR-419914ba-9896-4dd1-853b-fb514fcb9f38
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
システムの状態が root の個人領域に住むことがなく、NixOS / nix-darwin モジュールも同じ場所をそのまま使える。

## 棄却した代替案

### system mode でもユーザー state 基底を使う

実装の差分は無いが、root で実行すると `/root/.local/state` の下に置かれ、システムの状態が root の home に
住む。後で場所を移すと、既存の世代が孤児になる。

### system mode の基底を `/var/lib/layat` に置く

FHS には沿うが、nix の profile の慣習から外れ、nix の周辺で探したときに見つけにくい。

### ユーザー state 基底を `~/.local/state` に決め打ちする

`$XDG_STATE_HOME` を変えている環境で、nix 本体の profile と layat の profile の置き場所がずれる。
