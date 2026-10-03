#!/usr/bin/env bash
# 生成器の選択（→ ADR-0056）: --manifest 経路が LAYAT_GENERATOR / layat.toml に左右されないこと
# （HM 等の activation は実行環境の環境変数・cwd を制御できない）と、未知の生成器名・不正な
# 設定ファイルが exit 1 + 1 行・--json で E_INPUT になること、正しい指定では通常どおり配置することを
# アサート。
set -euo pipefail
source "$(dirname "$0")/../lib.sh"
e2e_isolate

PROJ="$E2E_WORK/proj"
mkdir -p "$PROJ/srcrepo/skills/nix"
echo "SKILLBODY" >"$PROJ/srcrepo/skills/nix/SKILL.md"

cat >"$PROJ/flake.nix" <<EOF
{
$(e2e_flake_inputs)
  outputs = { self, nixpkgs, layat }: {
    layat = nixpkgs.lib.genAttrs $E2E_SYSTEMS (system: {
      docs = layat.lib.mkManifest {
        pkgs = nixpkgs.legacyPackages.\${system};
        root = layat.lib.projectRoot;
        entries.".layat-out/docs" = { src = ./srcrepo; subpath = "skills/nix"; };
      };
    });
  };
}
EOF

cd "$PROJ"
git init -q
git -c user.email=e2e@layat.test -c user.name=e2e add -A
git -c user.email=e2e@layat.test -c user.name=e2e commit -qm init

# 入力不正の --json は subject 確定前ならトップ errors[]、確定後なら subject の errors[] に載る。
# どちらの層でも E_INPUT であることを見る。
E_INPUT_EXPR='[.errors[]?, .results[]?.errors[]?] | any(.code == "E_INPUT")'

# 人間向けは exit 1 + stderr 1 行（→ ADR-0056 §5）。
assert_input_error_line() { # $1: 説明, $2...: layat 引数
	local desc="$1" err="$E2E_WORK/stderr.txt" code=0
	shift
	layat "$@" 2>"$err" || code=$?
	if [ "$code" -eq 1 ] && [ "$(wc -l <"$err")" -eq 1 ]; then
		e2e_pass "$desc: exit 1 + 1 行（$(cat "$err")）"
	else
		e2e_fail "$desc: exit $code・stderr $(wc -l <"$err") 行（期待 exit 1 + 1 行）: $(cat "$err")"
	fi
}

e2e_step "(1) LAYAT_GENERATOR=bogus と未知キー入り layat.toml があっても apply --manifest は配置する"
FARM="$(nix build ".#layat.$E2E_SYSTEM.docs" --no-link --print-out-paths)"
printf 'generator = "nix"\nbogus = true\n' >"$PROJ/layat.toml"
if LAYAT_GENERATOR=bogus layat apply --manifest "$FARM"; then
	e2e_pass "apply --manifest: exit 0"
else
	e2e_fail "apply --manifest は環境変数・設定ファイルを読まずに通るべき"
fi
assert_symlink "$PROJ/.layat-out/docs"
assert_file_eq "$PROJ/.layat-out/docs/SKILL.md" "SKILLBODY"
rm -rf "$PROJ/.layat-out"

e2e_step "(2) apply --generator bogus は exit 1・--json で E_INPUT"
rm -f "$PROJ/layat.toml"
assert_input_error_line "apply --generator bogus" apply --generator bogus docs
ENV_FLAG="$E2E_WORK/flag.json"
run_json 1 "$ENV_FLAG" apply --generator bogus docs
assert_json "$ENV_FLAG" "未知の --generator は E_INPUT" "$E_INPUT_EXPR"
assert_absent "$PROJ/.layat-out/docs"

e2e_step "(3) 未知キー入りの layat.toml で apply は exit 1・--json で E_INPUT"
printf 'generator = "nix"\nbogus = true\n' >"$PROJ/layat.toml"
assert_input_error_line "apply（不正な layat.toml）" apply docs
ENV_TOML="$E2E_WORK/toml.json"
run_json 1 "$ENV_TOML" apply docs
assert_json "$ENV_TOML" "不正な layat.toml は E_INPUT" "$E_INPUT_EXPR"
assert_absent "$PROJ/.layat-out/docs"
rm -f "$PROJ/layat.toml"

e2e_step "(4) LAYAT_GENERATOR=nix の apply は通常どおり配置する"
if LAYAT_GENERATOR=nix layat apply docs; then
	e2e_pass "LAYAT_GENERATOR=nix apply: exit 0"
else
	e2e_fail "LAYAT_GENERATOR=nix の apply は通るべき"
fi
assert_symlink "$PROJ/.layat-out/docs"
assert_file_eq "$PROJ/.layat-out/docs/SKILL.md" "SKILLBODY"

e2e_finish
