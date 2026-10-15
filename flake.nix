{
  description = "Lays contents at root-relative targets, as the manifest says.";

  # layat 自体を clone して nix develop / build / flake check する際に cachix からビルド済み
  # バイナリを引くための設定（trusted-user / accept-flake-config 前提）。flake の nixConfig は
  # input に伝播しないため、layat を flake input として消費する側のキャッシュ取得には効かない。
  nixConfig = {
    extra-substituters = [
      "https://cache.nixos.org/"
      "https://nix-community.cachix.org"
      "https://yasunori0418.cachix.org"
    ];
    extra-trusted-public-keys = [
      "cache.nixos.org-1:6NCHdD59X431o0gWypbMrAURkbJ16ZPMQFGspcDShjY="
      "nix-community.cachix.org-1:mB9FSh9qf2dCimDSUo8Zy7bkq5CX+/rkCWyvRCYg3Fs="
      "yasunori0418.cachix.org-1:mC1j+M5A6063OHaOB5bH2nS0BiCW/BJsSRiOWjLeV9o="
    ];
  };

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      # nix-unit の flake-parts モジュールは nixpkgs-lib follows を要求する。
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    # checks.hm-module 専用。lib/ は home-manager に依存しない。
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    # lib 評価テスト。
    nix-unit = {
      url = "github:nix-community/nix-unit";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    haumea = {
      url = "github:nix-community/haumea";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    namaka = {
      url = "github:nix-community/namaka";
      inputs.haumea.follows = "haumea";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{ flake-parts, ... }:
    let
      # flake output（lib）とテスト入力で同一実体を共有する（self 参照を避ける）。
      layatLib = import ./lib;
      # flake / Go バイナリ共通のバージョン（VERSION の semver から末尾改行を落とす）。
      version = inputs.nixpkgs.lib.strings.trim (builtins.readFile ./VERSION);
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
      ];
    in
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        inputs.treefmt-nix.flakeModule
        inputs.nix-unit.modules.flake.default
        # layat output を perSystem から転置する flake-parts module。
        # 循環参照を避けるため、公開（flake.flakeModules.default）と同一パスを直接 import する。
        ./modules/flake-parts.nix
      ];
      inherit systems;
      perSystem =
        {
          config,
          pkgs,
          lib,
          ...
        }:
        let
          # Go ビルド・lint の入力（go.mod + go.sum + internal/ + cmd/。docs 変更で再ビルドしないよう絞る）。
          goSrc = lib.fileset.toSource {
            root = ./.;
            fileset = lib.fileset.unions [
              ./go.mod
              ./go.sum
              ./internal
              ./cmd
            ];
          };
          # nix sandbox で go ツールを回す環境。goModules を vendor/ に展開してオフライン解決する。
          # build dir 直下は Go が go.mod を無視するため、サブディレクトリで作業する。
          goToolEnv = ''
            export HOME="$TMPDIR"
            export GOCACHE="$TMPDIR/go-cache"
            export GOTOOLCHAIN=local
            # cgo 未使用。サンドボックスに C コンパイラを持ち込まずピュア Go で検査する。
            export CGO_ENABLED=0
            export GOFLAGS=-mod=vendor
            export GOPROXY=off
            mkdir -p build && cd build
            cp -r --no-preserve=mode ${goSrc}/. .
            cp -r --no-preserve=mode ${config.packages.layat.goModules} vendor
          '';
        in
        {
          # layat CLI（cmd/layat）+ 配置エンジン（internal/）。vendorHash は依存変更時に更新する。
          packages.layat = pkgs.buildGoModule {
            pname = "layat";
            inherit version;
            src = goSrc;
            vendorHash = "sha256-blRx10aRnzpZDI0sqLJJB3L5OYj2BDY9XrmauXNAVEM=";
            doCheck = true;
            env.GOTOOLCHAIN = "local";
            # VERSION の値を main.version へ埋め込む。素の go build では "dev" のまま。
            ldflags = [
              "-X"
              "main.version=${version}"
            ];
            # go test を -race -coverprofile 付きで回し、func サマリを出す（閾値ゲートは持たない）。
            # ldflags を go test へ渡さない（TestVersionDefault は main.version="dev" を前提にする）。
            checkPhase = ''
              runHook preCheck
              go test -race -coverprofile="$TMPDIR/cover.out" ./...
              go tool cover -func="$TMPDIR/cover.out" | tee "$TMPDIR/coverage-func.txt"
              runHook postCheck
            '';
            # coverprofile / func レポートを成果物へ同梱し、CI が cache hit でも $out から読めるようにする。
            postInstall = ''
              install -Dm644 "$TMPDIR/cover.out" "$out/share/layat/coverage/cover.out"
              install -Dm644 "$TMPDIR/coverage-func.txt" "$out/share/layat/coverage/coverage-func.txt"
            '';
            # ビルド済みバイナリの `--version` に VERSION の値が埋め込まれていることを確かめる。
            doInstallCheck = true;
            installCheckPhase = ''
              runHook preInstallCheck
              got=$("$out/bin/layat" --version)
              case "$got" in
                "layat version ${version}") : ;;
                *) echo "FAIL: layat --version = '$got', want 'layat version ${version}'"; exit 1 ;;
              esac
              runHook postInstallCheck
            '';
            meta = {
              description = "Lays contents at root-relative targets, as the manifest says.";
              mainProgram = "layat";
            };
          };

          # ドッグフーディング用の project mode config。本 repo の docs を .layat-example/docs へ配置する。
          # `nix flake check` は `warning: unknown flake output 'layat'`（exit 0）を出す。
          layat.default = layatLib.mkManifest {
            inherit pkgs;
            root = layatLib.projectRoot;
            entries.".layat-example/docs" = {
              src = inputs.self;
              subpath = "docs";
            };
          };

          treefmt = {
            projectRootFile = "flake.nix";
            programs.nixfmt = {
              enable = true;
              package = pkgs.nixfmt;
            };
            # Go 整形。
            programs.gofmt.enable = true;
          };

          # 静的解析を flake check に載せる。
          checks.go-vet = pkgs.runCommandLocal "layat-go-vet" { nativeBuildInputs = [ pkgs.go ]; } ''
            ${goToolEnv}
            go vet ./...
            touch "$out"
          '';
          checks.golangci-lint =
            pkgs.runCommandLocal "layat-golangci-lint"
              {
                nativeBuildInputs = [
                  pkgs.go
                  pkgs.golangci-lint
                ];
              }
              ''
                ${goToolEnv}
                export GOLANGCI_LINT_CACHE="$TMPDIR/golangci-cache"
                golangci-lint run ./...
                touch "$out"
              '';
          # go test（unit + tmpdir 統合テスト）も flake check で回す。
          checks.layat = config.packages.layat;

          # nix-unit: デフォルト適用・manifest 構造の不変条件をアサートする。
          # check は sandbox 内で flake を再 import するため、全 direct input をローカルに渡す。
          nix-unit.inputs = {
            inherit (inputs)
              nixpkgs
              flake-parts
              treefmt-nix
              nix-unit
              haumea
              namaka
              ;
          };
          nix-unit.tests = import ./tests/nix-unit.nix {
            inherit (pkgs) lib;
            layat = layatLib;
          };

          # namaka: normalizeManifest 出力のスナップショット回帰。不一致は評価時に throw する。
          checks.namaka = builtins.seq (inputs.namaka.lib.load {
            src = ./tests/namaka;
            inputs = {
              inherit (pkgs) lib;
              layat = layatLib;
            };
          }) (pkgs.runCommandLocal "layat-namaka-snapshots" { } "touch \"$out\"");

          # HM モジュールの評価アサート。homeManagerConfiguration を評価し、activation の配線と
          # manifest の内容を検証する。実 activate は e2e が担う。
          checks.hm-module =
            let
              # store hash 揺れを避ける fake な flake-input 相当（nix-unit / namaka と同じ test double）。
              fakeSrc = {
                outPath = "/nix/store/00000000000000000000000000000000-fake-src";
              };
              skill = {
                src = fakeSrc;
                subpath = "skills/nix";
              };
              # 共通の土台に layat の設定だけを差し替えて評価し、config を返す。
              mkHm =
                layatCfg:
                (inputs.home-manager.lib.homeManagerConfiguration {
                  inherit pkgs;
                  modules = [
                    ./modules/home-manager.nix
                    {
                      # ラッパー（flake.homeManagerModules.default）と同じく pin 版 layat を注入する。
                      _module.args.layatPackage = config.packages.layat;
                      home.username = "layat-test";
                      home.homeDirectory = "/home/layat-test";
                      home.stateVersion = "24.05";
                      # nixpkgs と HM の release 文字列ずれによる無害な warning を抑制する。
                      home.enableNixpkgsReleaseCheck = false;
                      layat = {
                        enable = true;
                      }
                      // layatCfg;
                    }
                  ];
                }).config;
              # suffix 省略時に既定値 "layat-backup" が activation に渡ることも検証する。
              multi = mkHm {
                configs.default.entries.".claude/skills/nix" = skill;
                configs.work.entries.".config/work" = skill;
                backup.enable = true;
              };
              sugar = mkHm { entries.".claude/skills/nix" = skill; };
              canonical = mkHm { configs.default.entries.".claude/skills/nix" = skill; };
              empty = mkHm { };
              jobs3 = mkHm {
                configs.default.entries.".claude/skills/nix" = skill;
                jobs = 3;
              };
              recopy = mkHm {
                configs.default.entries.".claude/skills/nix" = skill;
                recopy = true;
              };
              # home.activation の dag entry の生スクリプト。
              activationScript = pkgs.writeText "layat-activation" multi.home.activation.layat.data;
            in
            pkgs.runCommandLocal "layat-hm-module-check"
              {
                sugarActivation = sugar.home.activation.layat.data;
                canonicalActivation = canonical.home.activation.layat.data;
                sugarWarnings = lib.concatStringsSep "\n" sugar.warnings;
                renameWarning = "has been renamed to `layat.configs.default.entries'";
                emptyLaunches = lib.boolToString (empty.home.activation ? layat);
                jobs3Activation = jobs3.home.activation.layat.data;
                recopyActivation = recopy.home.activation.layat.data;
              }
              ''
                script=${activationScript}

                # (1) home.file へ翻訳せず、全 config の link-farm を辞書順の <name>=<farm> で
                #     1 回の apply --all 起動に渡し、既定の --jobs 0 と --backup=<既定 suffix> を末尾に付けること。
                test "$(grep -c ' apply ' "$script")" -eq 1 \
                  || { echo "FAIL: activation が layat apply を 1 回だけ起動していません"; cat "$script"; exit 1; }
                grep -qE " apply --all --manifest 'default=/nix/store/[a-z0-9]+-layat-manifest' --manifest 'work=/nix/store/[a-z0-9]+-layat-manifest' --jobs 0 --backup=layat-backup$" "$script" \
                  || { echo "FAIL: activation が apply --all --manifest <name>=<farm> … --jobs 0 --backup=<suffix> の形ではありません"; cat "$script"; exit 1; }

                # (2) config ごとに別 store path の manifest が root=homeRoot を pin し、自分の entries だけを持つこと。
                farmDefault=$(grep -oE "default=/nix/store/[^ ']+" "$script" | cut -d= -f2)
                farmWork=$(grep -oE "work=/nix/store/[^ ']+" "$script" | cut -d= -f2)
                test "$farmDefault" != "$farmWork" \
                  || { echo "FAIL: 2 configs の manifest が同じ store path です"; cat "$script"; exit 1; }
                for spec in "$farmDefault:.claude/skills/nix:.config/work" "$farmWork:.config/work:.claude/skills/nix"; do
                  IFS=: read -r farm own other <<<"$spec"
                  json="$farm/manifest.json"
                  test -f "$json" || { echo "FAIL: $json がありません"; exit 1; }
                  grep -q '"rootKind":"home"' "$json" \
                    || { echo "FAIL: manifest が homeRoot を pin していません"; cat "$json"; exit 1; }
                  grep -qF "\"$own\"" "$json" \
                    || { echo "FAIL: configs.<name>.entries が manifest に反映されていません"; cat "$json"; exit 1; }
                  ! grep -qF "\"$other\"" "$json" \
                    || { echo "FAIL: 他の config の entries が manifest に混入しています"; cat "$json"; exit 1; }
                done

                # (3) layat.entries のみの config は configs.default.entries と同一 manifest を生み、
                #     rename の deprecation warning を出すこと。
                test "$sugarActivation" = "$canonicalActivation" \
                  || { echo "FAIL: layat.entries と configs.default.entries の activation が異なります"; echo "$sugarActivation"; echo "$canonicalActivation"; exit 1; }
                grep -qF -- "$renameWarning" <<<"$sugarWarnings" \
                  || { echo "FAIL: layat.entries の deprecation warning がありません"; echo "$sugarWarnings"; exit 1; }

                # (4) layat.jobs の値がそのまま --jobs に乗ること。
                grep -qE " --manifest 'default=[^']+' --jobs 3$" <<<"$jobs3Activation" \
                  || { echo "FAIL: layat.jobs = 3 が --jobs 3 として渡っていません"; echo "$jobs3Activation"; exit 1; }

                # (5) layat.recopy = true のときだけ --recopy が同じ起動に乗ること。
                grep -qE " apply --all --manifest 'default=[^']+' --jobs 0 --recopy$" <<<"$recopyActivation" \
                  || { echo "FAIL: layat.recopy = true が --recopy として渡っていません"; echo "$recopyActivation"; exit 1; }
                ! grep -qF -- "--recopy" "$script" && ! grep -qF -- "--recopy" <<<"$jobs3Activation" \
                  || { echo "FAIL: layat.recopy の既定で --recopy が渡っています"; cat "$script"; echo "$jobs3Activation"; exit 1; }

                # (6) config が 1 つも無いとき activation は layat を起動しないこと。
                test "$emptyLaunches" = false \
                  || { echo "FAIL: configs が空でも activation が layat を起動します"; exit 1; }

                touch "$out"
              '';
        };
      flake = {
        lib = layatLib;

        # `layat init <template>` / `nix flake init -t <ref>#<template>` で展開する starter テンプレ。
        templates = {
          standalone = {
            path = ./templates/standalone;
            description = "layat standalone config（homeRoot 例 + バリエーションコメント）";
          };
          project = {
            path = ./templates/project;
            description = "layat project config（projectRoot + devShell + shellHook + .gitignore）";
          };
          default = inputs.self.templates.project;
        };

        # flake-parts module を consumer 向けに公開する。consumer は
        # `imports = [ inputs.layat.flakeModules.default ]` してから `perSystem.layat.<name> = ...` を書く。
        flakeModules.default = ./modules/flake-parts.nix;

        # HM モジュールに pin 版 layat CLI を _module.args として注入する薄いラッパー。
        homeManagerModules.default =
          { pkgs, ... }:
          {
            imports = [ ./modules/home-manager.nix ];
            _module.args.layatPackage = inputs.self.packages.${pkgs.stdenv.hostPlatform.system}.layat;
          };
        nixosModules.default = ./modules/nixos.nix;
        darwinModules.default = ./modules/nix-darwin.nix;
      };
    };
}
