#!/usr/bin/env bash
# apply --all の並列化（→ ADR-0038 / ADR-0039・issue #155）: 集約順が辞書順で並列度に依らないこと、
# cross-config target 衝突が build 前に error で止まること、--no-wait 併用で全 config が適用されること。
# fixture ごとに別 git repo（= 別 project root）を切り、lock・.root backref・配置先を混ぜない。
set -euo pipefail
source "$(dirname "$0")/../lib.sh"
e2e_isolate

GIT=(git -c user.email=e2e@layat.test -c user.name=e2e)

# mk_fixture <dir> <name>=<target>...: projectRoot の config を並べた fixture flake を git repo に置く。
mk_fixture() {
	local dir="$1" spec configs=""
	shift
	mkdir -p "$dir/srcrepo/d"
	echo "BODY" >"$dir/srcrepo/d/file"
	for spec in "$@"; do
		configs+="      ${spec%%=*} = layat.lib.mkManifest {
        pkgs = nixpkgs.legacyPackages.\${system};
        root = layat.lib.projectRoot;
        entries.\"${spec#*=}\" = { src = ./srcrepo; subpath = \"d\"; };
      };
"
	done
	cat >"$dir/flake.nix" <<EOF
{
$(e2e_flake_inputs)
  outputs = { self, nixpkgs, layat }: {
    layat = nixpkgs.lib.genAttrs $E2E_SYSTEMS (system: {
$configs    });
  };
}
EOF
	(cd "$dir" && git init -q && "${GIT[@]}" add -A && "${GIT[@]}" commit -qm init)
}

# run_code <期待 exit> <stderr 保存先> <layat 引数...>: stderr を保存して exit code を突き合わせる。
run_code() {
	local want="$1" err="$2" code=0
	shift 2
	layat "$@" 2>"$err" || code=$?
	if [ "$code" -eq "$want" ]; then
		e2e_pass "exit $code: layat $*"
	else
		e2e_fail "exit $code (期待 $want): layat $* ($(cat "$err"))"
	fi
}

# ---- (1) 辞書順集約と並列度非依存 ---------------------------------------------
# config 名は宣言順と辞書順をずらす（Nix の attrset は名前順に並ぶが、宣言順の偶然一致を避ける）。
LEX="$E2E_WORK/lex"
mk_fixture "$LEX" zeta=out/zeta alpha=out/alpha mid=out/mid delta=out/delta
cd "$LEX"
LEX_ORDER='["alpha", "delta", "mid", "zeta"]'

e2e_step "apply --all -v --json: results[] と done 行が辞書順（→ REQ-4cbd9a0d）"
ENV_LEX="$E2E_WORK/lex.json"
ERR_LEX="$E2E_WORK/lex.err"
code=0
layat apply --all -v --json >"$ENV_LEX" 2>"$ERR_LEX" || code=$?
if [ "$code" -eq 0 ]; then e2e_pass "exit 0: apply --all -v --json"; else e2e_fail "exit $code: $(cat "$ERR_LEX")"; fi
assert_json "$ENV_LEX" "results[].subject.name が辞書順・全 success" \
	"[.results[].subject.name] == $LEX_ORDER and all(.results[]; .status == \"success\")"
# 集約行 `layat: apply --all done (...)` は除き、config ごとの done 行だけを順に取る。
DONE_ORDER="$(sed -n 's/^layat: apply \([^- ][^ ]*\) done .*/\1/p' "$ERR_LEX" | paste -sd' ')"
if [ "$DONE_ORDER" = "alpha delta mid zeta" ]; then
	e2e_pass "-v の done 行が辞書順: $DONE_ORDER"
else
	e2e_fail "-v の done 行が辞書順でない: '$DONE_ORDER'"
fi
for n in alpha delta mid zeta; do assert_symlink "$LEX/out/$n"; done

e2e_step "定常状態で既定並列度と --jobs 1 の results[] が同一（→ ADR-0039）"
ENV_DEFAULT="$E2E_WORK/lex-default.json"
ENV_SERIAL="$E2E_WORK/lex-serial.json"
run_json 0 "$ENV_DEFAULT" apply --all
run_json 0 "$ENV_SERIAL" apply --all --jobs 1
# 実行時刻（startedAt / finishedAt）だけは実行ごとに変わるため落として比べる。
STABLE='.results | walk(if type == "object" then del(.startedAt, .finishedAt) else . end)'
if diff <(jq -S "$STABLE" "$ENV_DEFAULT") <(jq -S "$STABLE" "$ENV_SERIAL") >"$E2E_WORK/lex.diff"; then
	e2e_pass "results[] が実行時刻を除き同一"
else
	e2e_fail "results[] が並列度で変わる: $(cat "$E2E_WORK/lex.diff")"
fi
assert_json "$ENV_SERIAL" "--jobs 1 でも辞書順" "[.results[].subject.name] == $LEX_ORDER"

# ---- (2) cross-config target 衝突の前段 error --------------------------------
CONF="$E2E_WORK/conflict"
mk_fixture "$CONF" left=shared/cfg right=shared/cfg solo=solo/cfg
cd "$CONF"

for mode in "" "--dryrun"; do
	e2e_step "apply --all $mode: 衝突は exit 1 で build 前に止まる（→ ADR-0038）"
	ERR_CONF="$E2E_WORK/conflict${mode}.err"
	# shellcheck disable=SC2086 # mode は空か 1 語。空のときに空引数を渡さない。
	run_code 1 "$ERR_CONF" apply --all $mode
	for needle in left right shared/cfg; do
		if grep -qF "$needle" "$ERR_CONF"; then
			e2e_pass "stderr が $needle を含む"
		else
			e2e_fail "stderr に $needle が無い: $(cat "$ERR_CONF")"
		fi
	done
	assert_absent "$CONF/shared"
	assert_absent "$CONF/solo"
done

e2e_step "apply --all --json: 衝突はトップレベル errors[] に載り results[] は空"
ENV_CONF="$E2E_WORK/conflict.json"
run_json 1 "$ENV_CONF" apply --all
assert_json "$ENV_CONF" "status=error・results=[]・トップ errors[] が E_LAYAT_FAILED 1 件" \
	'.status == "error" and .results == [] and (.errors | length) == 1 and .errors[0].code == "E_LAYAT_FAILED"'
assert_json "$ENV_CONF" "エラーメッセージが両 config 名と衝突 target を含む" \
	'.errors[0].message | contains("left") and contains("right") and contains("shared/cfg")'
assert_absent "$CONF/shared"
assert_absent "$CONF/solo"

# ---- (3) --no-wait 併用 -------------------------------------------------------
# 初回 apply として走らせる（配置と同じ root の .root backref の並行書き込みを実際に通す）。
NW="$E2E_WORK/nowait"
mk_fixture "$NW" one=nw/one three=nw/three two=nw/two
cd "$NW"

e2e_step "apply --all --no-wait -v: 進行中の apply が無ければ全 config が applied"
ERR_NW="$E2E_WORK/nowait.err"
run_code 0 "$ERR_NW" apply --all --no-wait -v
if grep -qF "layat: apply --all done (applied 3 / skipped 0 / failed 0 / selected 3)" "$ERR_NW"; then
	e2e_pass "集約行が applied 3 / skipped 0 / failed 0"
else
	e2e_fail "集約行が期待と違う: $(cat "$ERR_NW")"
fi
for n in one three two; do assert_symlink "$NW/nw/$n"; done

e2e_step "apply --all --no-wait --json: 全 subject が success"
ENV_NW="$E2E_WORK/nowait.json"
run_json 0 "$ENV_NW" apply --all --no-wait
assert_json "$ENV_NW" "3 config が辞書順で全 success・errors[] なし" \
	'[.results[].subject.name] == ["one", "three", "two"] and all(.results[]; .status == "success") and (has("errors") | not)'

e2e_finish
