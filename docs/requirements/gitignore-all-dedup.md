---
id: "REQ-1f128917-4424-4e37-8a88-e0bb23a09da7"
type: requirement
name: "gitignore --all は projectRoot の全 config の target をソート + 重複除去して出力する"
status: stable
derives_from:
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  `layat gitignore --all` SHALL output the targets of every projectRoot config, sorted and
  deduplicated, because a repository has a single `.gitignore` and listing them together
  is the natural form.
specification_ja: |
  `layat gitignore --all` は projectRoot の全 config の target をソート + 重複除去して
  出力しなければならない。repo の `.gitignore` は 1 つなので一括列挙が自然であるため。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# REQ-1f128917-4424-4e37-8a88-e0bb23a09da7: gitignore --all は projectRoot の全 config の target をソート + 重複除去して出力する

## 仕様

```bash
layat gitignore --all           # projectRoot の全 config の target をソート + 重複除去して出力
```

`gitignore --all` は projectRoot の全 config の target を **ソート + 重複除去**して
出力する（repo の `.gitignore` は 1 つなので一括列挙が自然）。

出力の書式（アンカー形式）は [gitignore は配置 target を stdout へ列挙するだけでファイルを書き込まない](gitignore-anchored-output.md)、`--json` 時に cross-config dedup をしない
非対称は [--all は config ごとの SubjectResult を単一実行と同一形状で積む](json-all-subject-results.md) の担当。

## 出典

`docs/spec.md`「CLI 仕様」→「サブコマンド体系」の `list-generations --all` /
`gitignore --all` の箇条書き（後段）。

決定の実体は [gitignore --all は project mode の全 config の target を重複なく整列して出す](../adr/gitignore-all-sorted-deduplicated.md)「gitignore --all は project mode の全 config の target を重複なく整列して出す」。
