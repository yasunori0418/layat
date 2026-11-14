---
id: "ADR-baf6346d-783b-4607-ab1c-c2c93c09e064"
type: adr
name: "--all は config ごとの SubjectResult を積む"
section: cli-output
status: stable
sources:
  - id: "issue-552"
    resource: "https://github.com/yasunori0418/layat/issues/552"
    title: "#552"
  - id: "issue-outturn-1"
    resource: "https://github.com/yasunori0418/outturn/issues/1"
    title: "yasunori0418/outturn#1"
origin: "layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §7 を書き直した"
justifies:
  - "REQ-059eb4d5-63fb-4f8e-b705-11b5e2ed4ae5"
references:
  - "ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71"
  - "ADR-feb60906-1517-46b8-8fa5-e007df4d3e11"
  - "ADR-88231ffc-4836-4252-be34-894944fd28c3"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:26:59+09:00"
---
# --all は config ごとの SubjectResult を積む

- ステータス: stable
- 日付: 2026-11-01
- 関連: [--json は outturn 規約に常時準拠する](outturn-conformance-principle.md), [--json は常に results\[\] エンベロープを出す](results-envelope-always.md), [CLI が一次 UX で devShell に pin 版を同梱する](cli-primary-ux-devshell-pinned.md), GitHub Issue #552, yasunori0418/outturn#1, outturn ADR-0013
- 起点: layat の outturn 準拠化 grilling（2026-07-06）と outturn#1 の方針確定。ADR-0043 §7 を書き直した

## 背景

`apply --all` / `list-generations --all` / `gitignore --all` は、複数の config を 1 回の実行で扱う。
[--json は常に results\[\] エンベロープを出す](results-envelope-always.md)。この決定の上で、複数の
config の結果をどう並べるかを決める必要がある。

## 決定

`--all` は、選んだ config ごとに 1 つの `SubjectResult` を `results[]` に積む。形は単一 config の実行と
同じで、`--all` を区別するフィールドを持たない。トップの `status` は全 subject の集約とする。
`gitignore --all --json` は config 間で重複を除かず、各 subject が自分の paths を持つ。

細目は [--all は config ごとの SubjectResult を単一実行と同一形状で積む](../requirements/json-all-subject-results.md) に置く。

## 根拠

単一実行を N=1 の特別な場合として扱えば、消費側は 1 つの形だけを読めばよい。`gitignore` の JSON は
どの config に由来するかを保つほうが有用で、重複の除去は機械の側で容易にできる。テキストの既定出力は
重複を除いた単一のリストのままで、テキストは集約、JSON は config ごと、という非対称を受け入れる。

## 棄却した代替案

### `gitignore --all --json` でも重複を除いた単一のリストを出す

どの config に由来するかが失われる。
