#!/usr/bin/env bash
# 上流より古い item と、人の確認が無い high の risk を列挙する。
#
#   sara-fresh [--json]
#
# exit code は 0 = 列挙なし / 1 = 列挙あり / 2 = sara check の失敗・JSON の形の異常・frontmatter を読めない item。

set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage: sara-fresh [--json]

sara ドキュメントグラフの item を generated / verified で確かめ、2 節で出す:
  1. 再検証待ち: refines / derives_from / satisfies / threatens / mitigates / covers / follows で
     指している上流の generated.at が、自身の基準時刻より新しい item
  2. level が high の risk のうち、generated.at 以降の human: の verified を持たないもの

基準時刻は generated.at と、generated.at 以降の verified[].at のうち最も新しい時刻。
status が deprecated の item は上流・下流のどちらとしても扱わない。

exit code:
  0  どの節も空
  1  いずれかの節が空でない
  2  sara check の失敗・JSON の形の異常・frontmatter を読めない item
EOF
}

json_out=0
case "${1-}" in
  -h | --help)
    usage
    exit 0
    ;;
  --json) json_out=1 ;;
  "") ;;
  *)
    usage
    exit 2
    ;;
esac
if [ "$#" -gt 1 ]; then
  usage
  exit 2
fi

# sara.toml のあるリポジトリルートで sara を実行する。
# SARA_FRESH_ROOT はテストが一時ディレクトリのグラフを指すための seam。
repo_root=${SARA_FRESH_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || printf '.')}
cd "$repo_root" || {
  echo "sara-fresh: リポジトリルートへ移動できない: $repo_root" >&2
  exit 2
}

# sara 呼び出しの seam（テストが JSON の形の異常を再現するため）。
sara_cmd=${SARA_FRESH_SARA:-sara}

# グラフが invalid の間は列挙しない（壊れたグラフから作る一覧は信用できないため）。
if ! graph=$("$sara_cmd" check --format json); then
  echo "sara-fresh: sara check が失敗した（先にグラフを valid にすること）" >&2
  exit 2
fi

if ! printf '%s' "$graph" | jq -e '(.items | type) == "array"' >/dev/null; then
  echo "sara-fresh: sara check の JSON に .items 配列が無い（sara の出力形状が変わった可能性）" >&2
  exit 2
fi
if ! printf '%s' "$graph" | jq -e '.valid == true' >/dev/null; then
  echo "sara-fresh: sara check が invalid を報告した（先にグラフを valid にすること）" >&2
  exit 2
fi

# generated と verified は sara の fields に出ないため、各 item の frontmatter から読む。
# yq の --front-matter は先頭のファイルにしか効かないので、frontmatter を awk で
# multi-document の 1 本にまとめ、ファイル名を __fresh_file キーで添えて yq を 1 回だけ呼ぶ。
mapfile -t files < <(printf '%s' "$graph" | jq -r '.items[].file')
frontmatter=''
if [ "${#files[@]}" -gt 0 ]; then
  if ! frontmatter=$(awk '
      function quote(s) { gsub(/\047/, "\047\047", s); return "\047" s "\047" }
      FNR == 1 {
        print "---"
        print "__fresh_file: " quote(FILENAME)
        if ($0 != "---") nextfile
        next
      }
      $0 == "---" { nextfile }
      { print }
    ' "${files[@]}" | yq -o=json -I=0 '{"file": .__fresh_file, "generated": .generated, "verified": .verified}'); then
    echo "sara-fresh: frontmatter を YAML として読めない item がある:" >&2
    for f in "${files[@]}"; do
      yq --front-matter=extract '.' "$f" >/dev/null 2>&1 || echo "  $f" >&2
    done
    exit 2
  fi
fi

# frontmatter は item 数に比例して大きく、引数長の上限を超えるため、ファイルとして渡す。
if ! result=$(printf '%s' "$graph" | jq --slurpfile fm <(printf '%s\n' "$frontmatter") '
  # UTC オフセット付きの ISO 8601 を epoch 秒へ直す。読めない値は null。
  def epoch:
    if type != "string" then null
    else (capture("^(?<d>[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2})(?<f>\\.[0-9]+)?(?<z>Z|(?<s>[+-])(?<h>[0-9]{2}):(?<m>[0-9]{2}))$")
          | (try ((.d + "Z") | fromdateiso8601) catch null) as $t
          | if $t == null then null
            else $t + ("0" + (.f // "") | tonumber)
              - (if .z == "Z" then 0
                 else ((.h | tonumber) * 3600 + (.m | tonumber) * 60) * (if .s == "+" then 1 else -1 end) end)
            end) // null
    end;
  def upstream_relations: ["refines", "derives_from", "satisfies", "threatens", "mitigates", "covers", "follows"];

  ($fm | map({key: .file, value: .}) | from_entries) as $by_file
  | [.items[]
     | select(.fields.status? != "deprecated")
     | ($by_file[.file] // {}) as $f
     | . + {
         gen_at: $f.generated.at?,
         gen_ep: ($f.generated.at? | epoch),
         verified_raw: $f.verified
       }
     | . + {verified: (if (.verified_raw | type) == "array"
                       then [.verified_raw[] | {by: (.by? // ""), at: .at?, ep: (.at? | epoch)}]
                       else null end)}] as $items
  | [$items[]
     | (if .gen_ep == null then "\(.file): generated.at を時刻として読めない（実際: \(.gen_at // "無し")）" else empty end),
       (if .verified_raw != null and .verified == null then "\(.file): verified がリストでない" else empty end),
       (.file as $file | .verified // [] | .[] | select(.ep == null) | "\($file): verified の at を時刻として読めない（実際: \(.at // "無し")）")
    ] as $errors
  | if ($errors | length) > 0 then {errors: $errors}
    else
      ($items | map(.gen_ep as $g | . + {base_ep: ([$g] + [(.verified // [])[] | .ep | select(. >= $g)]) | max})
       | map({key: .id, value: .}) | from_entries) as $by_id
      | def row: {id, name, file: (.file | ltrimstr("./"))};
      {
        stale: ([$by_id[]
          | . as $down
          | [upstream_relations[] as $rel
             | ($down.relations[$rel]? // [])[]
             | $by_id[.] // empty
             | select(.gen_ep > $down.base_ep)
             | {id, relation: $rel, at: .gen_at}] | sort_by(.id, .relation)
          | select(length > 0) as $ups
          | $down | row + {upstreams: $ups}] | sort_by(.id)),
        unverified_high_risk: ([$by_id[]
          | select(.type == "risk" and .fields.level? == "high")
          | .gen_ep as $g
          | select([(.verified // [])[] | select((.by | startswith("human:")) and .ep >= $g)] | length == 0)
          | row] | sort_by(.id))
      }
    end'); then
  echo "sara-fresh: sara check の JSON と frontmatter を突き合わせられない" >&2
  exit 2
fi

if printf '%s' "$result" | jq -e 'has("errors")' >/dev/null; then
  echo "sara-fresh: frontmatter を読めない item がある:" >&2
  printf '%s' "$result" | jq -r '.errors[] | "  " + .' >&2
  exit 2
fi

if [ "$json_out" -eq 1 ]; then
  printf '%s' "$result" | jq '.'
else
  printf '%s' "$result" | jq -r '
    def section(title; rows):
      ["## " + title]
      + (if (rows | length) == 0 then ["なし"] else rows end);
    section("再検証待ち（上流の generated.at が基準時刻より新しい item）";
      [.stale[] | "\(.id)\t\(.name)\t\(.file)\t\([.upstreams[].id] | unique | join(","))"])
    + [""]
    + section("human: の verified が無い high の risk";
      [.unverified_high_risk[] | "\(.id)\t\(.name)\t\(.file)"])
    | .[]'
fi

if printf '%s' "$result" | jq -e '(.stale | length) + (.unverified_high_risk | length) > 0' >/dev/null; then
  exit 1
fi
