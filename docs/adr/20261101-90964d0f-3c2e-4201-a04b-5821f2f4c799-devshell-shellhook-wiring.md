---
id: "ADR-90964d0f-3c2e-4201-a04b-5821f2f4c799"
type: adr
name: "devShell の shellHook で配線する"
status: 採用
origin: "ADR-0005 の決定「devShell shellHook を正式な配線レイヤーに加える」を書き直した"
justifies:
  - "REQ-a0bdf6db-6c0c-476c-916a-61ee4e4510d9"
references:
  - "ADR-0003"
  - "ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d"
---
# ADR-90964d0f-3c2e-4201-a04b-5821f2f4c799: devShell の shellHook で配線する

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-0003, ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d, `docs/concept.md`, `docs/spec.md`
- 起点: ADR-0005 の決定「devShell shellHook を正式な配線レイヤーに加える」を書き直した

## 背景

project mode の主な用途は、リポジトリに入ったら repo 内の決まったパスへ自動で配置されることで
ある。HM・NixOS・nix-darwin の各モジュールは engine を起動するだけの配線で、root と起動の
タイミングを供給する（ADR-0003）。project mode にも、同じ役割を担う起動のきっかけが要った。

## 決定

devShell の `shellHook` を、モジュールと同型の正式な配線に加える。`nix develop` や direnv
（`use flake`）でシェルに入った時点で engine を起動し、git toplevel を root として project mode
で配置する。devShell は配置のロジックを持たず、root と起動のタイミングだけを供給する。規範の
細目は REQ-a0bdf6db-6c0c-476c-916a-61ee4e4510d9 に置く。

## 根拠

シェルへの入室は、作業ツリーに入ったことを示すいちばん自然なきっかけである。`shellHook` は
起動のタイミングを与えるだけで配置のロジックを持たないので、モジュールを配線に徹させる方針と
同じ形に収まる。入室のたびに実行されるので、変更が無いときは世代を積まない世代スキップ
（ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d）と組で成り立つ。

## 棄却した代替案

### 利用者が毎回手で `layat apply` を実行する

配線は要らないが、clone した直後や定義を変えた後に配置を忘れやすい。repo に入ると揃うという
project mode の主な用途を満たせない。
