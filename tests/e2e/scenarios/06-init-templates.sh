#!/usr/bin/env bash
# init + templates: `layat init <t>` でテンプレを展開し、展開後 flake が `nix flake check` を通ることを検証する。
# layat input は `--override-input layat path:$REPO_ROOT`、展開元は LAYAT_TEMPLATE_REF で局所リポジトリへ向ける。
set -euo pipefail
source "$(dirname "$0")/../lib.sh"
e2e_isolate

# unknown flake output 'layat' warning は exit 0（許容）。それ以外の評価エラーは FAIL。
for t in standalone project; do
	e2e_step "layat init ${t}（LAYAT_TEMPLATE_REF=path:\$REPO_ROOT で展開）"
	d="$E2E_WORK/init-${t}"
	mkdir -p "$d"
	(
		cd "$d"
		LAYAT_TEMPLATE_REF="path:$REPO_ROOT" layat init "$t"
	)

	assert_exists "$d/flake.nix"
	if [ "$t" = "project" ]; then
		# project テンプレは .gitignore 雛形を同梱する。
		assert_exists "$d/.gitignore"
	fi

	e2e_step "展開した ${t} テンプレが nix flake check を通る（layat を局所リポジトリへ override）"
	if nix flake check "$d" --override-input layat "path:$REPO_ROOT"; then
		e2e_pass "nix flake check 通過: ${t}"
	else
		e2e_fail "nix flake check が失敗: ${t}"
	fi
done

e2e_step "init --json: results:[] + トップレベル info"
d="$E2E_WORK/init-json"
mkdir -p "$d"
cd "$d"
ENV_INIT="$E2E_WORK/init.json"
LAYAT_TEMPLATE_REF="path:$REPO_ROOT" run_json 0 "$ENV_INIT" init standalone
assert_json "$ENV_INIT" "results は空（init は主体を持たない）" \
	'.results == [] and .status == "success" and .dryRun == false'
assert_json "$ENV_INIT" "info に展開テンプレート名と ref が載る" \
	'.info.template == "standalone" and (.info.ref | startswith("path:"))'
assert_exists "$d/flake.nix"

e2e_finish
