---
id: "ADR-75d1dcc2-e270-4193-a72c-863a57967baf"
type: adr
name: "project template の .gitignore 手順をコード近くのコメントで示す"
status: 採用
origin: "ADR-0018 §2 を書き直した"
justifies:
  - "REQ-196ddabf-6569-4303-942e-050872972501"
supersedes:
  - "ADR-0018"
references:
  - "ADR-baf6346d-783b-4607-ab1c-c2c93c09e064"
---
# ADR-75d1dcc2-e270-4193-a72c-863a57967baf: project template の .gitignore 手順をコード近くのコメントで示す

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-baf6346d-783b-4607-ab1c-c2c93c09e064
- 起点: ADR-0018 §2 を書き直した

## 背景

project mode の配置物はリポジトリに含めない一時的なもので、利用者はその target を `.gitignore` に
載せる必要がある。登録は一度きりの手作業で、`layat gitignore <name>` がその行を出す。project template
のどこでこの手順を伝えるかは決まっていなかった。

## 決定

project template は、`flake.nix` の manifest 定義のそばに `layat gitignore <name>` の出力を
`.gitignore` へ足すよう促すコメントを置く。あわせて template に `.gitignore` を同梱し、その先頭に
再生成の手順を書いたヘッダコメントを置く。

project template のファイル構成は REQ-196ddabf-6569-4303-942e-050872972501 に置く。

## 根拠

利用者が project template で手を入れるのは `flake.nix` の manifest と `.gitignore` である。手順を
その 2 か所に書けば、target を書き換えたその場で目に入る。`.gitignore` のヘッダは、後から target を
足したときにも再生成のコマンドを思い出させる。

## 棄却した代替案

### 手順を README に書いて同梱する

説明の量は取れるが、template のファイルが増える。README はコードから離れた場所にあり、manifest を
書き換えるときには読まれない。
