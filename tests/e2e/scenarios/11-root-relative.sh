#!/usr/bin/env bash
# root 相対 out-of-store: mkRootRelativeSymlink の entry が target 起点の相対リンクで張られ、
# 再 apply の冪等性・--root 上書き後の root での実在判定と配置・リポジトリ移動後の到達性と移動先での再 apply・
# entry 削除時の stale 除去を満たすことをアサート。
set -euo pipefail
source "$(dirname "$0")/../lib.sh"
e2e_isolate

PROJ="$E2E_WORK/proj"
mkdir -p "$PROJ/assets/skills/demo"
echo "DEMOBODY" >"$PROJ/assets/skills/demo/SKILL.md"

# keep="both" で subpath 無し / 有りの 2 entry、keep="demo" で前者だけを宣言する fixture flake を書く。
write_flake() {
	local entries='entries.".claude/skills/demo" = { src = layat.lib.mkRootRelativeSymlink "assets/skills/demo"; };'
	if [ "$1" = "both" ]; then
		entries="$entries"' entries.".out/sub" = { src = layat.lib.mkRootRelativeSymlink "assets"; subpath = "skills/demo"; };'
	fi
	cat >"$PROJ/flake.nix" <<EOF
{
$(e2e_flake_inputs)
  outputs = { self, nixpkgs, layat }: {
    layat = nixpkgs.lib.genAttrs $E2E_SYSTEMS (system: {
      docs = layat.lib.mkManifest {
        pkgs = nixpkgs.legacyPackages.\${system};
        root = layat.lib.projectRoot;
        $entries
      };
    });
  };
}
EOF
}

# apply の stderr から layat の warning 行だけを抜き出す（nix の評価ログを除く）。
layat_warnings() { grep '^layat: ' "$1" || true; }

cd "$PROJ"
write_flake both
git init -q
git -c user.email=e2e@layat.test -c user.name=e2e add -A
git -c user.email=e2e@layat.test -c user.name=e2e commit -qm both

e2e_step "apply: target 起点の相対リンクで張られ内容へ到達できる"
layat apply docs
assert_symlink "$PROJ/.claude/skills/demo" "../../assets/skills/demo"
assert_file_eq "$PROJ/.claude/skills/demo/SKILL.md" "DEMOBODY"
assert_symlink "$PROJ/.out/sub" "../assets/skills/demo"
assert_file_eq "$PROJ/.out/sub/SKILL.md" "DEMOBODY"

e2e_step "再 apply は冪等（warning 無し・readlink 不変）"
REAPPLY_ERR="$E2E_WORK/reapply.err"
layat apply docs 2>"$REAPPLY_ERR"
if [ -z "$(layat_warnings "$REAPPLY_ERR")" ]; then
	e2e_pass "再 apply で warning が出ない"
else
	e2e_fail "再 apply で warning が出た: $(layat_warnings "$REAPPLY_ERR")"
fi
assert_symlink "$PROJ/.claude/skills/demo" "../../assets/skills/demo"
assert_symlink "$PROJ/.out/sub" "../assets/skills/demo"

e2e_step "--root 上書き: 上書き先に src が無ければ上書き先のパスを名指しして停止し、何も置かない"
ALT="$E2E_WORK/alt"
mkdir -p "$ALT"
ALT_ERR="$E2E_WORK/alt.err"
alt_code=0
layat apply docs --root "$ALT" 2>"$ALT_ERR" || alt_code=$?
if [ "$alt_code" -ne 0 ]; then
	e2e_pass "exit $alt_code: 上書き先に src が無い apply は失敗する"
else
	e2e_fail "exit 0 (期待 非 0): 上書き先に src が無いのに apply が成功した"
fi
if grep -qF "root-relative link source does not exist" "$ALT_ERR" && grep -qF "$ALT/assets/skills/demo" "$ALT_ERR"; then
	e2e_pass "エラーが上書き先の src パスを名指しする"
else
	e2e_fail "エラーが上書き先の src パスを名指ししない: $(cat "$ALT_ERR")"
fi
assert_absent "$ALT/.claude/skills/demo"
assert_absent "$ALT/.out/sub"

e2e_step "--root 上書き: 上書き先に src があれば上書き先の下に相対リンクで張られ、元の root は変わらない"
mkdir -p "$ALT/assets/skills/demo"
echo "ALTBODY" >"$ALT/assets/skills/demo/SKILL.md"
layat apply docs --root "$ALT"
assert_symlink "$ALT/.claude/skills/demo" "../../assets/skills/demo"
assert_file_eq "$ALT/.claude/skills/demo/SKILL.md" "ALTBODY"
assert_symlink "$ALT/.out/sub" "../assets/skills/demo"
assert_file_eq "$ALT/.out/sub/SKILL.md" "ALTBODY"
assert_symlink "$PROJ/.claude/skills/demo" "../../assets/skills/demo"
assert_file_eq "$PROJ/.claude/skills/demo/SKILL.md" "DEMOBODY"

e2e_step "リポジトリを移動しても readlink は不変で内容へ到達できる"
MOVED="$E2E_WORK/moved"
cd "$E2E_WORK"
mv "$PROJ" "$MOVED"
cd "$MOVED"
assert_symlink "$MOVED/.claude/skills/demo" "../../assets/skills/demo"
assert_file_eq "$MOVED/.claude/skills/demo/SKILL.md" "DEMOBODY"
assert_symlink "$MOVED/.out/sub" "../assets/skills/demo"
assert_file_eq "$MOVED/.out/sub/SKILL.md" "DEMOBODY"

e2e_step "移動先での再 apply は別 root の初回 apply（foreign の last-wins と記録なしの warning だけ）"
MOVED_ERR="$E2E_WORK/moved.err"
moved_code=0
layat apply docs 2>"$MOVED_ERR" || moved_code=$?
if [ "$moved_code" -eq 0 ]; then
	e2e_pass "exit 0: 移動先での再 apply"
else
	e2e_fail "exit $moved_code (期待 0): 移動先での再 apply: $(cat "$MOVED_ERR")"
fi
GOT_WARN="$(layat_warnings "$MOVED_ERR" | sort)"
WANT_WARN="$(printf '%s\n' \
	"layat: no last-applied record; starting from an empty baseline" \
	"layat: overwriting an unrecorded symlink (foreign; last-wins): .claude/skills/demo" \
	"layat: overwriting an unrecorded symlink (foreign; last-wins): .out/sub" | sort)"
if [ "$GOT_WARN" = "$WANT_WARN" ]; then
	e2e_pass "warning は foreign の last-wins 置換 2 件と記録なし 1 件だけ"
else
	e2e_fail "warning が期待と不一致: got=[$GOT_WARN] want=[$WANT_WARN]"
fi
assert_symlink "$MOVED/.claude/skills/demo" "../../assets/skills/demo"
assert_symlink "$MOVED/.out/sub" "../assets/skills/demo"

e2e_step "entry を外して apply すると stale 除去される"
PROJ="$MOVED"
write_flake demo
git -c user.email=e2e@layat.test -c user.name=e2e add -A
git -c user.email=e2e@layat.test -c user.name=e2e commit -qm "drop sub"
layat apply docs
assert_symlink "$MOVED/.claude/skills/demo" "../../assets/skills/demo"
assert_file_eq "$MOVED/.claude/skills/demo/SKILL.md" "DEMOBODY"
assert_absent "$MOVED/.out/sub"
assert_file_eq "$MOVED/assets/skills/demo/SKILL.md" "DEMOBODY"

e2e_finish
