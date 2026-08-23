#!/usr/bin/env bash
# prune: 実 nix で apply した project mode の系列から root（git toplevel）を消し、`nput prune` が
# その <roothash> 系列だけを丸ごと削除することをアサートする（→ DSG-096dc893-21f4-45e3-9347-986e9275b4d1・
# ADR-0034）。engine のユニットは tmpdir に手で組んだ fixture で判定分岐を覆っているので、ここが見るのは
# 「実 apply が置いた本物のレイアウト（実 nix-env --set が張る世代リンクと indirect gcroot）を
# prune が拾って消し切れるか」と「消えた結果 store が GC 可能になるか」の 2 点。
#
# `.pending` は逆にユニット側だけが持つ形で、ここには現れない（apply 成功時に cleanupPending が
# 消す → internal/engine/engine.go 手順 10・ADR-0011。世代リンクが gcroot を継承する）。
#
# GC 可能になったことは `nix-store --gc --print-roots` に当該系列の `profile-N-link` が出なくなる
# ことで見る（実 GC は回さない。`<state>` 配下の世代リンクは /nix/var/nix/gcroots/auto 経由の
# indirect root で、系列を消すまで store path を保持し続ける）。
#
# 走査基底は 2 つある（ユーザー state 基底と system 基底 → ADR-0036 §3）。system 基底は絶対パス
# なので $HOME / XDG_STATE_HOME の差し替えでは動かせず、`e2e_isolate` が
# NPUT_SYSTEM_PROFILE_BASE で隔離先へ向けている（→ cmd/nput/prune.go）。これにより破壊的な
# `prune --yes` も実機の /nix/var/nix/profiles/nput には触れない。下でその隔離自体も確認する。
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

# print-roots を 1 回読んでキャッシュへ落とす。失敗（daemon 不在・権限）は空出力と区別が付かず、
# 「gcroot が消えた」の陰性判定が偽陽性で通ってしまうので、ここで落とす。
PRINT_ROOTS_CACHE="$E2E_WORK/print-roots.out"
refresh_print_roots() {
	if ! nix-store --gc --print-roots >"$PRINT_ROOTS_CACHE" 2>/dev/null; then
		e2e_fail "nix-store --gc --print-roots が失敗した（gcroot の判定ができない）"
		return 1
	fi
	if [ ! -s "$PRINT_ROOTS_CACHE" ]; then
		e2e_fail "print-roots が空（この環境には最低でも nput 自身の gcroot がある前提）"
		return 1
	fi
	return 0
}

# キャッシュした print-roots に当該系列配下の gcroot（世代リンク profile-N-link）が何件あるか。
# print-roots の各行はリンクパスを二重引用符で囲んだ形（`"<path>" -> <store path>`）なので、
# 系列 dir を接頭辞に持つ行だけを数える。パスに正規表現メタ文字（mktemp の tmpdir が持つ `.`）が
# 入るため -F の固定文字列で照合する（行頭固定は落ちるが、`"` 込みの接頭辞は他の行と衝突しない）。
gcroot_count_under() { # $1: 系列ディレクトリ
	grep -cF "\"$1/" "$PRINT_ROOTS_CACHE" || true
}

e2e_step "2 つの project を apply（一方を孤児にし、他方は生存させる）"
make_project gone ".nput-out/gone"
make_project live ".nput-out/live"

GONE_ROOT="$E2E_WORK/gone"
LIVE_ROOT="$E2E_WORK/live"
GONE_SERIES="$(series_dir_of "$GONE_ROOT")" || { e2e_fail "孤児にする系列の backref が見つからない"; exit 1; }
LIVE_SERIES="$(series_dir_of "$LIVE_ROOT")" || { e2e_fail "生存させる系列の backref が見つからない"; exit 1; }
e2e_log "gone=$GONE_SERIES live=$LIVE_SERIES"

e2e_step "system 基底が隔離先へ向いている（実機の共有状態を触らない）"
if [ "$NPUT_SYSTEM_PROFILE_BASE" = "$E2E_WORK/system" ]; then
	e2e_pass "NPUT_SYSTEM_PROFILE_BASE=$NPUT_SYSTEM_PROFILE_BASE"
else
	e2e_fail "system 基底が隔離されていない: '${NPUT_SYSTEM_PROFILE_BASE:-未設定}'"
fi

e2e_step "apply が作ったレイアウト（backref・profile・世代リンク）"
assert_exists "$GONE_SERIES/.root"
assert_symlink "$GONE_SERIES/docs/profile"
assert_symlink "$GONE_SERIES/docs/profile-1-link"
assert_exists "$LIVE_SERIES/.root"

e2e_step "世代リンクが indirect gcroot として store を保持している"
refresh_print_roots
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
# 行の形は `remove-series\t<roothash>\t<root>\t<names>`（→ cmd/nput/prune.go の printPrunePlan）。
# roothash のフィールドは [^<TAB>]* で押さえる（`.*` はタブを越えて貪欲に食い、フィールド構造を
# 検証しない形になる）。
if grep -q "^remove-series	[^	]*	$GONE_ROOT	docs$" "$DRYRUN_OUT"; then
	e2e_pass "dryrun が孤児系列を root パス付きで挙げる"
else
	e2e_fail "dryrun の行が孤児系列を挙げていない: $(cat "$DRYRUN_OUT")"
fi
if grep -q "^remove-series	[^	]*	$LIVE_ROOT	" "$DRYRUN_OUT"; then
	e2e_fail "dryrun が生存系列を対象に挙げてはいけない"
else
	e2e_pass "dryrun は生存系列を挙げない"
fi
assert_exists "$GONE_SERIES/.root"
assert_symlink "$GONE_SERIES/docs/profile-1-link"

e2e_step "prune --dryrun --json: 同じ在庫をエンベロープの info が運ぶ（→ issue #134）"
ENV_DRYRUN="$E2E_WORK/prune-dryrun.json"
run_json 0 "$ENV_DRYRUN" prune --dryrun
assert_json "$ENV_DRYRUN" "results は空（prune は config を名指ししない）" \
	'.results == [] and .dryRun == true and .status == "success"'
assert_json "$ENV_DRYRUN" "removed に孤児系列が root / names 付きで載る" \
	"[.info.removed[] | select(.root == \"$GONE_ROOT\")] | (length == 1) and (.[0].names == [\"docs\"])"
assert_json "$ENV_DRYRUN" "生存系列は removed に載らない" \
	"[.info.removed[] | select(.root == \"$LIVE_ROOT\")] | length == 0"
# 隔離側に想定外の skip が無いこと。skipped を見ないと、系列が backref 読み取り失敗などで
# skip 側へ落ちても removed に孤児が居る限り通ってしまう。
assert_json "$ENV_DRYRUN" "隔離した state 基底の系列は 1 件も skip されていない" \
	"[.info.skipped[] | select(.series.dir | startswith(\"$BASE/\"))] | length == 0"

e2e_step "prune --yes: 孤児系列が丸ごと消え、生存系列は残る"
nput prune --yes -v
assert_absent "$GONE_SERIES"
assert_exists "$LIVE_SERIES/.root"
assert_symlink "$LIVE_SERIES/docs/profile"

e2e_step "生存 root の配置物は prune で変わらない"
assert_symlink "$LIVE_ROOT/.nput-out/live"
assert_file_eq "$LIVE_ROOT/.nput-out/live/file" "live"

e2e_step "GC 回収可能になった（print-roots に系列配下の gcroot が出ない）"
# refresh_print_roots が print-roots の取得失敗・空出力を fail にするので、下の -eq 0 は
# 「root として数えられなくなった」ことだけを意味する（コマンドが落ちた結果の 0 と混ざらない）。
refresh_print_roots
if [ "$(gcroot_count_under "$LIVE_SERIES")" -ge 1 ]; then
	e2e_pass "生存系列の gcroot は保持されている"
else
	e2e_fail "生存系列の gcroot まで消えてはいけない"
fi
if [ "$(gcroot_count_under "$GONE_SERIES")" -eq 0 ]; then
	e2e_pass "print-roots から孤児系列の gcroot が消えた"
else
	e2e_fail "prune 後も gcroot が残っている: $(grep -F "\"$GONE_SERIES/" "$PRINT_ROOTS_CACHE")"
fi

e2e_step "再実行は no-op（孤児が無ければ何も消さない）"
RERUN_ERR="$E2E_WORK/prune-rerun.err"
nput prune --yes -v 2>"$RERUN_ERR"
if grep -q 'no-op' "$RERUN_ERR"; then
	e2e_pass "再実行が no-op を報告する"
else
	e2e_fail "再実行は何も消さず no-op を報告すべき: $(cat "$RERUN_ERR")"
fi
# 生存系列は .root だけでなく <name> 側まで残る（.root しか見ないと <name> を消す実装が素通りする）。
assert_exists "$LIVE_SERIES/.root"
assert_symlink "$LIVE_SERIES/docs/profile"

e2e_finish
