---
id: "ADR-764ffff1-6b5a-4ed7-89d0-5700eb3ee883"
type: adr
name: "state の基底を設定で上書きさせない"
status: 採用
issues:
  - "#406"
  - "#140"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）と Issue #140（epic #129: system mode の実装）の計画（2026-10-18）。ADR-0057 §6b・ADR-0062 §2 を書き直した"
justifies:
  - "REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b"
  - "REQ-0d1f557b-7ae7-41ea-b17b-cdb5c80935d4"
---
# ADR-764ffff1-6b5a-4ed7-89d0-5700eb3ee883: state の基底を設定で上書きさせない

- ステータス: 採用
- 日付: 2026-11-03
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）と Issue #140（epic #129: system mode の実装）の計画（2026-10-18）。ADR-0057 §6b・ADR-0062 §2 を書き直した

## 背景

layat の state には 2 種類の基底がある。世代の profile の基底（user は `<state>/nix/profiles/layat`、
system は `/nix/var/nix/profiles/layat`）と、last-applied 記録の基底（system mode は `/var/lib/layat`）。
基底を変えると、既存の世代と記録を見失う。

一方で、root 権限を user namespace で得る e2e は、実システムの基底を汚さずに配置と世代を検証したい。

## 決定

state の基底を設定ファイルのキーで上書きさせない。上書きの口は engine の入力と、CLI がそれへ渡す
テスト用の環境変数 `LAYAT_SYSTEM_PROFILE_BASE`・`LAYAT_SYSTEM_STATE_DIR` だけにする。環境変数は
サポート対象外とし、空文字は上書きと見なさない。

設定キーを設けない規範は REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b、system mode の基底は
REQ-0d1f557b-7ae7-41ea-b17b-cdb5c80935d4 に置く。

## 根拠

profile の基底は世代をどこに置くかで、ファイル配置の関心ではない。置き場所は世代バックエンドの責務で、
契約の UserBase・SystemBase がその所有を表す。設定キーを足すと、世代管理を layat から切り出すときに
layat 側へ互換の負債が残る。

記録の基底も、変えれば既存の記録を見失う点は同じなので、profile の基底に揃える。テストの隔離に
要るのは環境変数の口だけ。

## 棄却した代替案

### 設定ファイルのキーで基底を上書きできるようにする

利用者の需要が無く、上書きのたびに既存の世代と記録を見失うおそれがある。世代管理を切り出すときにも
互換の負債になる。

### テスト用の口も持たない

e2e が実システムの基底を書き換えずに system mode を検証できない。
