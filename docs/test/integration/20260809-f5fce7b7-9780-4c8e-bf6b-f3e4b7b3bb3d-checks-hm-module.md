---
id: "CASE-f5fce7b7-9780-4c8e-bf6b-f3e4b7b3bb3d"
type: test_case
name: "checks.hm-module — activation の配線を評価アサートで検証する flake check"
target: "checks.hm-module"
covers:
  - "TC-5a72b00e-2c03-4703-acdc-33cffb6fa994"
  - "TC-17fb5cfd-a691-475e-84e8-aa1853711b16"
---
# CASE-f5fce7b7-9780-4c8e-bf6b-f3e4b7b3bb3d: checks.hm-module — activation の配線を評価アサートで検証する flake check

## 対象

`flake.nix` の `checks.hm-module`（対象のモジュール実体は `modules/home-manager.nix`）

## 検証内容

standalone な home-manager configuration を設定違いで複数評価し、生成された activation
スクリプトと、それが参照する文書に対してアサートする。activate はしない。

- activation が統合先のファイル配置機構へ翻訳せず、名前つき config ごとのビルド済み配置物を
  名前の辞書順に `<name>=<配置物>` の組で並べ、engine を `apply --all` の 1 回の起動で渡して
  いること
- 2 つの config の配置物が別の store パスになり、どちらの文書も実在して root の種別として
  home を固定し、それぞれの config で宣言した entry をその属性キーを配置先として持ち、
  他の config の entry を持たないこと
- 退避のオプションを有効にし、かつサフィックスを省略したとき、既定のサフィックスが同じ起動の
  末尾へフラグとして乗ること
- 並列度のオプションを省略したとき、既定値 0 が `--jobs 0` として同じ起動へ明示で乗り、
  値を与えたときはその値が `--jobs` へそのまま渡ること
- 再コピーのオプションを有効にしたときだけ `--recopy` が同じ起動の末尾へ乗り、既定では
  どの起動にも乗らないこと
- 旧来の `layat.entries` だけで書いた設定が、`configs.default.entries` に同じ entry を書いた
  設定と同一の起動（同一の配置物）を生み、rename の deprecation warning を出すこと
- config が 1 つも無いとき、activation が engine を起動しないこと

配置元は固定の store パスを持つ test double で与え、文書の内容をマシンによらず安定させる。
home-manager はこの check のためだけの入力として持ち込み、ライブラリ本体には依存を波及させ
ない。公開用のラッパーを経由せず、モジュール本体へ CLI を直接注入して評価する。
