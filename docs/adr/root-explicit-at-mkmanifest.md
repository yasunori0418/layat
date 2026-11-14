---
id: "ADR-f85f1afd-9b87-4f7a-922e-97c62554f198"
type: adr
name: "root の明示を mkManifest 層で必須にする"
section: root-modes
status: stable
sources:
  - id: "issue-2"
    resource: "https://github.com/yasunori0418/layat/issues/2"
    title: "#2"
origin: "Issue #2（root デフォルト）。ADR-0007 §2 の root の明示必須とモジュールによる pin を書き直した"
justifies:
  - "REQ-4ec3accc-8bb6-461f-9024-dcf0027849e4"
  - "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# root の明示を mkManifest 層で必須にする

- ステータス: stable
- 日付: 2026-11-01
- 関連: GitHub Issue #2, `docs/concept.md`, `docs/spec.md`
- 起点: Issue #2（root デフォルト）。ADR-0007 §2 の root の明示必須とモジュールによる pin を書き直した

## 背景

root は配置先の基準で、どれを選ぶかで配置が ephemeral か永続か、rollback を使えるかが変わる。
root に暗黙の既定値（`$HOME` など）を持たせるかどうかと、誰が root を決めるかを決める必要が
あった。

## 決定

root に暗黙の既定値を持たせない。`mkManifest` と CLI の entrypoint の層では `root` の指定を必須に
し、省略は評価時のエラーにする。モジュール（HM は `homeRoot`、NixOS と nix-darwin は
`systemRoot`、devShell は `projectRoot`）は自分の性質に合わせて root を固定し、モジュールの
利用者は root を指定し直さない。規範の細目は [root は明示必須で暗黙デフォルトを持たない](../requirements/root-explicit-required.md) と
[全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](../requirements/module-common-options.md) に置く。

## 根拠

暗黙の既定値は、ephemeral か永続か、rollback があるかという重い違いを、利用者に気付かせない
まま選ばせる。配置を利用者が明示的に決めるという方針にも反する。一方、モジュールは置かれた場所
から root が自然に決まるので、そこで固定すれば利用者の手間は増えない。

## 棄却した代替案

### 暗黙の既定値を残す（`$HOME` または project）

書く量は減るが、モードの重い違いを暗黙に選ばせる。どちらを既定にしても恣意的になる。

### モジュールでも root の指定を必須にする

一貫はするが、モジュールが root を一つに決められるのに、利用者に同じ値を書かせることになる。
