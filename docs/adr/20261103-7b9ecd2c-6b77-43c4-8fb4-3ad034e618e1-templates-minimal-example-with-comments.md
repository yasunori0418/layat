---
id: "ADR-7b9ecd2c-6b77-43c4-8fb4-3ad034e618e1"
type: adr
name: "template に置く動く example を 1 config に絞る"
status: 採用
origin: "ADR-0018 §1 を書き直した"
justifies:
  - "REQ-196ddabf-6569-4303-942e-050872972501"
references:
  - "ADR-baf6346d-783b-4607-ab1c-c2c93c09e064"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-7b9ecd2c-6b77-43c4-8fb4-3ad034e618e1: template に置く動く example を 1 config に絞る

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-baf6346d-783b-4607-ab1c-c2c93c09e064
- 起点: ADR-0018 §1 を書き直した

## 背景

`layat init <template>` は `nix flake init -t` の薄いラッパーで、`templates/standalone` と
`templates/project` を展開する。template に動く例をいくつ置くか、`subpath`・`method = "copy"`・
`mkOutOfStoreSymlink`・複数 entry といった書き方のバリエーションをどう見せるかは決まっていなかった。

## 決定

各 template は動く example を 1 config だけ置き、書き方のバリエーションはコメントで示す。
`standalone` は homeRoot の例を 1 つ、`project` は projectRoot の例を 1 つと devShell を置く。

各 template のファイル構成とコメントで示す範囲は REQ-196ddabf-6569-4303-942e-050872972501 に置く。

## 根拠

template は利用者が書き換えて使う出発点で、動く例が多いほど不要な分を削る手間が増える。layat は
ファイルを生成せず、ファイルを展開するのは nix の templates 機構である。template は書き方を学ぶための
最小形に留めるのが筋が通る。

コメントは評価に影響しないので、例を増やさずに書き方を示せる。動く例が 1 つなら、展開した直後に
そのまま `layat apply` が成功する。

## 棄却した代替案

### 複数のパターンを動く形で同梱する

展開してすぐ使える例は増えるが、starter が肥大し、利用者は使わない entry を削ることから始める
ことになる。バリエーションはコメントで足りる。

### ほぼ空のプレースホルダにする

最小にはなるが、初めて使う利用者が manifest の書き方を掴めず、`layat init` を使う意味が薄れる。
