# integration

## テスト条件（`test_condition`）

* [manifest の読み取りが v1 を正確に復元し版の境界・未知フィールド・不整合な root 種別を拒否する](./20260809-172548ea-154a-4e22-a169-8252a43e3781-manifest-decode-schema-guard.md) - 契約の消費側が、渡された文書をどう読みどう拒むかを確かめる。
* [非 NixOS で HM configuration を実際に activate すると engine が配置する](./20260809-2f51dc14-f938-439d-b0c8-1d535914b2d3-hm-real-activation.md) - 評価アサートの層（TC-5a72b00e-2c03-4703-acdc-33cffb6fa994）で確かめられない部分、すなわち activate を実際に実行した ときに engine が起動して配置が行われることを確かめる。隔離した一時 HOME の下で、 activation package を実際にビルドして実行する。
* [activation が評価時点で engine 起動へ配線され root と entry とオプションを届ける](./20260809-5a72b00e-2c03-4703-acdc-33cffb6fa994-activation-engine-kick-wiring.md) - 統合層のモジュールを評価し、生成された activation の内容に対してアサートする。実際に activate はしない。この層が独立して必要なのは、ここで確かめる性質が**評価時に決まるのに activate 後にしか観測できない**ためで、評価だけで先に捕まえられる。
* [展開したテンプレが局所参照へ差し替えた上でそのまま評価を通る](./20260809-7e1e6a63-80d5-452c-bc49-81854ffac5a4-template-scaffold-evaluability.md) - はじめて使う人が最初に得る成果物が、手を入れずに評価を通ることを確かめる。
* [legacy entrypoint が canonical 形で同じ配置と出力契約を保ち素の nix-shell を壊さない](./20260809-d9c78439-e478-4779-977c-9c4c02dd4e93-legacy-entrypoint-parity.md) - flake を持たない起点からでも layat が同じ意味で使えることを確かめる。評価を再現可能にする ため、参照する nixpkgs を lock の内容へ固定してから実行する。
* [project mode の root が git toplevel へ一意に解決し外では曖昧に成功しない](./20260809-ec0bded4-c4bc-4a61-a2b5-0b652134b223-project-root-resolution.md) - 配置先の起点が呼び出した場所に依存しないことを確かめる。
* [project mode の世代は公開されず、拒否のときは理由が利用者へ届く](./20260811-f304d682-bc5e-4369-a427-f766fdfdd961-project-mode-generation-nondisclosure.md) - 世代をユーザーへ見せるコマンド（`rollback` / `list-generations`）が、project mode の config の世代を公開しないことを確かめる。engine 層が観測値の nil を正しく運ぶかどうかとは 別に、公開そのものを CLI が止めているかを見る。公開を止める処理は 2 つあり（名指しの拒否と 全件列挙からの除外）、この条件は両方を持つ。
* [root 相対 symlink が配置 root に対して相対形で張られ、worktree 移動後も有効で、前世代との比較が崩れない](./20261012-f1da084f-c0f0-4aba-85d9-f273c04eb373-root-relative-resolution.md) - root 相対 src の解決が配置 root と `target` の位置だけで決まることを確かめる。
* [HM の複数 config が独立した配置物と profile を持ち、衝突と部分失敗が config 単位で扱われる](./20261015-17fb5cfd-a691-475e-84e8-aa1853711b16-hm-multi-config-activation.md) - home-manager モジュールに名前つき config を複数与え、評価と実 activate の 2 層で確かめる。
* [NixOS の実 activate で system mode の配置・世代・記録が host の構成に収束する](./20261024-1b492e41-ee81-4945-a42b-0669dd574809-nixos-real-activation.md) - NixOS モジュールを組み込んだ host を VM で起動し、構成の切り替えを実際に activate して、 その後のファイルシステムに対してアサートする。評価時の配線だけでは観測できない、 activate 後の結果を確かめる。
* [NixOS / nix-darwin の activation が評価時点で system root を pin した engine 起動へ配線される](./20261024-bd436b15-d2c3-4847-9797-582e37ef2367-system-module-activation-wiring.md) - NixOS / nix-darwin のモジュールを host の構成として評価し、生成された activation の内容に 対してアサートする。実際に activate はしない。ここで確かめる性質は評価時に決まるが、 activate 後にしか観測できないため、評価だけで先に捕まえる。
* [事前展開した apply コマンドが全 config を渡し、root の選別を CLI のフィルタだけに委ねる](./20261025-4d477235-5917-469b-926b-119266048643-prebuilt-apply-command.md) - 評価時の文字列と、それを実行した結果の配置の両方を確かめる。

## テストケース（`test_case`）

* [e2e 01-project — project mode の配置と --all の部分失敗を実 nix で通す](./20260809-1fc8b8b9-d8c6-45f4-96a6-862de693822e-e2e-01-project.md) - `tests/e2e/scenarios/01-project.sh`（ハーネスは `tests/e2e/run.sh` / `tests/e2e/lib.sh`）
* [e2e 05-hm — 非 NixOS で HM configuration を実際に activate する](./20260809-3d858338-ceba-420b-87b7-995acff8a0a1-e2e-05-hm.md) - `tests/e2e/scenarios/05-hm.sh`
* [manifest_test.go — v1 の復元と版の境界・未知フィールド・root 種別の拒否・既知値の確認](./20260809-4179dcb2-f199-4469-ae87-0927ece06e65-manifest-test.md) - `internal/manifest/manifest_test.go`（対象は `internal/manifest/manifest.go` の読み取りと検証）
* [e2e 06-init-templates — テンプレを展開して展開物の評価が通ることを確かめる](./20260809-474dc938-dec8-4aae-9d91-607cfcd6dec0-e2e-06-init-templates.md) - `tests/e2e/scenarios/06-init-templates.sh`
* [e2e 07-legacy — 非 flake entrypoint から同じ配置と出力契約を通す](./20260809-8dd7bf87-687b-4ff9-bea2-0c4b3088c1c2-e2e-07-legacy.md) - `tests/e2e/scenarios/07-legacy.sh`
* [gitutil_test.go — toplevel 解決の一意性とリポジトリ外・git 不在の失敗](./20260809-aa967f9a-8020-44f8-b45f-57244711e1b5-gitutil-test.md) - `internal/gitutil/gitutil_test.go`（対象は `internal/gitutil/gitutil.go` の toplevel 解決）
* [checks.hm-module — activation の配線を評価アサートで検証する flake check](./20260809-f5fce7b7-9780-4c8e-bf6b-f3e4b7b3bb3d-checks-hm-module.md) - `flake.nix` の `checks.hm-module`（対象のモジュール実体は `modules/home-manager.nix`）
* [e2e 09-apply-all — apply --all の辞書順集約・並列度非依存・前段衝突 error・--no-wait 併用](./20260926-79561175-24f5-40a5-a031-509d7f90614b-e2e-09-apply-all.md) - `tests/e2e/scenarios/09-apply-all.sh`
* [e2e 10-generator-select — --manifest を使う場面の環境非依存と入力不正の E_INPUT](./20261002-e85f46cc-5876-4646-a7b2-15a7c71a22b8-e2e-10-generator-select.md) - `tests/e2e/scenarios/10-generator-select.sh`
* [e2e 11-root-relative — root 相対 symlink の相対形・移動耐性・古くなった配置の除去を実 nix で通す](./20261012-96b537f7-0bc9-4669-bead-743d0fc30e00-e2e-11-root-relative.md) - `tests/e2e/scenarios/11-root-relative.sh`（ハーネスは `tests/e2e/run.sh` / `tests/e2e/lib.sh`）
* [checks.darwin-module — nix-darwin の postActivation の配線を評価アサートで検証する flake check](./20261024-2def907b-3815-4801-8ea6-03662ee66aa3-checks-darwin-module.md) - `flake.nix` の `checks.darwin-module`（fixture は `tests/modules/darwin-module.nix`、対象のモジュール 実体は `modules/nix-darwin.nix`）。darwin 以外の system では空の派生に差し替わる
* [checks.nixos-vm — NixOS モジュールの実 activate を VM で検証する flake check](./20261024-68cfc77e-05c0-4b22-93fb-c7fdc5a91fac-checks-nixos-vm.md) - `flake.nix` の `checks.nixos-vm`（fixture は `tests/modules/nixos-vm.nix`、対象のモジュール 実体は `modules/nixos.nix`）。`pkgs.testers.runNixOSTest` で動き、KVM を要する。 x86_64-linux 以外の system では空の派生に差し替わる
* [checks.nixos-module — NixOS の activation の配線を評価アサートで検証する flake check](./20261024-f6c9eb4d-d2ed-4992-b57b-a0fbbc9963b6-checks-nixos-module.md) - `flake.nix` の `checks.nixos-module`（fixture は `tests/modules/nixos-module.nix`、対象のモジュール 実体は `modules/nixos.nix`）。Linux 以外の system では空の派生に差し替わる
* [e2e 14-flake-parts — mkApplyCommand を埋めた devShell が project config だけを配置する](./20261025-9cd12012-54c0-4c97-8502-9bfb4aa7363e-e2e-14-flake-parts.md) - `tests/e2e/scenarios/14-flake-parts.sh`（ハーネスは `tests/e2e/run.sh` / `tests/e2e/lib.sh`）
