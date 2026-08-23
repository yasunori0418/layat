#!/usr/bin/env bash
# prune: 実 nix で apply した project mode の系列から root（git toplevel）を消し、`nput prune` が
# その <roothash> 系列だけを丸ごと削除することをアサートする（→ DSG-096dc893-21f4-45e3-9347-986e9275b4d1・
# ADR-0034）。engine のユニットは tmpdir に手で組んだ fixture で判定分岐を覆っているので、ここが見るのは
# 「実 apply が作った本物のレイアウト（.pending / 世代リンク / .root）を prune が拾って消し切れるか」と
# 「消えた結果 store が GC 可能になるか」の 2 点。
#
# GC 可能になったことは `nix-store --gc --print-roots` に当該系列の `.pending` / `profile-N-link` が
# 出なくなることで見る（実 GC は回さない。`<state>` 配下の out-link と世代リンクは
# /nix/var/nix/gcroots/auto 経由の indirect root で、系列を消すまで store path を保持し続ける）。
#
# 走査基底は 2 つあるが（ユーザー state 基底と system 基底 → ADR-0036 §3）、system 基底
# /nix/var/nix/profiles/nput は実機の共有状態なので触らない。検証は e2e_isolate が切った
# XDG_STATE_HOME 側の系列だけで行う。
set -euo pipefail
source "$(dirname "$0")/../lib.sh"
e2e_isolate

BASE="$XDG_STATE_HOME/nix/profiles/nput"

# 1 つの project を書き出して apply する。root（git toplevel）ごとに独立した <roothash> 系列が
# 立つので、消す側と残す側を別ディレクトリで作る。
make_project() { # $1: プロジェクト dir 名, $2: entry の target
	local proj="$E2E_WORK/$1"
	mkdir -p "$proj/srcrepo/d"
	echo "$1" >"$proj/srcrepo/d/file"
	cat >"$proj/flake.nix" <<EOF
{
$(e2e_flake_inputs)
  outputs = { self, nixpkgs, nput }: {
    nput = nixpkgs.lib.genAttrs $E2E_SYSTEMS (system: {
      docs = nput.lib.mkManifest {
        pkgs = nixpkgs.legacyPackages.\${system};
        root = nput.lib.projectRoot;
        entries."$2" = { src = ./srcrepo; subpath = "d"; };
      };
    });
  };
}
EOF
	(
		cd "$proj"
		git init -q
		git -c user.email=e2e@nput.test -c user.name=e2e add -A
		git -c user.email=e2e@nput.test -c user.name=e2e commit -qm init
		nput apply docs
	)
}

# base 直下の系列のうち、.root が $1 を指すものの <roothash> ディレクトリを返す。
# roothash の計算式（→ paths.RootHash）をシナリオ側へ写し取ると実装と二重管理になるので、
# 実際に置かれた backref の内容から引く。
series_dir_of() { # $1: root の絶対パス
	local d
	for d in "$BASE"/*/; do
		[ -f "$d/.root" ] || continue
		if [ "$(cat "$d/.root")" = "$1" ]; then
			printf '%s' "${d%/}"
			return 0
		fi
	done
	return 1
}

# print-roots に当該系列配下の gcroot（.pending / profile-N-link）が何件残っているか。
# print-roots の行頭はリンクパスを二重引用符で囲んだ形（`"<path>" -> <store path>`）なので、
# 系列 dir を接頭辞に持つ行だけを数える。パスに正規表現メタ文字（mktemp の tmpdir が持つ `.`）が
# 入るため -F の固定文字列で照合する（行頭固定は落ちるが、`"` 込みの接頭辞は他の行と衝突しない）。
gcroot_count_under() { # $1: 系列ディレクトリ
	nix-store --gc --print-roots 2>/dev/null | grep -cF "\"$1/" || true
}

e2e_step "2 つの project を apply（一方を孤児にし、他方は生存させる）"
make_project gone ".nput-out/gone"
make_project live ".nput-out/live"

GONE_ROOT="$E2E_WORK/gone"
LIVE_ROOT="$E2E_WORK/live"
GONE_SERIES="$(series_dir_of "$GONE_ROOT")" || { e2e_fail "孤児にする系列の backref が見つからない"; e2e_finish; }
LIVE_SERIES="$(series_dir_of "$LIVE_ROOT")" || { e2e_fail "生存させる系列の backref が見つからない"; e2e_finish; }
e2e_log "gone=$GONE_SERIES live=$LIVE_SERIES"

e2e_step "apply が作ったレイアウト（backref・profile・世代リンク）"
assert_exists "$GONE_SERIES/.root"
assert_symlink "$GONE_SERIES/docs/profile"
assert_exists "$GONE_SERIES/docs/profile-1-link"
assert_exists "$LIVE_SERIES/.root"

e2e_step "世代リンクが indirect gcroot として store を保持している"
if [ "$(gcroot_count_under "$GONE_SERIES")" -ge 1 ]; then
	e2e_pass "print-roots に孤児予定系列の gcroot がある（prune 前）"
else
	e2e_fail "prune 前は系列配下の gcroot が print-roots に出るべき"
fi

e2e_step "root（git toplevel）を消して孤児にする"
rm -rf "$GONE_ROOT"
assert_absent "$GONE_ROOT"

e2e_step "prune --dryrun: 孤児系列を挙げるが FS は変えない（→ ADR-0034 §2）"
DRYRUN_OUT="$E2E_WORK/prune-dryrun.out"
nput prune --dryrun >"$DRYRUN_OUT"
cat "$DRYRUN_OUT"
if grep -q "^remove-series	.*	$GONE_ROOT	docs$" "$DRYRUN_OUT"; then
	e2e_pass "dryrun が孤児系列を root パス付きで挙げる"
else
	e2e_fail "dryrun の行が孤児系列を挙げていない: $(cat "$DRYRUN_OUT")"
fi
if grep -q "	$LIVE_ROOT	" "$DRYRUN_OUT"; then
	e2e_fail "dryrun が生存系列を対象に挙げてはいけない"
else
	e2e_pass "dryrun は生存系列を挙げない"
fi
assert_exists "$GONE_SERIES/.root"
assert_exists "$GONE_SERIES/docs/profile-1-link"

e2e_step "prune --dryrun --json: 同じ在庫をエンベロープの info が運ぶ（→ issue #134）"
ENV_DRYRUN="$E2E_WORK/prune-dryrun.json"
run_json 0 "$ENV_DRYRUN" prune --dryrun
assert_json "$ENV_DRYRUN" "results は空（prune は config を名指ししない）" \
	'.results == [] and .dryRun == true and .status == "success"'
assert_json "$ENV_DRYRUN" "removed に孤児系列が root / names 付きで載る" \
	"[.info.removed[] | select(.root == \"$GONE_ROOT\")] | (length == 1) and (.[0].names == [\"docs\"])"
assert_json "$ENV_DRYRUN" "生存系列は removed に載らない" \
	"[.info.removed[] | select(.root == \"$LIVE_ROOT\")] | length == 0"

e2e_step "prune --yes: 孤児系列が丸ごと消え、生存系列は残る"
nput prune --yes -v
assert_absent "$GONE_SERIES"
assert_exists "$LIVE_SERIES/.root"
assert_symlink "$LIVE_SERIES/docs/profile"

e2e_step "生存 root の配置物は prune で変わらない"
assert_symlink "$LIVE_ROOT/.nput-out/live"
assert_file_eq "$LIVE_ROOT/.nput-out/live/file" "live"

e2e_step "GC 回収可能になった（print-roots に系列配下の gcroot が出ない）"
if [ "$(gcroot_count_under "$GONE_SERIES")" -eq 0 ]; then
	e2e_pass "print-roots から孤児系列の gcroot が消えた"
else
	e2e_fail "prune 後も gcroot が残っている: $(nix-store --gc --print-roots 2>/dev/null | grep -F "\"$GONE_SERIES/")"
fi
if [ "$(gcroot_count_under "$LIVE_SERIES")" -ge 1 ]; then
	e2e_pass "生存系列の gcroot は保持されている"
else
	e2e_fail "生存系列の gcroot まで消えてはいけない"
fi

e2e_step "再実行は no-op（孤児が無ければ何も消さない）"
nput prune --yes
assert_exists "$LIVE_SERIES/.root"

e2e_finish
