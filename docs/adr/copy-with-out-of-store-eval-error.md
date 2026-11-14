---
id: "ADR-5e46ebd6-e44b-48dc-a93f-a8a3a3903967"
type: adr
name: "copy と out-of-store の併用を eval 時に拒否する"
section: lib-entries
status: stable
origin: "ADR-0013 §4 を書き直した"
justifies:
  - "REQ-16faf428-77f3-492f-b858-222c5274cbf7"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:26:59+09:00"
---
# copy と out-of-store の併用を eval 時に拒否する

- ステータス: stable
- 日付: 2026-11-01
- 関連: `docs/spec.md`
- 起点: ADR-0013 §4 を書き直した

## 背景

[root 相対マーカーを別関数 mkRootRelativeSymlink で作る](mkrootrelativesymlink-separate-function.md)。out-of-store マーカーには `mkOutOfStoreSymlink` と、この `mkRootRelativeSymlink` がある。
entry の `src` にこれらを渡し、`method = "copy"` を指定することは、型の上では書ける。この
組み合わせで engine が何をするかを決める必要があった。

## 決定

`method = "copy"` と out-of-store マーカーの併用は、`normalizeManifest` が評価時に `lib.throwIf`
で拒否する。絶対パスのマーカーにも root 相対のマーカーにも同じく適用する。規範の細目は
[意図が矛盾する組み合わせをクロスフィールドの確認で評価時に拒否する](../requirements/normalize-manifest-cross-field-checks.md) に置く。

## 根拠

out-of-store はリンク先の内容をライブに保つための method で、copy は place-once で
スナップショットを置く method である。両者の意図は矛盾するので、片方の振る舞いを暗黙に選ぶより、
書いた時点で止めたほうが利用者の誤解が残らない。評価時に契約を守らせる方針とも一貫する。

## 棄却した代替案

### warning を出して symlink に降格する

設定は通るが、利用者が指定した `method` と違う振る舞いを暗黙に選ぶことになる。
