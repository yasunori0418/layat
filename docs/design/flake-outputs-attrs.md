---
id: "DSG-16373ec2-3496-4b12-b3b1-ef74e0435b58"
type: design
name: "layat 自身の flake outputs は packages / templates / 各モジュール / flakeModules / lib の 5 系統で構成する"
status: stable
satisfies:
  - "REQ-14f0aec9-abae-4621-82f3-40536a1ad904"
  - "REQ-6be1cbf1-6c6e-498b-8acb-7f4b80037169"
  - "REQ-c50df875-2cb0-4e72-8a21-858359a11cae"
  - "REQ-637599dc-a1ec-4af5-9e97-e882c7df56d0"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:40:55+09:00"
---
# layat 自身の flake outputs は packages / templates / 各モジュール / flakeModules / lib の 5 系統で構成する

## 設計

```nix
outputs = { ... }: {
  # layat CLI（一次 UX・配置エンジンを import）
  packages.<system>.layat = ...;   # buildGoModule（cmd/layat + internal）

  # 環境セットアップ（layat init = nix flake init -t のラッパー）
  templates.standalone = { path = ./templates/standalone; description = "..."; };
  templates.project    = { path = ./templates/project;    description = "..."; };

  # モジュール統合
  homeManagerModules.default = ./modules/home-manager.nix;
  nixosModules.default  = ./modules/nixos.nix;
  darwinModules.default = ./modules/nix-darwin.nix;

  # flake-parts module（perSystem.layat を flake.layat.<system> へ transpose）
  flakeModules.default = ./modules/flake-parts.nix;

  # 関数呼び出し（モジュールシステム不使用）
  lib = import ./lib;
};
```

5 系統は、layat が提供するものの**使い方の違い**で分けている。要求は「何を提供するか」を
定めるが、それを flake のどの attr 名で公開するかは設計側の判断になる。

- **`packages.<system>.layat`**: CLI の配布口（[layat CLI は PATH 常駐の一次 UX で、project mode は devShell 同梱を canonical とする](../requirements/cli-primary-ux-installation.md)）。
  `buildGoModule` で `cmd/layat` + `internal` をまとめてビルドする形にすることで、
  [CLI が持ち込む依存は許可した第三者ライブラリと pin した Go に限り、いずれも固定する](../requirements/cli-dependency-policy.md) が求める vendorHash による依存の固定が output の作り方として表れる
  （Go を nixpkgs の go に pin することと `toolchain` ディレクティブ不使用は
  同じ要求の規範で、output の形では表れない）。project mode で devShell へ同梱する
  ときもこの attr を参照する
- **`templates.<name>`**: [layat init は nix flake init -t への透明なラッパーとしファイルを生成しない](../requirements/init-template-wrapper.md) という要求があるため、
  テンプレート実体は nix 標準の `templates` output に
  置く必要がある。`standalone` / `project` の 2 本を置くことも同じ要求が定める
  （各テンプレートの中身は [template は動く example を 1 config だけ置きバリエーションはコメントで示す](../requirements/init-template-contents.md) の担当で、本 item は attr 名とパスの公開のみ）
- **`homeManagerModules` / `nixosModules` / `darwinModules`**: 各モジュールシステムの
  慣例 attr 名に合わせる。中身は `modules/` の各ファイルを指すだけ
  （NixOS / nix-darwin を system mode の配線として実装する設計は [NixOS / nix-darwin モジュールは modules/kick.nix を介した system mode の配線とする](module-stubs.md)）
- **`flakeModules.default`**: [flake-parts での使い方は直書きと同一の derivation を生み CLI のアドレッシングを変えない](../requirements/flake-parts-module-path.md) が求める flake-parts の使い方の提供口。
  consumer が import すると `perSystem.layat.<name>` を書けるようになる
- **`lib`**: モジュールシステムを介さない関数呼び出しの使い方。`import ./lib` を
  そのまま公開し、公開 API 面は [lib は公開 API・型定義・manifest 生成・マーカー構築子・apply コマンド組み立ての 5 ファイルへ分割する](lib-module-split.md) の `lib/default.nix` が決める

**ユーザー側の entrypoint が公開する `layat.<system>.<name>` は、この 5 系統とは別**で、
layat 自身の outputs ではなく consumer 側の output である（規範は [entrypoint は `layat.<name>` に named manifest を公開し CLI は形ごとの attr path で build する](../requirements/named-manifest-addressing.md)）。

## 出典

`docs/design.md`「flake.nix outputs 設計」冒頭のコードブロック。
