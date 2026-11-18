---
id: "ADR-af18b71e-0a67-47d6-8382-0e3e00987533"
type: adr
name: "暗黙の入力は配置を決める値だけを、暗黙に解決したときだけ info に記録する"
description: "--json の実行で暗黙に解決した entrypoint・generator・root だけを info にキーとして記録し、明示引数で与えた値と、状態の置き場・実行環境の値は記録しない"
section: cli-output
status: stable
sources:
  - id: "issue-827"
    resource: "https://github.com/yasunori0418/layat/issues/827"
    title: "#827 暗黙の入力の記録"
  - id: "outturn-adr-0028"
    resource: "https://github.com/yasunori0418/outturn/blob/628a4704dbbe/docs/adr/0028-working-directory-required-and-implicit-input.md"
    title: "yasunori0418/outturn ADR-0028"
origin: "outturn の ADR-0028 で、暗黙の入力の解決結果を info に記録することが MUST になったことへの追従（Issue #827、2026-11-18）"
justifies:
  - "REQ-50af28ba-ddf0-4488-abd7-244f1a5597cb"
  - "REQ-a5053191-1c6a-449b-9c5e-5ff49dc5aead"
references:
  - "ADR-6255a1ed-8df4-4230-b77d-cb9303a03e71"
  - "ADR-c6bc7060-7af4-4889-9ce6-5800601c9dcc"
  - "ADR-d15a216f-b5d6-4df9-9825-b23a341bec1d"
  - "ADR-d0949b61-59b7-4a96-8a47-998438ee872a"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-18T23:46:39+09:00"
---
# 暗黙の入力は配置を決める値だけを、暗黙に解決したときだけ info に記録する

- ステータス: stable
- 日付: 2026-11-18
- 関連: [--json は outturn 規約に常時準拠する](outturn-conformance-principle.md), [読み取り専用の列挙は info で返す](readonly-enumeration-as-info.md), [templates と layat init は固定 ref で導入する](templates-and-init-fixed-ref.md), [workingDirectory はコマンドの処理の前に観測し、取得できない --json の実行はエンベロープを出さない](working-directory-observed-before-run.md), GitHub Issue #827, yasunori0418/outturn ADR-0028
- 起点: outturn の ADR-0028 で、暗黙の入力の解決結果を info に記録することが MUST になったことへの追従（Issue #827、2026-11-18）

## 背景

outturn の spec §8 の `standalone` は、暗黙の入力を使った実行で、その解決結果を `info` に記録することを
求める。暗黙の入力とは、明示引数で与えられなかった入力を、ツールが作業ディレクトリ・環境変数・既定の
パスの設定ファイルから自分で解決したものである。

layat は `-f` を省くと作業ディレクトリから entrypoint を探す。`--generator` を省くと、環境変数・設定
ファイル・既定値から generator を決める。`--root` を省くと、root の種類に応じて git の toplevel や
`$HOME` を root にする。このほかにも、state dir・system の基底・PATH から探すコマンドを環境から決める。
どの値を、どの条件で、どこに記録するかを決める必要があった。

## 決定

記録する値は、何を配置するかを決める `entrypoint`・`generator`・`root` の 3 種類とする。state dir と
system の基底は状態の置き場、PATH から探すコマンドは実行環境として扱い、記録しない。init のテンプレートの
ref は固定の定数で、入力に数えない。

暗黙に解決した値だけを、`info` のキーとして出す。明示引数で与えた値のキーは出さない。出すキーが 1 つも
無い実行では、`info` そのものを出さない。

- `entrypoint` と `generator` はトップの `info` に置く。`--all` でも 1 度だけ出す。
- `root` は主体ごとの `result.info` に置く。

`generator` は、`--generator` を省いた実行で、採用した generator の名前を記録する。既定の `nix` に
なったときも記録する。環境変数・設定ファイル・既定値のどれで決まったかという由来は記録しない。

`entrypoint` は、`-f` を省いて探索で見つかった実行だけで記録する。値は `-f` に渡せば同じ entrypoint に
なる形にする。flake では `flake.nix` のあるディレクトリの絶対パス、legacy ではファイルの絶対パスである。
`--manifest` を渡した実行は entrypoint を探さず generator も選ばないので、どちらも記録しない。

`root` は、`--root` を省いた実行で、root の種類が project か home のときだけ記録する。system の root は
`/` に、fixed の root は manifest の値に決まり、環境から解決しない。home の root を明示する手段は
`--root` である。

## 根拠

entrypoint・generator・root が決まれば、どの config をどこへ配置するかが決まる。この 3 つを記録すれば、
消費側は記録した値を明示引数に渡して、同じ配置を再現できる。state dir や PATH は、同じ配置を別の場所に
記録するか、別の実行ファイルで行うかを変えるだけで、配置の中身を変えない。

暗黙に解決したときだけ記録するのは、明示した値は消費側が既に知っているからである。キーがあれば
暗黙に解決した値、無ければ明示した値と読み分けられる。

## 棄却した代替案

### 明示か暗黙かによらず常に出す

キーの有無から、値が暗黙に解決されたかどうかを読み取れなくなる。明示した値を出し直すだけのキーも増える。

### state dir まで記録する

state dir は状態の置き場で、配置の中身を変えない。記録すると、system の基底や PATH のコマンドなど、
環境から決まる値をどこまで記録するかの線引きが無くなる。

### `entrypoint` と `generator` だけを記録する

project の root は作業ディレクトリの git の toplevel から決まり、同じ entrypoint でも作業ディレクトリに
よって配置先が変わる。home の root も `$HOME` から決まる。root を記録しないと、配置先を決めた暗黙の
入力が残らない。
