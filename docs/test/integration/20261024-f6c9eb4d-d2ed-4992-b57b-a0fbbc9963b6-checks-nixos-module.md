---
id: "CASE-f6c9eb4d-d2ed-4992-b57b-a0fbbc9963b6"
type: test_case
name: "checks.nixos-module — NixOS の activation の配線を評価アサートで検証する flake check"
target: "checks.nixos-module"
covers:
  - "TC-bd436b15-d2c3-4847-9797-582e37ef2367"
---
# CASE-f6c9eb4d-d2ed-4992-b57b-a0fbbc9963b6: checks.nixos-module — NixOS の activation の配線を評価アサートで検証する flake check

## 対象

`flake.nix` の `checks.nixos-module`（fixture は `tests/modules/nixos-module.nix`、対象のモジュール
実体は `modules/nixos.nix`）。Linux 以外の system では空の派生に差し替わる

## 検証内容

`nixosSystem` を設定違いで複数評価し、`system.activationScripts.layat` と、それが参照する
文書に対してアサートする。toplevel は評価せず、activate もしない。

- activation が名前つき config ごとのビルド済み配置物を名前の辞書順に `<name>=<配置物>` の組で
  並べ、engine を `apply --all` の 1 回の起動で渡していること
- 2 つの config の配置物が別の store パスになり、どちらの文書も root の種別として system を
  固定し、自分の entry だけを持つこと
- 退避のオプションを有効にしてサフィックスを省略したとき、既定のサフィックスが同じ起動に
  乗ること。並列度は既定の 0 が `--jobs 0` として明示で乗り、値を与えればその値が乗ること。
  再コピーは有効にしたときだけ `--recopy` が乗ること
- スクリプトの依存が `etc` と `users` で、dry-activate に対応する印が立っていること
- スクリプトが `NIXOS_ACTION` が dry-activate のときだけ `--dryrun` を付ける分岐を持ち、
  PATH に nix の実行ファイルのディレクトリを前置していること
- config が 1 つも無いとき、activation が engine を起動しないこと
- subdir を与えた config の文書だけ target に subdir が前置され、root の種別は system のまま
  であること

配置元は固定の store パスを持つ test double で与える。公開用のラッパーを経由せず、
モジュール本体へ CLI を直接注入して評価する。
