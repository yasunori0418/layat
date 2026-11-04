---
id: "ADR-10dcf5a9-f298-4123-9a50-24f43e25d30c"
type: adr
name: "コアを store パスを root 相対の target へ置くプリミティブにする"
status: 採用
origin: "ADR-0004 §a を書き直した"
justifies:
  - "REQ-37b56673-6e40-4a1b-a2a7-5d3c084e3e66"
references:
  - "ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed"
  - "ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71"
  - "ADR-eeeb9dc8-9089-43ec-946a-c3ed76e98137"
---
# ADR-10dcf5a9-f298-4123-9a50-24f43e25d30c: コアを store パスを root 相対の target へ置くプリミティブにする

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed, ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71, ADR-eeeb9dc8-9089-43ec-946a-c3ed76e98137, `docs/concept.md`
- 起点: ADR-0004 §a を書き直した

## 背景

layat には north-star となる構想がある。nixpkgs のパッケージ群（store パス）を活かしつつ、配置だけを
利用者に操作させて、Arch や Gentoo のようなミニマル Linux ディストリビューションの基盤を作ることである。
これを踏まえて、コアの中心となる抽象を決める必要があった。

## 決定

layat のコアは「nix store のパスを root 相対の target に置く」プリミティブとする。`$HOME` への
配置（`home.file` 相当）は、root を `$HOME` にしたときの一適用にすぎない。

root が取りうる値は REQ-37b56673-6e40-4a1b-a2a7-5d3c084e3e66 に、root の表し方は
ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed に置く。

## 根拠

root を `$HOME` に固定すると、distro の構想へ進むときにコアを作り直すことになる。root を
一般化しておけば、プロジェクトや `/` への配置にも引数の違いだけで届く。

## 棄却した代替案

### `$HOME` への配置を中心の抽象にし、distro は将来別に設計する

当面の実装は単純になるが、system 層へ広げるときに root の抽象を後から足すリファクタが要る。
