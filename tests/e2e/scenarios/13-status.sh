#!/usr/bin/env bash
# status: project mode と home mode の config を 1 つずつ apply し、`layat status` が記録基底を
# 走査して両方を列挙すること・`[name]` で絞れること・`--json` の info を検証する。
# 隔離した system の記録基底に置いた記録が `system:<name>` で出ることも見る。
set -euo pipefail
source "$(dirname "$0")/../lib.sh"
e2e_isolate

PROJ="$E2E_WORK/cfg"
mkdir -p "$PROJ/srcrepo/a"
echo "AAA" >"$PROJ/srcrepo/a/file"

cat >"$PROJ/flake.nix" <<EOF
{
$(e2e_flake_inputs)
  outputs = { self, nixpkgs, layat }: {
    layat = nixpkgs.lib.genAttrs $E2E_SYSTEMS (system: {
      home = layat.lib.mkManifest {
        pkgs = nixpkgs.legacyPackages.\${system};
        root = layat.lib.homeRoot;
        entries.".cfg/a" = { src = ./srcrepo; subpath = "a"; };
      };
      proj = layat.lib.mkManifest {
        pkgs = nixpkgs.legacyPackages.\${system};
        root = layat.lib.projectRoot;
        entries.".layat-out/proj" = { src = ./srcrepo; subpath = "a"; };
      };
    });
  };
}
EOF

cd "$PROJ"
git init -q
git -c user.email=e2e@layat.test -c user.name=e2e add -A
git -c user.email=e2e@layat.test -c user.name=e2e commit -qm init
PROJ_ROOT="$(pwd -P)"

e2e_step "home mode の home と project mode の proj を apply"
layat apply home
layat apply proj
assert_symlink "$HOME/.cfg/a"
assert_symlink "$PROJ/.layat-out/proj"

e2e_step "layat status: 両方の subject を見出し行で出す"
STATUS_CODE=0
STATUS_TEXT="$(layat status)" || STATUS_CODE=$?
if [ "$STATUS_CODE" -eq 0 ]; then
	e2e_pass "exit 0: layat status"
else
	e2e_fail "exit $STATUS_CODE (期待 0): layat status"
fi
HEADERS="$(printf '%s\n' "$STATUS_TEXT" | grep '^# ' | sed 's/ (applied .*)$//' || true)"
HEADERS_WANT="$(printf '# home:home\n# project:proj@%s' "$PROJ_ROOT")"
if [ "$HEADERS" = "$HEADERS_WANT" ]; then
	e2e_pass "見出しは home:home、project:proj@<root> の順"
else
	e2e_fail "見出しが期待と不一致: '$HEADERS'（期待: '$HEADERS_WANT'・全体: '$STATUS_TEXT'）"
fi
if printf '%s\n' "$STATUS_TEXT" | grep -q "^\.cfg/a	symlink	/nix/store/"; then
	e2e_pass "home の entry 行が <target>\\t<method>\\t<dest> で出る"
else
	e2e_fail "home の entry 行が無い: '$STATUS_TEXT'"
fi

e2e_step "layat status proj: proj だけに絞る"
FILTERED_CODE=0
FILTERED_TEXT="$(layat status proj)" || FILTERED_CODE=$?
FILTERED="$(printf '%s\n' "$FILTERED_TEXT" | grep '^# ' | sed 's/ (applied .*)$//' || true)"
if [ "$FILTERED_CODE" -eq 0 ] && [ "$FILTERED" = "# project:proj@$PROJ_ROOT" ]; then
	e2e_pass "status proj は exit 0 で project:proj@<root> だけを出す"
else
	e2e_fail "status proj が期待と不一致: exit=$FILTERED_CODE 見出し='$FILTERED'"
fi

e2e_step "layat status nope: 一致 0 件は stdout 空・stderr に 1 行・exit 0"
NONE_CODE=0
NONE_OUT="$(layat status nope 2>"$E2E_WORK/none.err")" || NONE_CODE=$?
if [ "$NONE_CODE" -eq 0 ] && [ -z "$NONE_OUT" ] && grep -q 'no applied config named nope' "$E2E_WORK/none.err"; then
	e2e_pass "一致 0 件は exit 0・stdout 空・stderr に通知"
else
	e2e_fail "一致 0 件の扱いが違う: exit=$NONE_CODE stdout='$NONE_OUT' stderr='$(cat "$E2E_WORK/none.err")'"
fi

e2e_step "layat status --json: info の mode・root・entries の target"
ENV_STATUS="$E2E_WORK/status.json"
run_json 0 "$ENV_STATUS" status
assert_json "$ENV_STATUS" "subject は home:home と project:proj@<root>" \
	"[.results[].subject.name] == [\"home:home\", \"project:proj@$PROJ_ROOT\"]"
assert_json "$ENV_STATUS" "home の info は mode=home・root 空・entries の target が .cfg/a" \
	'first(.results[] | select(.subject.name == "home:home")).result.info
	 | .mode == "home" and .root == "" and ([.entries[].target] == [".cfg/a"])'
assert_json "$ENV_STATUS" "proj の info は mode=project・root が git toplevel・entries の target が .layat-out/proj" \
	"first(.results[] | select(.subject.name | startswith(\"project:proj@\"))).result.info
	 | .mode == \"project\" and .root == \"$PROJ_ROOT\" and ([.entries[].target] == [\".layat-out/proj\"])"
assert_json "$ENV_STATUS" "items=[]・generation スロット無し・status=success・dryRun=false" \
	'(.results | all(.result.items == [] and (has("generation") | not)))
	 and .status == "success" and .dryRun == false and .command == "status"'

e2e_step "隔離した system の記録基底に置いた記録が system:<name> で出る"
# home の記録をそのまま system 基底の <name> 直キーへ置く。直キーの mode は基底で決まる。
mkdir -p "$LAYAT_SYSTEM_STATE_DIR/layat/sys"
cp "$XDG_STATE_HOME/layat/home/last-applied.json" "$LAYAT_SYSTEM_STATE_DIR/layat/sys/last-applied.json"
ENV_SYS="$E2E_WORK/status-sys.json"
run_json 0 "$ENV_SYS" status
assert_json "$ENV_SYS" "system 基底の記録が user 基底の後に system:sys で並ぶ" \
	"[.results[].subject.name] == [\"home:home\", \"project:proj@$PROJ_ROOT\", \"system:sys\"]"
assert_json "$ENV_SYS" "system:sys の info.mode は system" \
	'first(.results[] | select(.subject.name == "system:sys")).result.info.mode == "system"'
SYS_CODE=0
SYS_TEXT="$(layat status sys)" || SYS_CODE=$?
SYS_FILTERED="$(printf '%s\n' "$SYS_TEXT" | grep '^# ' | sed 's/ (applied .*)$//' || true)"
if [ "$SYS_CODE" -eq 0 ] && [ "$SYS_FILTERED" = "# system:sys" ]; then
	e2e_pass "status sys は exit 0 で system:sys だけを出す"
else
	e2e_fail "status sys が期待と不一致: exit=$SYS_CODE 見出し='$SYS_FILTERED'"
fi

e2e_finish
