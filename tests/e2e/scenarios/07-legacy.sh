#!/usr/bin/env bash
# legacy entrypoint (shell.nix, passthru canonical form): `layat apply` / `apply --all` /
# plain `nix-shell` compatibility.
set -euo pipefail
source "$(dirname "$0")/../lib.sh"
e2e_isolate
e2e_pin_nix_path

PROJ="$E2E_WORK/proj"
mkdir -p "$PROJ/srcrepo/skills/nix"
echo "SKILLBODY" >"$PROJ/srcrepo/skills/nix/SKILL.md"

cat >"$PROJ/shell.nix" <<EOF
{ pkgs ? import <nixpkgs> {}
, layat ? { lib = import $REPO_ROOT/lib; }
}:
pkgs.mkShell {
  packages = [ ];
  shellHook = "layat apply docs --no-wait";
  passthru.layat = {
    docs = layat.lib.mkManifest {
      inherit pkgs;
      root = layat.lib.projectRoot;
      entries.".layat-out/docs" = { src = ./srcrepo; subpath = "skills/nix"; };
    };
    extra = layat.lib.mkManifest {
      inherit pkgs;
      root = layat.lib.projectRoot;
      entries.".layat-out/extra" = { src = ./srcrepo; subpath = "skills/nix"; };
    };
  };
}
EOF

cd "$PROJ"
git init -q
git -c user.email=e2e@layat.test -c user.name=e2e add -A
git -c user.email=e2e@layat.test -c user.name=e2e commit -qm init

e2e_step "layat apply docs（legacy shell.nix・passthru 形）"
layat apply docs

e2e_step "git toplevel 配下に配置されたか"
TARGET="$PROJ/.layat-out/docs"
assert_symlink "$TARGET"
assert_file_eq "$TARGET/SKILL.md" "SKILLBODY"
# legacy -f eval では相対 path リテラル `src = ./srcrepo` が作業木パスのまま解決されるため、
# store symlink であることまでは assert しない。

e2e_step "layat apply --all（passthru.layat.* を一括適用）"
layat apply --all
assert_symlink "$PROJ/.layat-out/extra"
assert_file_eq "$PROJ/.layat-out/extra/SKILL.md" "SKILLBODY"

e2e_step "素の nix-shell がそのまま動く（passthru に壊されない）＋ shellHook が apply を kick する"
rm -rf "$PROJ/.layat-out"
nix-shell --run true "$PROJ/shell.nix"
assert_symlink "$TARGET"
assert_file_eq "$TARGET/SKILL.md" "SKILLBODY"

e2e_step "legacy entrypoint でも gitignore --json が info.paths を運ぶ"
ENV_LEGACY="$E2E_WORK/gitignore-legacy.json"
run_json 0 "$ENV_LEGACY" gitignore docs
assert_json "$ENV_LEGACY" "info.paths が anchor 形・items=[]" \
	'.results[0].result.info.paths == ["/.layat-out/docs"] and .results[0].result.items == []'

e2e_finish
