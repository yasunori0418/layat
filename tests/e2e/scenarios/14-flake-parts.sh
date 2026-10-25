#!/usr/bin/env bash
# flake-parts + mkApplyCommand: projectRoot と homeRoot の config を 1 つずつ持つ fixture の devShell の
# shellHook に `mkApplyCommand { …; root = "project"; }` を埋め、`nix develop -c true` だけで
# project 側が git toplevel 配下に配置され、home 側は `$HOME` に何も出ないことを検証する。
set -euo pipefail
source "$(dirname "$0")/../lib.sh"
e2e_isolate

PROJ="$E2E_WORK/proj"
mkdir -p "$PROJ/srcrepo/a"
echo "AAA" >"$PROJ/srcrepo/a/file"

cat >"$PROJ/flake.nix" <<EOF
{
$(e2e_flake_inputs)
  inputs.flake-parts.follows = "layat/flake-parts";
  outputs = inputs@{ flake-parts, layat, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [ layat.flakeModules.default ];
      systems = $E2E_SYSTEMS;
      perSystem = { config, pkgs, system, ... }: {
        layat.proj = layat.lib.mkManifest {
          inherit pkgs;
          root = layat.lib.projectRoot;
          entries.".layat-out/proj" = { src = ./srcrepo; subpath = "a"; };
        };
        layat.home = layat.lib.mkManifest {
          inherit pkgs;
          root = layat.lib.homeRoot;
          entries.".cfg/home" = { src = ./srcrepo; subpath = "a"; };
        };
        devShells.default = pkgs.mkShell {
          shellHook = (layat.lib.mkApplyCommand {
            inherit pkgs;
            layatPackage = layat.packages.\${system}.layat;
            manifests = config.layat;
            root = "project";
          }) + " --no-wait";
        };
      };
    };
}
EOF

cd "$PROJ"
git init -q
git -c user.email=e2e@layat.test -c user.name=e2e add -A
git -c user.email=e2e@layat.test -c user.name=e2e commit -qm init
PROJ_ROOT="$(pwd -P)"

e2e_step "サブディレクトリから nix develop -c true（shellHook が apply --all --project-root を走らせる）"
mkdir -p "$PROJ/sub"
(cd "$PROJ/sub" && nix develop "$PROJ" -c true)

e2e_step "project 側は git toplevel 配下に配置される"
assert_symlink "$PROJ_ROOT/.layat-out/proj"
assert_file_eq "$PROJ_ROOT/.layat-out/proj/file" "AAA"
assert_absent "$PROJ/sub/.layat-out"

e2e_step "home 側は --project-root で外れ、隔離した \$HOME に何も出ない"
assert_absent "$HOME/.cfg"
# nix 自身が作る .nix-* / .cache は除いて数える。
LEFT="$(find "$HOME" -mindepth 1 \( -name '.nix-*' -o -name .cache \) -prune -o -print)"
if [ -z "$LEFT" ]; then
	e2e_pass "\$HOME に配置物なし"
else
	e2e_fail "\$HOME に配置物がある: $(echo "$LEFT" | tr '\n' ' ')"
fi

e2e_finish
