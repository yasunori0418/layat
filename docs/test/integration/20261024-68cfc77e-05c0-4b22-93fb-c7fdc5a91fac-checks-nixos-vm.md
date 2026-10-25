---
id: "CASE-68cfc77e-05c0-4b22-93fb-c7fdc5a91fac"
type: test_case
name: "checks.nixos-vm — NixOS モジュールの実 activate を VM で検証する flake check"
target: "checks.nixos-vm"
covers:
  - "TC-1b492e41-ee81-4945-a42b-0669dd574809"
---
# CASE-68cfc77e-05c0-4b22-93fb-c7fdc5a91fac: checks.nixos-vm — NixOS モジュールの実 activate を VM で検証する flake check

## 対象

`flake.nix` の `checks.nixos-vm`（fixture は `tests/modules/nixos-vm.nix`、対象のモジュール
実体は `modules/nixos.nix`）。`pkgs.testers.runNixOSTest` で動き、KVM を要する。
x86_64-linux 以外の system では空の派生に差し替わる

## 検証内容

layat の entry を差し替えた 2 つの構成（a / b）を用意し、a の VM だけを起動する。b の toplevel は
VM の nix DB に登録して switch 先に使う。boot 完了後に次を順に確かめる。

- a の `switch-to-configuration switch` で `/etc` 配下の target が store の配置元を指す
  symlink になり、`/nix/var/nix/profiles/layat/<name>/profile-1-link` と
  `/var/lib/layat/<name>/last-applied.json` が立ち、世代が 1 件であること
- 同じ構成で `/run/current-system/activate` を再実行しても、配置が保たれ世代が 1 件のままであること
- b の `switch-to-configuration switch` で b の target が配置され、a の target が除去され、
  世代が 2 件になり、続けて再 activate しても 2 件のままであること
- a の `switch-to-configuration switch` を再実行すると a の target が戻り、b の target が除去されること

初回 boot 時の activation の結果はアサートしない。配置元は VM 内で中身を確かめられる
小さな store パスで与える。
