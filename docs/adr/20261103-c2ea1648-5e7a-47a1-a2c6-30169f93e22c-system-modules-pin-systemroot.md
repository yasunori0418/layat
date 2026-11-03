---
id: "ADR-c2ea1648-5e7a-47a1-a2c6-30169f93e22c"
type: adr
name: "NixOS / nix-darwin モジュールは root を systemRoot に固定する"
status: 採用
issues:
  - "#519"
  - "#493"
origin: "Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）。ADR-0066 §1 を書き直した"
justifies:
  - "REQ-8d965ca2-f8fd-44a4-87f3-94e850e9f85b"
  - "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
references:
  - "ADR-0003"
  - "ADR-0064"
  - "ADR-0066"
  - "ADR-f85f1afd-9b87-4f7a-922e-97c62554f198"
---
# ADR-c2ea1648-5e7a-47a1-a2c6-30169f93e22c: NixOS / nix-darwin モジュールは root を systemRoot に固定する

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-0003, ADR-0064, ADR-0066, ADR-f85f1afd-9b87-4f7a-922e-97c62554f198, GitHub Issue #519, #493
- 起点: Issue #493（epic: NixOS / nix-darwin モジュールの実装）の grilling（2026-10-24）。ADR-0066 §1 を書き直した

## 背景

モジュールは engine を起動する配線であり、root と activation の時機だけを供給する。NixOS / nix-darwin
モジュールの root を何にするかについて、当初は `layat.user` で配置先のユーザーを受け取り、その home を
root にする想定だった。一方、ユーザーの home への配置は home-manager モジュールが既に担っている。
home-manager モジュールは NixOS / nix-darwin の中でも使える。

## 決定

NixOS / nix-darwin モジュールは、どの config の root も `systemRoot` に固定する。config の `subdir` が
null でなければ `systemRoot { subdir = <subdir>; }` とする。配置先のユーザーを受け取る `layat.user`
オプションは設けない。

root の供給元とオプションの規範は REQ-8d965ca2-f8fd-44a4-87f3-94e850e9f85b と
REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10 に置く。

## 根拠

host の activation は root 権限で実行される。ここからユーザーの home へ配置すると、ファイルの所有者の扱い・
ユーザーの解決・profile 基底の選び方をモジュールが抱えることになる。home への配置は home-manager
モジュールが担っており、同じ home へ配置する経路を 2 本にしない。

`subdir` の扱いを home-manager モジュールの `homeRoot` と同じ形にそろえるので、利用者は統合層を
変えても config の書き方を変えずに済む。

## 棄却した代替案

### `layat.user` を受け取り、そのユーザーの home へ home mode で配置する

当初の想定。home-manager モジュールと重なる経路を root 権限で持つことになる。
