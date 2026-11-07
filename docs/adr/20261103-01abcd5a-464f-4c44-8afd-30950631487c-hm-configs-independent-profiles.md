---
id: "ADR-01abcd5a-464f-4c44-8afd-30950631487c"
type: adr
name: "HM モジュールは layat.configs.<name> ごとに独立した profile を持つ"
status: stable
sources:
  - id: "issue-585"
    resource: "https://github.com/yasunori0418/layat/issues/585"
    title: "#585"
origin: "ADR-0035 §1 を書き直した"
justifies:
  - "REQ-c6891aeb-13c0-4ae7-9ad1-5c343735266a"
  - "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:35:23+09:00"
---
# ADR-01abcd5a-464f-4c44-8afd-30950631487c: HM モジュールは layat.configs.<name> ごとに独立した profile を持つ

- ステータス: stable
- 日付: 2026-11-03
- 関連: GitHub Issue #585, `docs/spec.md`
- 起点: ADR-0035 §1 を書き直した

## 背景

CLI を単体で使う場合は entrypoint の `layat.<name>` ごとに profile を分け、役割ごとに独立して世代を積める。
HM モジュールが 1 つの entries（= 1 つの profile）しか持たないと、CLI 単体で複数の profile を
使い分けている利用者は、HM へ移るときに同じ構成を表せない。

## 決定

モジュールの共通オプションに `layat.configs.<name>` を設け、`<name>` を profile 名とする。config ごとに
`mkManifest` を呼んで独立した link-farm を作り、1 config = 1 manifest = 1 profile とする。複数の
config を 1 つの manifest へ合成するヘルパーは提供しない。

オプションの型は REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10、config と profile の対応は
REQ-c6891aeb-13c0-4ae7-9ad1-5c343735266a に置く。

## 根拠

- `<name>` は CLI 単体で使う `layat.<name>` と同じ「config 名」の次元で、home mode の profile dir の鍵として
  そのまま使われる。profile dir のレイアウトは変わらない
- 世代の単位を profile に保つ。合成すると 1 つの世代に複数の役割が含まれ、片方だけを戻せなくなる
- 名前を `configs` にするのは、spec の用語（config = named manifest）と揃えるためである

## 棄却した代替案

### 1 つの entries のまま留める

役割ごとの profile の分離が CLI 単体での利用だけのものになり、HM へ移る利用者の障壁になる。

### `layat.profiles.<name>` と名付ける

世代管理の内部語である profile dir と紛れる。

### 複数の config を 1 つの manifest へ合成するヘルパーを提供する

1 config = 1 manifest = 1 profile の対応が崩れ、世代の単位が曖昧になる。
