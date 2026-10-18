#!/usr/bin/env bash
# system mode: root 取得下で `--root` 上書きの sysroot へ apply → 世代 2 → rollback → reset を回し、
# 隔離した system 基底に profile / backref / 記録が立つことと、非 root の変更系が E_PERMISSION で
# 何も変えずに止まり、読み取り系（apply --dryrun / list-generations）は通ることをアサート。
set -euo pipefail
source "$(dirname "$0")/../lib.sh"
e2e_isolate
e2e_root_init

CFG="$E2E_WORK/cfg"
SYSROOT="$E2E_WORK/sysroot"
mkdir -p "$CFG/srcrepo/a" "$CFG/srcrepo/b" "$SYSROOT"
echo "AAA" >"$CFG/srcrepo/a/file"
echo "BBB" >"$CFG/srcrepo/b/file"

# target / subpath を引数に取り fixture flake を書き出す（世代ごとに entry を入れ替える）。
write_flake() {
	cat >"$CFG/flake.nix" <<EOF
{
$(e2e_flake_inputs)
  outputs = { self, nixpkgs, layat }: {
    layat = nixpkgs.lib.genAttrs $E2E_SYSTEMS (system: {
      sys = layat.lib.mkManifest {
        pkgs = nixpkgs.legacyPackages.\${system};
        root = layat.lib.systemRoot;
        entries."$1" = { src = ./srcrepo; subpath = "$2"; };
      };
    });
  };
}
EOF
}

# root として layat を実行する。
as_root() { E2E_AS_ROOT=1 "$@"; }

# 世代の本数は root で数える（非 root の読み取りは下の検査で別に見る）。
gens_count() {
	local out
	out="$(as_root layat list-generations sys --root "$SYSROOT")" || return 1
	printf '%s\n' "$out" | grep -c . || true
}

# sysroot 配下・profile ディレクトリ・記録を 1 つの文字列に写す（非 root の不変判定に使う）。
snapshot() {
	find "$SYSROOT" "$SERIES/sys" -printf '%p %y %l\n' | sort
	cat "$RECORD"
}

cd "$CFG"
write_flake ".cfg/a" "a"

e2e_step "世代 1: root で apply（entry a）→ sysroot 配下に配置"
as_root layat apply sys --root "$SYSROOT"
assert_symlink "$SYSROOT/.cfg/a"
assert_file_eq "$SYSROOT/.cfg/a/file" "AAA"

e2e_step "system 基底に <roothash>/sys の profile・世代リンクと backref .root が立つ"
SERIES=""
for d in "$LAYAT_SYSTEM_PROFILE_BASE"/*/; do
	if [ -f "$d/.root" ] && [ "$(cat "$d/.root")" = "$SYSROOT" ]; then SERIES="${d%/}"; fi
done
if [ -n "$SERIES" ]; then
	e2e_pass "backref .root が sysroot を指す系列: $SERIES"
else
	e2e_fail "sysroot を指す backref が system 基底に無い"
	exit 1
fi
ROOTHASH="$(basename "$SERIES")"
assert_symlink "$SERIES/sys/profile"
assert_symlink "$SERIES/sys/profile-1-link"
RECORD="$LAYAT_SYSTEM_STATE_DIR/layat/$ROOTHASH/sys/last-applied.json"
assert_exists "$RECORD"
assert_json "$RECORD" "記録の linkFarm が profile の指す link-farm と一致する" \
	".linkFarm == \"$(readlink -f "$SERIES/sys/profile")\""

e2e_step "list-generations: 単発は世代を mode=system で出し、--all は <roothash> 系列を除外する"
if [ "$(gens_count)" -eq 1 ]; then e2e_pass "世代が 1 件"; else e2e_fail "世代が 1 件でない"; fi
ENV_GENS="$E2E_WORK/gens.json"
as_root run_json 0 "$ENV_GENS" list-generations sys --root "$SYSROOT"
assert_json "$ENV_GENS" "単発の subject は素の sys で info.mode == system" \
	'.results[0] | .subject.name == "sys" and .result.info.mode == "system"'
ENV_ALL="$E2E_WORK/gens-all.json"
run_json 0 "$ENV_ALL" list-generations --all
# --all は基底直下の profile だけを列挙する。--root 上書きの系列は <roothash>/sys なので現れない。
assert_json "$ENV_ALL" "--all の results に system:sys は無い（<roothash> 系列は除外）" \
	'[.results[].subject.name] == []'

e2e_step "世代 2: entry を b に入れ替えて root で再 apply"
write_flake ".cfg/b" "b"
as_root layat apply sys --root "$SYSROOT"
assert_symlink "$SYSROOT/.cfg/b"
assert_absent "$SYSROOT/.cfg/a"
assert_symlink "$SERIES/sys/profile-2-link"
if [ "$(gens_count)" -eq 2 ]; then e2e_pass "世代が 2 件"; else e2e_fail "世代が 2 件でない"; fi

e2e_step "非 root: apply / rollback / reset --yes は E_PERMISSION で止まり何も変えない"
BEFORE="$(snapshot)"
i=0
for cmd in "apply sys" "rollback sys" "reset sys --yes"; do
	i=$((i + 1))
	env_file="$E2E_WORK/nonroot-$i.json"
	# shellcheck disable=SC2086
	run_json 1 "$env_file" $cmd --root "$SYSROOT"
	assert_json "$env_file" "$cmd: subject sys の errors[] が E_PERMISSION" \
		'any(.results[] | select(.subject.name == "sys") | .errors[]; .code == "E_PERMISSION")'
done
if [ "$(snapshot)" = "$BEFORE" ]; then
	e2e_pass "非 root の変更系の前後で sysroot・profile ディレクトリ・記録が不変"
else
	e2e_fail "非 root の変更系が sysroot・profile ディレクトリ・記録のどれかを変えた"
fi

e2e_step "非 root: apply --dryrun と list-generations は exit 0"
run_json 0 "$E2E_WORK/nonroot-dryrun.json" apply sys --dryrun --root "$SYSROOT"
run_json 0 "$E2E_WORK/nonroot-gens.json" list-generations sys --root "$SYSROOT"
if [ "$(snapshot)" = "$BEFORE" ]; then
	e2e_pass "読み取り系の前後でも不変"
else
	e2e_fail "読み取り系が sysroot・profile ディレクトリ・記録のどれかを変えた"
fi

e2e_step "root で rollback → 世代 1 の配置（entry a）へ戻る"
as_root layat rollback sys --root "$SYSROOT"
assert_symlink "$SYSROOT/.cfg/a"
assert_file_eq "$SYSROOT/.cfg/a/file" "AAA"
assert_absent "$SYSROOT/.cfg/b"
assert_symlink "$SERIES/sys/profile" "profile-1-link"

e2e_step "root で reset --yes → 配置が撤去され、profile と世代は動かない"
PROFILE_BEFORE="$(readlink "$SERIES/sys/profile")"
as_root layat reset sys --yes --root "$SYSROOT"
assert_absent "$SYSROOT/.cfg/a"
assert_absent "$SYSROOT/.cfg/b"
assert_symlink "$SERIES/sys/profile" "$PROFILE_BEFORE"
if [ "$(gens_count)" -eq 2 ]; then e2e_pass "reset 後も世代が 2 件"; else e2e_fail "reset が世代の本数を動かした"; fi

e2e_finish
