---
id: "ADR-419914ba-9896-4dd1-853b-fb514fcb9f38"
type: adr
name: "profile は専用ディレクトリレイアウトに置く"
status: stable
origin: "実装前の残セマンティクス確定の第 7 巡（2026-06-14）。ADR-0025 §4 を書き直した"
justifies:
  - "REQ-2aa3abbc-90b2-486e-92de-d785554bdeb3"
references:
  - "ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed"
  - "ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988"
  - "ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-419914ba-9896-4dd1-853b-fb514fcb9f38: profile は専用ディレクトリレイアウトに置く

- ステータス: stable
- 日付: 2026-11-01
- 関連: ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed, ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988, ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75
- 起点: 実装前の残セマンティクス確定の第 7 巡（2026-06-14）。ADR-0025 §4 を書き直した

## 背景

`profileDir` という語が、profile リンクそのものと、それを収めるディレクトリの 2 つの意味で使われていた。
`nix-env --profile P` は P を profile リンクとして扱い、世代を兄弟の `P-N-link` として P の親ディレクトリに
作る。P が store の世代を指すリンクなら、`<profileDir>/.pending` のように P の下へ out-link を置くと
read-only な store の中にリンクを作ろうとして失敗する。

`<roothash>` 階層には backref と複数の config が同居するので、ロックの単位が root 単位なのか config 単位
なのかも曖昧だった。

## 決定

config ごとに専用のディレクトリを与え、これを `profileDir` と呼ぶ。profile リンク・世代のリンク・build の
pending out-link は、このディレクトリの中に兄弟として並べる。pending out-link の名前は `.pending` とする。
backref は `<roothash>` 階層に置き、同じ root の config で共有する。

ディレクトリの中の並びと基底ディレクトリは REQ-2aa3abbc-90b2-486e-92de-d785554bdeb3 に置く。

## 根拠

世代のリンクと pending out-link が profile リンクの兄弟として並ぶので、read-only な store の中にリンクを作ろうとする失敗が
構造的に起きない。profile リンク・世代・pending がどの config のものかは、ディレクトリを見れば分かる。

config ごとにディレクトリが 1 つなので、pending out-link は 1 個で足り、名前に config 名を含める必要もない。
ディレクトリが config の単位と一致するので、ロックの単位もそのディレクトリに揃えられる。

このレイアウトはキーで指す先の物理的な形を決めるもので、キーの決め方は変えない。

## 棄却した代替案

### `profileDir` を profile リンクそのものにする

pending out-link を profile リンクの下に置くことになり、read-only な store の中にリンクを作ろうとして失敗する。

### 共有ディレクトリに config 名を前置した兄弟を並べる

`<name>` / `<name>-N-link` / `.pending-<name>` を 1 つのディレクトリに並べる案。階層は浅くなるが、ロックを
config ごとの別のロックファイルで持つ必要があり、どのリンクがどの config のものかも一目で分からない。
