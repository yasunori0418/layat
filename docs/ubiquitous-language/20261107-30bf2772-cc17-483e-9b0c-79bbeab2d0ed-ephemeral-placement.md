---
id: "UBQ-30bf2772-cc17-483e-9b0c-79bbeab2d0ed"
type: ubiquitous_term
name: "ephemeral placement"
status: stable
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T01:45:24+09:00"
---

# UBQ-30bf2772-cc17-483e-9b0c-79bbeab2d0ed: ephemeral placement

## Definition

The nature of placements in project mode: regenerated on each clone and never committed to the project, so activation never touches git state.

## Avoid

Confusing it with "vendoring" or "placement that commits artifacts".

Placements in project mode are assumed to be untracked. Do not mistake ephemeral placement for "vendoring that commits". The `.gitignore` listing comes from the stdout of a dedicated command (activation does not touch `.gitignore`).

## 定義

project mode における配置物の性質。クローンごとに再生成される前提でプロジェクトにコミットされないので、activation は git 状態に干渉しない。

## 避ける表現

「vendoring」や「成果物をコミットする配置」と混同すること。

project mode の配置物は untracked が前提なので、ephemeral placement を「コミットする vendoring」と取り違えない。`.gitignore` への列挙は専用コマンドの stdout 出力から得る（activation は `.gitignore` に触れない）。
