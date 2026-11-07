#!/usr/bin/env bash
# OKF 互換の入れ子キー（generated / verified / sources）と stale_after の形を frontmatter から確かめる。
#
# 実行:
#   nix develop '.?dir=dev#sara' -c dev/tests/okf-frontmatter.sh   # devShell / CI から直接
#   nix flake check ./dev                                           # checks.okf-frontmatter 経由

# 1 回の実行で全失敗を報告するため -e は使わない。
set -uo pipefail

# 1 にすると generated を持たない item を違反にする。
GENERATED_REQUIRED=0

fail=0
pass() { printf 'ok   - %s\n' "$1"; }
fault() {
  printf 'FAIL - %s\n' "$1"
  fail=1
}

# shellcheck source=dev/scripts/lib-testdoc.sh
. "$(dirname "$0")/../scripts/lib-testdoc.sh"

require_commands "okf-frontmatter.sh（nix develop '.?dir=dev#sara' から実行する）" yq jq awk date || exit 1
require_yq_go okf-frontmatter.sh || exit 1

cd "$(testdoc_repo_root)" || exit 1

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

now=$(date +%s)

# --- 入力の収集 --------------------------------------------------------------

# 先頭行が `---` で、frontmatter に `type:` を持つ .md を item として列挙する。
list_items() {
  find "$1" -name '*.md' -type f -exec awk '
    FNR == 1 { in_fm = ($0 == "---"); if (!in_fm) nextfile; next }
    $0 == "---" { nextfile }
    /^type:/ { print FILENAME; nextfile }
  ' {} + | LC_ALL=C sort
}

# item の frontmatter を、判定関数が読むレコードへ直す。1 行 1 要素で、列は US（\x1f）区切りの
# <ファイル> <キー> <添字> <形> <値 1> <値 2>。空の列を詰めないよう、IFS 空白でない区切りを使う。読めない item は stderr へ名前を出し、1 を返す。
collect_records() {
  local list=$1 out=$2 file json unreadable=0
  : >"$out"
  while IFS= read -r file; do
    if ! json=$(yq --front-matter=extract -o=json -I=0 '.' "$file" 2>/dev/null); then
      echo "$file"
      unreadable=1
      continue
    fi
    jq -r --arg file "$file" '
      def kind: if type == "object" then "object" elif type == "array" then "array" else "scalar" end;
      def str: if type == "string" then . elif . == null then "" else tostring end;
      def actor($key; $i): [$file, $key, $i, kind,
        (if type == "object" then (.by | str) else "" end),
        (if type == "object" then (.at | str) else "" end)];
      (if has("generated") then (.generated | actor("generated"; "-")) else [$file, "generated", "-", "missing", "", ""] end),
      (if has("verified") then
         (.verified | if type == "array" then (to_entries[] | .value | actor("verified"; "\(.key)")) else [$file, "verified", "-", kind, "", ""] end)
       else empty end),
      (if has("stale_after") then [$file, "stale_after", "-", (.stale_after | kind), (.stale_after | str), ""] else empty end),
      (if has("sources") then
         (.sources | if type == "array" then
            (to_entries[] | .key as $i | .value | [$file, "sources", "\($i)", kind,
              (if type == "object" then (.id | str) else "" end),
              (if type == "object" then (.resource | str) else "" end)])
          else [$file, "sources", "-", kind, "", ""] end)
       else empty end)
      | map(gsub("[\n\u001f]"; " ")) | join("\u001f")
    ' <<<"$json" >>"$out"
  done <"$list"
  return "$unreadable"
}

# --- 判定（本体。§自己検証が同じ関数を合成フィクスチャで叩く） ----------------
#
# 各関数は診断行を stdout へ出し、違反があれば 1 を返す。呼び出し側が pass / fault へ振り分ける。

iso_re='^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]+)?(Z|[+-][0-9]{2}:[0-9]{2})$'

# UTC オフセット付きの ISO 8601 で、date が時刻として読める値か。
is_iso_with_offset() {
  [[ "$1" =~ $iso_re ]] && date -d "$1" +%s >/dev/null 2>&1
}

# generated / verified の各要素が {by, at} の形で、by が空でなく at が過去の ISO 8601 か。
judge_actor_time() {
  local records=$1
  local violated=0 file key i shape by at epoch
  while IFS=$'\x1f' read -r file key i shape by at; do
    [[ "$key" == generated || "$key" == verified ]] || continue
    [[ "$shape" == missing ]] && continue
    local where="$file の $key"
    [[ "$i" != - ]] && where="${where}[$i]"
    if [[ "$key" == verified && "$i" == - ]]; then
      echo "$where がリストでない（実際: $shape）"
      violated=1
      continue
    fi
    if [[ "$shape" != object ]]; then
      echo "$where が {by, at} の組でない（実際: $shape）"
      violated=1
      continue
    fi
    if [[ -z "$by" ]]; then
      echo "$where.by が空"
      violated=1
    fi
    if ! is_iso_with_offset "$at"; then
      echo "$where.at が UTC オフセット付きの ISO 8601 でない（実際: ${at:-空}）"
      violated=1
      continue
    fi
    epoch=$(date -d "$at" +%s)
    if [[ "$epoch" -gt "$now" ]]; then
      echo "$where.at が未来の時刻（実際: $at）"
      violated=1
    fi
  done <"$records"
  return "$violated"
}

# generated を持たない item。required が 1 のときだけ違反にする。
judge_generated_required() {
  local records=$1 required=$2
  local violated=0 file key i shape rest
  [[ "$required" -eq 1 ]] || return 0
  while IFS=$'\x1f' read -r file key i shape rest; do
    if [[ "$key" == generated && "$shape" == missing ]]; then
      echo "$file に generated が無い"
      violated=1
    fi
  done <"$records"
  return "$violated"
}

# stale_after が UTC オフセット付きの ISO 8601 か。
judge_stale_after() {
  local records=$1
  local violated=0 file key i shape value rest
  while IFS=$'\x1f' read -r file key i shape value rest; do
    [[ "$key" == stale_after ]] || continue
    if [[ "$shape" != scalar ]] || ! is_iso_with_offset "$value"; then
      echo "$file の stale_after が UTC オフセット付きの ISO 8601 でない（実際: ${value:-$shape}）"
      violated=1
    fi
  done <"$records"
  return "$violated"
}

# sources がリストで、各要素の resource が空でなく、id がファイル内で一意か。
judge_sources() {
  local records=$1
  local violated=0 file key i shape id resource dups
  while IFS=$'\x1f' read -r file key i shape id resource; do
    [[ "$key" == sources ]] || continue
    if [[ "$i" == - ]]; then
      echo "$file の sources がリストでない（実際: $shape）"
      violated=1
    elif [[ "$shape" != object ]]; then
      echo "$file の sources[$i] が {id, resource, title} の組でない（実際: $shape）"
      violated=1
    elif [[ -z "$resource" ]]; then
      echo "$file の sources[$i].resource が空"
      violated=1
    fi
  done <"$records"
  dups=$(awk -F'\037' '$2 == "sources" && $3 != "-" && $5 != "" { print $1 "\037" $5 }' "$records" |
    LC_ALL=C sort | uniq -d)
  if [[ -n "$dups" ]]; then
    while IFS=$'\x1f' read -r file id; do
      echo "$file の sources に id '$id' が重複している"
    done <<<"$dups"
    violated=1
  fi
  return "$violated"
}

# 判定関数を records へ当て、診断行を fault へ流す。§自己検証がこの関数自体も検証する。
run_judges() {
  local records=$1 required=$2 diagnostics rc=0
  if diagnostics=$(judge_actor_time "$records"); then
    pass "generated / verified の各要素が {by, at} で、at が過去の UTC オフセット付き ISO 8601"
  else
    while IFS= read -r line; do fault "generated / verified: $line"; done <<<"$diagnostics"
    rc=1
  fi
  if diagnostics=$(judge_generated_required "$records" "$required"); then
    if [[ "$required" -eq 1 ]]; then
      pass "全 item が generated を持つ"
    else
      pass "generated の必須判定は無効（GENERATED_REQUIRED=0）"
    fi
  else
    while IFS= read -r line; do fault "generated の必須: $line"; done <<<"$diagnostics"
    rc=1
  fi
  if diagnostics=$(judge_stale_after "$records"); then
    pass "stale_after が UTC オフセット付きの ISO 8601"
  else
    while IFS= read -r line; do fault "stale_after: $line"; done <<<"$diagnostics"
    rc=1
  fi
  if diagnostics=$(judge_sources "$records"); then
    pass "sources の各要素の resource が空でなく、id がファイル内で一意"
  else
    while IFS= read -r line; do fault "sources: $line"; done <<<"$diagnostics"
    rc=1
  fi
  return "$rc"
}

# --- 1. 実データ ----------------------------------------------------------------

list_items docs >"$work/items"
if [[ ! -s "$work/items" ]]; then
  echo "okf-frontmatter.sh: docs/ に item を 1 件も見つけられなかった" >&2
  exit 1
fi

if unreadable=$(collect_records "$work/items" "$work/records"); then
  pass "全 item の frontmatter を読める（$(wc -l <"$work/items") 件）"
else
  while IFS= read -r file; do
    fault "入力: $file の frontmatter を yq で読めなかった（YAML の構文を確認する）"
  done <<<"$unreadable"
fi

run_judges "$work/records" "$GENERATED_REQUIRED"

# --- 2. 自己検証（判定関数が両側へ倒れるか） -----------------------------------
#
# 合成フィクスチャの item（$work/self/docs/ 配下）を収集から判定まで通し、
# 違反なしで真・違反ありで偽の両側へ倒れることを確かめる。

self=$work/self
mkdir -p "$self/docs"

write_item() {
  local name=$1
  shift
  {
    printf -- '---\nid: "X-%s"\ntype: requirement\n' "$name"
    printf '%s\n' "$@"
    printf -- '---\n# 本文\n'
  } >"$self/docs/$name.md"
}

# 違反のない item。Z と ±hh:mm の両方の表記を含める。
write_item ok \
  'generated:' '  by: "human:someone"' '  at: "2026-01-01T00:00:00+09:00"' \
  'verified:' '  - by: "human:someone"' '    at: 2026-01-02T00:00:00Z' \
  'stale_after: "2099-01-01T00:00:00+00:00"' \
  'sources:' '  - id: "issue-1"' '    resource: "https://example.com/1"' '    title: "a"' \
  '  - id: "issue-2"' '    resource: "https://example.com/2"' \
  '  - resource: "https://example.com/3"'
# generated を持たない item（必須判定の対象）。
write_item no-generated 'name: "x"'
# frontmatter の無い文書と、本文にだけ type: を持つ文書は item として数えない。
printf '# 概要\n\ntype: requirement\n' >"$self/docs/overview.md"

list_items "$self/docs" >"$self/items-all"
if [[ "$(wc -l <"$self/items-all")" -eq 2 ]] && ! grep -q overview.md "$self/items-all"; then
  pass "自己検証: list_items が frontmatter に type: を持つ .md だけを数える"
else
  fault "自己検証: list_items が frontmatter に type: を持つ .md だけを数える（実際: $(tr '\n' ' ' <"$self/items-all")）"
fi

self_fail=0

# 期待どおりに真（違反なし）を返すか。
expect_judge_pass() {
  local label=$1
  shift
  if ! "$@" >/dev/null; then
    fault "自己検証: $label — 違反のないフィクスチャで偽を返した（判定が過剰）"
    self_fail=1
  fi
}

# 期待どおりに偽（違反あり）を返し、診断行を出すか。
expect_judge_fail() {
  local label=$1
  shift
  local diagnostics
  if diagnostics=$("$@"); then
    fault "自己検証: $label — 違反のあるフィクスチャで真を返した（判定が常に真になる退行）"
    self_fail=1
  elif [[ -z "$diagnostics" ]]; then
    fault "自己検証: $label — 偽を返したが診断行が空（失敗の原因が報告されない）"
    self_fail=1
  fi
}

# 1 件の item を収集して records にする。
records_of() {
  local name=$1
  printf '%s\n' "$self/docs/$name.md" >"$self/$name.list"
  collect_records "$self/$name.list" "$self/$name.tsv" >/dev/null
  printf '%s' "$self/$name.tsv"
}

ok=$(records_of ok)
for judge in judge_actor_time judge_stale_after judge_sources; do
  expect_judge_pass "$judge" "$judge" "$ok"
done
expect_judge_pass "judge_generated_required（generated あり）" judge_generated_required "$ok" 1
missing=$(records_of no-generated)
expect_judge_pass "judge_generated_required（無効のとき）" judge_generated_required "$missing" 0
expect_judge_fail "judge_generated_required（generated 無し）" judge_generated_required "$missing" 1

# 違反のある item。判定内の分岐ごとに 1 つずつ用意する。
while IFS='|' read -r name judge body; do
  IFS=';' read -r -a lines <<<"$body"
  write_item "$name" "${lines[@]}"
  expect_judge_fail "$judge（$name）" "$judge" "$(records_of "$name")"
done <<'EOF'
gen-scalar|judge_actor_time|generated: "human:someone"
gen-empty-by|judge_actor_time|generated:;  by: "";  at: "2026-01-01T00:00:00+09:00"
gen-no-offset|judge_actor_time|generated:;  by: "human:someone";  at: "2026-01-01T00:00:00"
gen-date-only|judge_actor_time|generated:;  by: "human:someone";  at: "2026-01-01"
gen-future|judge_actor_time|generated:;  by: "human:someone";  at: "2999-01-01T00:00:00+09:00"
ver-not-list|judge_actor_time|verified:;  by: "human:someone";  at: "2026-01-01T00:00:00+09:00"
ver-bad-at|judge_actor_time|verified:;  - by: "human:someone";    at: "yesterday"
stale-no-offset|judge_stale_after|stale_after: "2099-01-01"
stale-list|judge_stale_after|stale_after:;  - "2099-01-01T00:00:00Z"
src-not-list|judge_sources|sources: "issue-1"
src-scalar-element|judge_sources|sources:;  - "https://example.com/1"
src-empty-resource|judge_sources|sources:;  - id: "issue-1";    resource: ""
src-dup-id|judge_sources|sources:;  - id: "issue-1";    resource: "https://example.com/1";  - id: "issue-1";    resource: "https://example.com/2"
EOF

# 読めない frontmatter は収集で名指しされ、判定へ黙って流れない。
write_item broken 'generated: [unclosed'
printf '%s\n' "$self/docs/broken.md" >"$self/broken.list"
if broken_out=$(collect_records "$self/broken.list" "$self/broken.tsv") || [[ "$broken_out" != *broken.md* ]]; then
  fault "自己検証: collect_records — 壊れた frontmatter で真を返したか、ファイルを名指ししない（実際: ${broken_out:-空}）"
  self_fail=1
fi

# run_judges の配線: 判定ごとに、その判定だけが落ちる records で偽を返す（fault・pass は捨てる）。
# 必須判定は no-generated のときだけ有効にし、他の判定の違反と混ぜない。
for case in gen-scalar:0 no-generated:1 stale-no-offset:0 src-dup-id:0; do
  name=${case%:*}
  if (fault() { :; } && pass() { :; } && run_judges "$self/$name.tsv" "${case#*:}"); then
    fault "自己検証: run_judges（$name）— 違反のある records で真を返した（判定が配線されていない）"
    self_fail=1
  fi
done

[[ "$self_fail" -eq 0 ]] && pass "自己検証: 判定関数は違反なしで真・違反ありで偽を返す"

exit "$fail"
