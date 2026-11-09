---
id: "ADR-10dcf5a9-f298-4123-9a50-24f43e25d30c"
type: adr
name: "コアを store パスを root 相対の target へ置くプリミティブにする"
status: stable
origin: "ADR-0004 §a を書き直した"
justifies:
  - "REQ-37b56673-6e40-4a1b-a2a7-5d3c084e3e66"
references:
  - "ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed"
  - "ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71"
  - "ADR-eeeb9dc8-9089-43ec-946a-c3ed76e98137"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:38:08+09:00"
---
# コアを store パスを root 相対の target へ置くプリミティブにする

- ステータス: stable
- 日付: 2026-11-03
- 関連: [root は 3 つの marker で表す](three-root-markers.md), [--json は outturn 規約に常時準拠する](outturn-conformance-principle.md), [layat.lib.mkEnv で buildEnv 合成環境を提供する](mkenv-buildenv.md), `docs/concept.md`
- 起点: ADR-0004 §a を書き直した

## 背景

layat には north-star となる構想がある。nixpkgs のパッケージ群（store パス）を活かしつつ、配置だけを
利用者に操作させて、Arch や Gentoo のようなミニマル Linux ディストリビューションの基盤を作ることである。
これを踏まえて、コアの中心となる抽象を決める必要があった。

## 決定

layat のコアは「nix store のパスを root 相対の target に置く」プリミティブとする。`$HOME` への
配置（`home.file` 相当）は、root を `$HOME` にしたときの一適用にすぎない。

root が取りうる値は [root は 3 マーカーと絶対パス文字列の union を取る](../requirements/root-marker-kinds.md) に、root の表し方は
[root は 3 つの marker で表す](three-root-markers.md) に置く。

## 根拠

root を `$HOME` に固定すると、distro の構想へ進むときにコアを作り直すことになる。root を
一般化しておけば、プロジェクトや `/` への配置にも引数を変えるだけで対応する。

## 棄却した代替案

### `$HOME` への配置を中心の抽象にし、distro は将来別に設計する

当面の実装は単純になるが、system 層へ広げるときに root の抽象を後から足すリファクタが要る。
