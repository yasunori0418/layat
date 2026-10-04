#!/usr/bin/env bash
# home mode: 仮 $HOME で apply → profile 世代コミット → `layat rollback` で前世代へ復帰 →
# `layat reset` が FS のみを撤去して profile / 世代を動かさないことをアサート。
# 同居させた projectRoot の config を `list-generations --all` が除外することも見る。
set -euo pipefail
source "$(dirname "$0")/../lib.sh"
e2e_isolate

PROJ="$E2E_WORK/cfg"
mkdir -p "$PROJ/srcrepo/a" "$PROJ/srcrepo/b"
echo "AAA" >"$PROJ/srcrepo/a/file"
echo "BBB" >"$PROJ/srcrepo/b/file"

# target / subpath を引数に取り fixture flake を書き出す（世代ごとに entry を入れ替える）。
# 第 2 config の proj は projectRoot で、`--all` の除外検証の直前だけ apply する。
write_flake() {
	local target="$1" sub="$2"
	cat >"$PROJ/flake.nix" <<EOF
{
$(e2e_flake_inputs)
  outputs = { self, nixpkgs, layat }: {
    layat = nixpkgs.lib.genAttrs $E2E_SYSTEMS (system: {
      home = layat.lib.mkManifest {
        pkgs = nixpkgs.legacyPackages.\${system};
        root = layat.lib.homeRoot;
        entries."$target" = { src = ./srcrepo; subpath = "$sub"; };
      };
      proj = layat.lib.mkManifest {
        pkgs = nixpkgs.legacyPackages.\${system};
        root = layat.lib.projectRoot;
        entries.".layat-out/proj" = { src = ./srcrepo; subpath = "$sub"; };
      };
    });
  };
}
EOF
}

# 世代の本数を数える。layat 自体の失敗は非ゼロで返し、0 件はそのままアサートへ渡す。
gens_count() {
	local out
	out="$(layat list-generations home)" || return 1
	printf '%s\n' "$out" | grep -c . || true
}

cd "$PROJ"
write_flake ".cfg/a" "a"
git init -q
git -c user.email=e2e@layat.test -c user.name=e2e add -A
git -c user.email=e2e@layat.test -c user.name=e2e commit -qm gen1

e2e_step "世代 1: apply（entry a）→ \$HOME 配下に配置"
layat apply home
assert_symlink "$HOME/.cfg/a"
assert_file_eq "$HOME/.cfg/a/file" "AAA"

e2e_step "成功時はデフォルト沈黙 / -v で配置レポート"
# 同一世代への再 apply（no-op）。配置レポート行の有無は layat の "完了" マーカーで判定する。
silent_err="$(layat apply home 2>&1 >/dev/null || true)"
if printf '%s' "$silent_err" | grep -q 'layat: apply home done'; then
	e2e_fail "成功時に配置レポートが出てはいけない（既定は沈黙）: '$silent_err'"
else
	e2e_pass "成功時は配置レポート無し（既定沈黙）"
fi
verbose_err="$(layat apply home -v 2>&1 >/dev/null || true)"
if printf '%s' "$verbose_err" | grep -q 'layat: apply home done'; then
	e2e_pass "-v で配置レポートが出る"
else
	e2e_fail "-v で配置レポートが出るべき: '$verbose_err'"
fi

e2e_step "profile 世代がコミットされたか（home mode の profile レイアウト）"
PROFILE="$XDG_STATE_HOME/nix/profiles/layat/home/profile"
assert_symlink "$PROFILE"
GENS="$(layat list-generations home)"
echo "$GENS"
if [ "$(printf '%s\n' "$GENS" | grep -c .)" -ge 1 ]; then
	e2e_pass "list-generations が世代を返す"
else
	e2e_fail "世代が 1 つも無い"
fi

e2e_step "apply --dryrun --json: 既存 profile では generation.before = after"
ENV_DRYRUN="$E2E_WORK/dryrun-gen.json"
run_json 0 "$ENV_DRYRUN" apply home --dryrun
assert_json "$ENV_DRYRUN" "generation が before = after の観測を運ぶ" \
	'.results[0].generation | (.before != null) and (.before == .after)'

e2e_step "世代 2: entry を b に入替えて apply（a は stale 除去）"
write_flake ".cfg/b" "b"
git -c user.email=e2e@layat.test -c user.name=e2e add -A
git -c user.email=e2e@layat.test -c user.name=e2e commit -qm gen2
layat apply home
assert_symlink "$HOME/.cfg/b"
assert_file_eq "$HOME/.cfg/b/file" "BBB"
assert_absent "$HOME/.cfg/a"

e2e_step "2 世代以上あること"
if [ "$(layat list-generations home | grep -c .)" -ge 2 ]; then
	e2e_pass "世代が 2 つ以上ある"
else
	e2e_fail "世代が 2 つ未満"
fi

e2e_step "layat rollback --json で前世代（entry a）へ復帰（RunE → emit の実経路）"
# 前世代へ戻す実行は一度しか成立しないため、--json 側だけで通す。
ENV_ROLLBACK="$E2E_WORK/rollback.json"
run_json 0 "$ENV_ROLLBACK" rollback home
assert_symlink "$HOME/.cfg/a"
assert_file_eq "$HOME/.cfg/a/file" "AAA"
assert_absent "$HOME/.cfg/b"
assert_json "$ENV_ROLLBACK" "generation が 2 → 1 の遷移を運ぶ" \
	'.results[0].generation | .before == 2 and .after == 1'
assert_json "$ENV_ROLLBACK" "items は復帰先 .cfg/a と撤去元 .cfg/b を全在庫として運ぶ" \
	'[.results[0].result.items[] | .info.target] | sort == [".cfg/a", ".cfg/b"]'
assert_json "$ENV_ROLLBACK" "items は全て success（failed / skipped を含まない）" \
	'.results[0].result.items | all(.status == "success")'
assert_json "$ENV_ROLLBACK" "changes は .cfg/a の add と .cfg/b の remove（どちらも可逆）2 件だけ" \
	'.results[0].result as $r
	 | ($r.items | map({key: .id, value: .info.target}) | from_entries) as $t
	 | [$r.changes[] | {kind, reversible, target: $t[.itemId]}]
	 | sort_by(.target) == [{kind: "add", reversible: true, target: ".cfg/a"},
	                        {kind: "remove", reversible: true, target: ".cfg/b"}]'
assert_json "$ENV_ROLLBACK" "status=success・dryRun=false・command=rollback" \
	'.status == "success" and .dryRun == false and .command == "rollback"'

e2e_step "list-generations --json: result.info.generations（items=[]）"
ENV_GENS="$E2E_WORK/list-generations.json"
run_json 0 "$ENV_GENS" list-generations home
assert_json "$ENV_GENS" "info.generations に 2 世代が {number, date, current} で載る" \
	'.results[0].result.info.generations | length == 2 and all(has("number") and has("date") and has("current"))'
assert_json "$ENV_GENS" "current は rollback 先の 1 世代だけ" \
	'[.results[0].result.info.generations[] | select(.current)] | map(.number) == [1]'
assert_json "$ENV_GENS" "items=[]・generation スロット無し・dryRun=false" \
	'.results[0].result.items == [] and (.results[0] | has("generation") | not) and .dryRun == false'

# ここから `list-generations --all`。home の profile が 2 世代を持つこの位置で除外を検証する。
e2e_step "project mode の proj を apply（--all の除外検証に実体を与える）"
# proj の profile は <state>/nix/profiles/layat/<roothash>/proj に置かれる。除外対象を disk 上に実在させる。
layat apply proj
assert_symlink "$PROJ/.layat-out/proj"
# roothash 階層（base/<roothash>/proj）の profile リンクが一意に 1 件だけ当たることを確かめる。
PROJ_CONFIG_PROFILES=("$XDG_STATE_HOME"/nix/profiles/layat/*/proj/profile)
if [ "${#PROJ_CONFIG_PROFILES[@]}" -eq 1 ] && [ -L "${PROJ_CONFIG_PROFILES[0]}" ]; then
	e2e_pass "proj の profile が roothash 階層に一意に実在する: ${PROJ_CONFIG_PROFILES[0]#"$XDG_STATE_HOME/"}"
else
	e2e_fail "proj の profile が roothash 階層に一意に無い（除外検証の前提が崩れる）: ${#PROJ_CONFIG_PROFILES[@]} 件 = ${PROJ_CONFIG_PROFILES[*]}"
fi

e2e_step "list-generations --all --json: home mode の config だけを列挙する"
# 読み取り専用の前後比較はここから始め、両方の出力経路をまたいで挟む。
ALL_PROFILE_BEFORE="$(readlink "$PROFILE")"
ALL_GENS_BEFORE="$(gens_count)"
ENV_GENS_ALL="$E2E_WORK/list-generations-all.json"
run_json 0 "$ENV_GENS_ALL" list-generations --all
# 陽性対照と除外を subject 名の集合の等式で固定する。
assert_json "$ENV_GENS_ALL" "results は home mode の home だけ（proj は roothash 階層なので除外）" \
	'[.results[].subject.name] == ["home"]'
# 以降も添字ではなく名前で引く。
assert_json "$ENV_GENS_ALL" "home の info.generations は名指し列挙と同じ 2 世代（current は世代 1）" \
	'first(.results[] | select(.subject.name == "home")).result.info as $i
	 | ($i.generations | length) == 2
	 and ([$i.generations[] | select(.current) | .number] == [1])'
assert_json "$ENV_GENS_ALL" "home は items=[]・generation スロット無し・status=success" \
	'first(.results[] | select(.subject.name == "home"))
	 | .result.items == [] and (has("generation") | not) and .status == "success"'
assert_json "$ENV_GENS_ALL" "集約 status=success・dryRun=false" \
	'.status == "success" and .dryRun == false'

e2e_step "list-generations --all は読み取り専用（profile も世代も配置も動かない）"
# 読み取り専用を profile リンク先・世代の本数・配置の 3 点で見る。ここで既定（テキスト）経路も通す。
ALL_TEXT_CODE=0
ALL_TEXT="$(layat list-generations --all)" || ALL_TEXT_CODE=$?
if [ "$ALL_TEXT_CODE" -eq 0 ]; then
	e2e_pass "exit 0: layat list-generations --all（テキスト経路）"
else
	e2e_fail "exit $ALL_TEXT_CODE (期待 0): layat list-generations --all"
fi
# テキスト経路の per-config ヘッダ `# <name>` の集合を比較する。
ALL_HEADERS="$(printf '%s\n' "$ALL_TEXT" | grep '^# ' || true)"
if [ "$ALL_HEADERS" = "# home" ]; then
	e2e_pass "テキスト経路のヘッダも home だけ（proj は現れない）"
else
	e2e_fail "テキスト経路のヘッダが '# home' だけでない: '$ALL_HEADERS'（全体: '$ALL_TEXT'）"
fi
# 世代本体の印字を、番号と current マーカーごとリテラルで固定する（gens_count の観測値は使わない）。
ALL_ROWS="$(printf '%s\n' "$ALL_TEXT" | grep '^[0-9]' | cut -f1,3 || true)"
ALL_ROWS_WANT="$(printf '1\t(current)\n2')"
if [ "$ALL_ROWS" = "$ALL_ROWS_WANT" ]; then
	e2e_pass "テキスト経路も世代 1（current）と 2 を出す"
else
	e2e_fail "テキスト経路の世代行が期待と不一致: '$ALL_ROWS'（期待: '$ALL_ROWS_WANT'・全体: '$ALL_TEXT'）"
fi
assert_symlink "$PROFILE" "$ALL_PROFILE_BEFORE"
ALL_GENS_AFTER="$(gens_count)"
if [ "$ALL_GENS_AFTER" -eq "$ALL_GENS_BEFORE" ]; then
	e2e_pass "--all のあとも世代の本数が変わらない（$ALL_GENS_BEFORE 件）"
else
	e2e_fail "--all が世代の本数を動かしてはいけない: $ALL_GENS_BEFORE → $ALL_GENS_AFTER"
fi
assert_symlink "$HOME/.cfg/a"
assert_file_eq "$HOME/.cfg/a/file" "AAA"

e2e_step "layat reset --json --yes: FS のみを撤去し profile / 世代は動かない"
# 撤去対象は profile が指す世代の manifest 由来。rollback で世代 1 に戻った後なので .cfg/a が対象。
PROFILE_BEFORE="$(readlink "$PROFILE")"
GENS_BEFORE="$(gens_count)"
# 前後一致の基準（世代 2 本）を非 JSON 経路でも観測できることを先に固定する。
if [ "$GENS_BEFORE" -eq 2 ]; then
	e2e_pass "reset 前の世代は 2 件"
else
	e2e_fail "reset 前の世代が 2 件でない: $GENS_BEFORE"
fi
ENV_RESET="$E2E_WORK/reset.json"
run_json 0 "$ENV_RESET" reset home --yes
assert_absent "$HOME/.cfg/a"
assert_json "$ENV_RESET" "items は撤去対象の entry（.cfg/a）で status は success" \
	'.results[0].result.items | map({target: .info.target, status}) == [{target: ".cfg/a", status: "success"}]'
assert_json "$ENV_RESET" "changes は .cfg/a の remove（symlink なので可逆）1 件だけ" \
	'.results[0].result as $r
	 | ($r.items | map({key: .id, value: .info.target}) | from_entries) as $t
	 | [$r.changes[] | {kind, reversible, target: $t[.itemId]}]
	   == [{kind: "remove", reversible: true, target: ".cfg/a"}]'
assert_json "$ENV_RESET" "generation スロットを持たない（FS のみの撤去）" \
	'.results[0] | has("generation") | not'
assert_json "$ENV_RESET" "status=success・dryRun=false・command=reset" \
	'.status == "success" and .dryRun == false and .command == "reset"'
# profile の非遷移: リンク先（現行世代）と世代の本数がどちらも reset 前後で変わらない。
assert_symlink "$PROFILE" "$PROFILE_BEFORE"
GENS_AFTER="$(gens_count)"
if [ "$GENS_AFTER" -eq "$GENS_BEFORE" ]; then
	e2e_pass "reset 後も世代の本数が変わらない（$GENS_BEFORE 件）"
else
	e2e_fail "reset が世代の本数を動かしてはいけない: $GENS_BEFORE → $GENS_AFTER"
fi

e2e_finish
