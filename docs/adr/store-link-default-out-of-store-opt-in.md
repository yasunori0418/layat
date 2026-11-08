---
id: "ADR-73c24620-d2c0-4249-97fa-1b6296f1e79c"
type: adr
name: "store link を既定にし、out-of-store は明示関数でのみ選ばせる"
status: stable
origin: "ADR-0001 §a と §b を書き直した"
justifies:
  - "REQ-99ca5381-6c53-426c-b145-7b4297c53868"
  - "REQ-eb363122-385a-499c-a074-c95efb949d07"
references:
  - "ADR-3e283fe2-0177-4caf-8b60-577c5ee3b10a"
  - "ADR-451679d5-473b-4faf-bf2d-a6f6d3e1e131"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:30+09:00"
---
# ADR-73c24620-d2c0-4249-97fa-1b6296f1e79c: store link を既定にし、out-of-store は明示関数でのみ選ばせる

- ステータス: stable
- 日付: 2026-11-03
- 関連: [entry の配置種別のフィールドを method と名付ける](mode-renamed-to-method.md), [src の型で素の文字列を拒否する](srctype-bare-string-rejected.md), `docs/concept.md`, `docs/spec.md`
- 起点: ADR-0001 §a と §b を書き直した

## 背景

初期の設計は、`src` に渡した値の Nix の型で配置の仕方を暗黙に切り替えていて、文字列を渡すと
Nix ストアを経由しない out-of-store symlink（手元のパスを直接指すライブな symlink）になる。

この分岐では、同じ `src` の値が型という見えにくい属性だけでストア由来の配置にもローカル FS への
配置にもなる。配置物はストア由来で再現できるという Nix の前提から外れる配置が、文字列を渡すだけで選ばれていた。
`path` リテラルが暗黙にストアへコピーされる罠とも見分けにくい。

## 決定

`src` の既定の挙動は常に Nix ストアへの symlink とする。out-of-store は lib の明示関数が返す
マーカーを `src` に渡したときにだけ選ばれる。値の型による暗黙の分岐は持たない。`src` が取る形と
マーカーの構築子は [src は path / set / marker の 3 種を取り store link を既定として out-of-store は marker で opt-in する](../requirements/src-kinds-store-default.md) と
[mkOutOfStoreSymlink は out-of-store symlink を表すマーカーを返す](../requirements/mkoutofstoresymlink-marker.md) に置く。

## 根拠

out-of-store のリンクはライブなパスを指す。本質的にストアへの link ではない。配置をストアに
揃えるとは、既定とコアをストアにし、out-of-store を明示した例外に降格することだと解釈した。
型による分岐を消せば、同じフィールドの値が見えない属性で別の配置に変わることは無くなる。
out-of-store を関数で明示的に選ぶ書き味は、home-manager の
`config.lib.file.mkOutOfStoreSymlink` と同じで、利用者のすでに知っている使い方と同じになる。

## 棄却した代替案

### 型による暗黙の分岐を残したまま関数も足す

同じ目的の書き方が 2 つになる。既定をストアに揃える意図と矛盾する。

### out-of-store を廃止する

開発中の dotfiles をライブに編集したいという用途を失う。

### 専用フィールド `outOfStore` を足す

フィールドが増える。`src` との排他を検証する規則も別に要る。マーカーなら `src` 1 つで済む。
