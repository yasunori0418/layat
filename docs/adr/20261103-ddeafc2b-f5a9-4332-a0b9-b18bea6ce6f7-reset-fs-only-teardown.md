---
id: "ADR-ddeafc2b-f5a9-4332-a0b9-b18bea6ce6f7"
type: adr
name: "reset で profile を触らずに配置物を撤去する"
status: stable
issues:
  - "#584"
origin: "ADR-0020 §2（layat reset で配置物を無い状態へ戻す）を書き直した"
justifies:
  - "REQ-31f2882e-d2e3-4e3b-b783-feb627d73ac6"
references:
  - "ADR-c4ad9987-850e-4834-bbea-d8de3b073548"
  - "ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c"
  - "ADR-acc31876-df36-485e-b356-92ae60e26db7"
  - "ADR-a51897c5-f709-4b7a-8db2-61e6ebb0040c"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-ddeafc2b-f5a9-4332-a0b9-b18bea6ce6f7: reset で profile を触らずに配置物を撤去する

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-c4ad9987-850e-4834-bbea-d8de3b073548, ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c, ADR-acc31876-df36-485e-b356-92ae60e26db7, ADR-a51897c5-f709-4b7a-8db2-61e6ebb0040c
- 起点: ADR-0020 §2（layat reset で配置物を無い状態へ戻す）を書き直した

## 背景

配置した物を明示的に無い状態へ戻す手段が無かった。古くなった配置の除去は config から消えた entry の
symlink を保守的に消すが（ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c）、copy は世代の外に置くので
自動では決して消さない（ADR-c4ad9987-850e-4834-bbea-d8de3b073548）。copy を消す手段はどこにも
無かった。

profile の世代は manifest 全体を単位にしているので、一部の entry だけを世代から抜くことはできない。

## 決定

サブコマンド `layat reset <name> [target...]` を設け、config の全 entry、または指定した entry の
配置物を撤去する。symlink は古くなった配置の除去と同じ保守的な条件で消し、copy target は削除する。
profile と世代には触れない。config に entry が残っていれば、次の apply で配置し直される。

撤去の対象と条件は REQ-31f2882e-d2e3-4e3b-b783-feb627d73ac6 に置く。削除前の確認は
ADR-acc31876-df36-485e-b356-92ae60e26db7 が定める。

## 根拠

reset を FS の撤去だけに限れば、entry 単位の指定と矛盾しない。世代は変わらないので、apply を
やり直せばいつでも元へ戻せる。project mode では、次に devShell へ入ったときの確認で自然に復帰する。

copy は reset が消す唯一の手段になる。利用者が編集した内容や、もともとあったファイルを消すおそれは
確認で守る。symlink は従来の保守的な条件のままにし、記録に無い物を壊さない。

恒久的に外したいときは config から entry を消して apply すればよく、profile そのものの削除は
`nix-env --delete-generations` に任せられる。

## 棄却した代替案

### `apply --reset` のようにフラグにする

「配置しない apply」は意味が矛盾する。entry 単位の指定もフラグでは表しにくい。

### copy は残す

データは安全だが、copy を無い状態に戻したいという要望を満たせない。

### profile と世代も消す

reset がアンインストールを兼ねることになり、entry 単位の撤去と両立しない。
