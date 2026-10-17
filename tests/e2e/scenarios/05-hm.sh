#!/usr/bin/env bash
# HM module: home-manager standalone configuration を非 NixOS で評価・activate し、activation
# （home.activation.layat）が複数 config を 1 回の apply --all で配置し profile ごとに世代を持つこと、
# 前段衝突では何も配置せず、部分失敗では他 profile の配置と世代コミットを残して非ゼロになることをアサート。
set -euo pipefail
source "$(dirname "$0")/../lib.sh"
e2e_isolate

# standalone HM の activate は configured username == \$USER を要求するため実ユーザー名を使う。
USERNAME="$(id -un)"
PROFILES="$XDG_STATE_HOME/nix/profiles/layat"

PROJ="$E2E_WORK/cfg"
mkdir -p "$PROJ/srcrepo/skills/nix" "$PROJ/srcrepo/skills/work"
echo "SKILLBODY" >"$PROJ/srcrepo/skills/nix/SKILL.md"
echo "WORKBODY" >"$PROJ/srcrepo/skills/work/SKILL.md"

# multi: layat.entries（configs.default の糖衣）と別名 config。collide: 2 config が同一 target。
# partial: 辞書順で後の p2 の target を foreign な実ファイルで塞いで conflict させる。
cat >"$PROJ/flake.nix" <<EOF
{
$(e2e_flake_inputs with-hm)
  outputs = { self, nixpkgs, home-manager, layat }:
    let
      system = "$E2E_SYSTEM";
      pkgs = nixpkgs.legacyPackages.\${system};
      nix = { src = ./srcrepo; subpath = "skills/nix"; };
      work = { src = ./srcrepo; subpath = "skills/work"; };
      mkHome = layatCfg: home-manager.lib.homeManagerConfiguration {
        inherit pkgs;
        modules = [
          layat.homeManagerModules.default
          {
            home.username = "$USERNAME";
            home.homeDirectory = "$HOME";
            home.stateVersion = "24.05";
            home.enableNixpkgsReleaseCheck = false;
            layat = { enable = true; } // layatCfg;
          }
        ];
      };
    in {
      homeConfigurations = {
        multi = mkHome {
          entries.".cfg/skill" = nix;
          configs.work.entries.".work/skill" = work;
        };
        collide = mkHome {
          configs.c1.entries.".collide/skill" = nix;
          configs.c2.entries.".collide/skill" = work;
        };
        partial = mkHome {
          configs.p1.entries.".p1/skill" = nix;
          configs.p2.entries.".p2/skill" = work;
        };
      };
    };
}
EOF

cd "$PROJ"
git init -q
git -c user.email=e2e@layat.test -c user.name=e2e add -A
git -c user.email=e2e@layat.test -c user.name=e2e commit -qm init

# $1: homeConfigurations の名前。activationPackage の store path を返す（eval / build が通ること自体も検証）。
build_act() { nix build ".#homeConfigurations.$1.activationPackage" --no-link --print-out-paths; }

# $1: profile 名。profile が世代 1 だけを持ち current がそれを指すこと。
assert_single_generation() {
	local dir="$PROFILES/$1"
	local gens; gens="$(find "$dir" -maxdepth 1 -name 'profile-*-link' 2>/dev/null | wc -l)"
	if [ "$(readlink "$dir/profile" 2>/dev/null)" = "profile-1-link" ] && [ "$gens" -eq 1 ]; then
		e2e_pass "profile $1 が独立した世代 1 を持つ"
	else
		e2e_fail "profile $1 の世代が不正: current=$(readlink "$dir/profile" 2>/dev/null) gens=$gens"
	fi
}

# $1: activationPackage, $2: ログ保存先, $3...: ログに載るべき断片。activate が非ゼロで終わること。
assert_activate_fails() {
	local act="$1" log="$2" needle
	shift 2
	if "$act/activate" >"$log" 2>&1; then
		e2e_fail "activate は非ゼロで終わるべき: $(cat "$log")"
	else
		e2e_pass "activate が非ゼロで終わる"
	fi
	for needle in "$@"; do
		if grep -qF -- "$needle" "$log"; then
			e2e_pass "activate の出力が「$needle」を含む"
		else
			e2e_fail "activate の出力に「$needle」が無い: $(cat "$log")"
		fi
	done
}

e2e_step "multi: activationPackage をビルドして activate（2 config を 1 回の apply --all で配置）"
ACT="$(build_act multi)"
e2e_log "activationPackage: $ACT"
"$ACT/activate"

e2e_step "multi: 両 config が \$HOME 配下に store symlink として配置されたか"
assert_symlink "$HOME/.cfg/skill"
assert_file_eq "$HOME/.cfg/skill/SKILL.md" "SKILLBODY"
assert_symlink "$HOME/.work/skill"
assert_file_eq "$HOME/.work/skill/SKILL.md" "WORKBODY"
case "$(readlink "$HOME/.cfg/skill")" in
	/nix/store/*) e2e_pass "store symlink を指す" ;;
	*) e2e_fail "store symlink を指すべき: $(readlink "$HOME/.cfg/skill")" ;;
esac

e2e_step "multi: default と work が別々の profile に独立した世代を持つ"
assert_single_generation default
assert_single_generation work
if [ "$(readlink -f "$PROFILES/default/profile")" != "$(readlink -f "$PROFILES/work/profile")" ]; then
	e2e_pass "2 profile の世代は別の store path を指す"
else
	e2e_fail "2 profile の世代が同じ store path を指す: $(readlink -f "$PROFILES/default/profile")"
fi

e2e_step "collide: 同一 target の configuration は eval を通り、activate が非ゼロで何も配置しない"
ACT_COLLIDE="$(build_act collide)"
assert_activate_fails "$ACT_COLLIDE" "$E2E_WORK/collide.log" "claimed by both c1 and c2" ".collide/skill"
assert_absent "$HOME/.collide"
assert_absent "$PROFILES/c1"
assert_absent "$PROFILES/c2"

e2e_step "partial: 後段 p2 の conflict で activate は非ゼロ、p1 の配置と世代コミットは完了する"
ACT_PARTIAL="$(build_act partial)"
mkdir -p "$HOME/.p2"
echo "FOREIGN" >"$HOME/.p2/skill"
assert_activate_fails "$ACT_PARTIAL" "$E2E_WORK/partial.log" "conflict(s) detected" ".p2/skill"
assert_symlink "$HOME/.p1/skill"
assert_file_eq "$HOME/.p1/skill/SKILL.md" "SKILLBODY"
assert_single_generation p1
assert_file_eq "$HOME/.p2/skill" "FOREIGN"
if [ -z "$(find "$PROFILES/p2" -maxdepth 1 -name 'profile*' 2>/dev/null)" ]; then
	e2e_pass "conflict した p2 は世代をコミットしない"
else
	e2e_fail "p2 に世代がある: $(ls -la "$PROFILES/p2")"
fi

e2e_step "エラーでも --json は適合エンベロープ + exit 1（HM fixture は layat 出力を持たない）"
# この flake は homeConfigurations だけを公開し layat 出力を持たないため、どの名前でも
# rootKind 先取り eval（layat.<system>.<name>.rootKind）が失敗する。名前 nosuch は任意で、
# 検証対象は「subject 確定後の eval 失敗が results[0].errors[]（主体起因の層）に載る」こと。
ENV_ERR="$E2E_WORK/error.json"
run_json 1 "$ENV_ERR" list-generations nosuch
assert_json "$ENV_ERR" "status=error・subject=nosuch の一様形" \
	'.status == "error" and .results[0].subject.name == "nosuch"'
assert_json "$ENV_ERR" "エラーは subject errors[] に構造化・items=[]" \
	'(.results[0].errors | length) >= 1 and .results[0].result.items == []'

e2e_step "生成器の失敗では nix 自身の診断が stderr へ届き、--json の message にも載る"
ERR_LOG="$E2E_WORK/error.stderr"
layat list-generations nosuch >/dev/null 2>"$ERR_LOG" || true
if grep -q "does not provide attribute" "$ERR_LOG"; then
	e2e_pass "stderr に nix の診断（does not provide attribute）が届く"
else
	e2e_fail "stderr に nix の診断が無い: $(cat "$ERR_LOG")"
fi
assert_json "$ENV_ERR" "errors[0].message に nix の診断（does not provide attribute）を含む" \
	'.results[0].errors[0].message | contains("does not provide attribute")'

e2e_finish
