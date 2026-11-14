---
id: "ADR-4dca6ef4-2791-48d5-aa82-7037630968a6"
type: adr
name: "ephemeral 配置は git に干渉しない"
section: placement
status: stable
origin: "ADR-0005 の決定「ephemeral 配置原則 — 配置物はコミット対象外」を書き直した"
justifies:
  - "REQ-e79178f5-5865-4444-a05d-3ab06f33cd6d"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:41+09:00"
---
# ephemeral 配置は git に干渉しない

- ステータス: stable
- 日付: 2026-11-01
- 関連: `docs/concept.md`, `docs/spec.md`
- 起点: ADR-0005 の決定「ephemeral 配置原則 — 配置物はコミット対象外」を書き直した

## 背景

project mode は git リポジトリの作業ツリーの中へ配置する。配置物をコミットするのか、
`.gitignore` や git の追跡状態に layat が関わるのかを決める必要があった。git で追跡されている
ファイルと layat の古くなった配置の除去がぶつかる可能性もあった。

## 決定

project mode の配置物は ephemeral とし、リポジトリにコミットしない。clone ごとに置き直す。
activation は git の状態に干渉せず、`.gitignore` に触れず、target が git で追跡されているかも
確認しない。規範の細目は [project mode の配置物は ephemeral とし、activation は git 状態に干渉しない](../requirements/project-mode-ephemeral.md) に置く。

## 根拠

配置物が常に追跡されていないことを原則として固定すれば、保守的な古くなった配置の除去が git で追跡された
ファイルとぶつかる事態は原理的に起きない。activation に git の確認を持ち込まずに済み、
activation は置くことだけに専念できる。

## 棄却した代替案

### activation が `.gitignore` へ自動で追記する

利用者の手間は減るが、layat が配置以外のファイルを書き換えるツールになり、設定を生成しない
という方針に反する。

### activation が target の git 追跡状態を確認する

コミット済みのファイルとの衝突を個別に検出できるが、activation が git に依存し、毎回の実行が
重くなる。原則で衝突を消せるので要らない。
