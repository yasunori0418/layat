---
id: "CASE-eec02695-ac60-4ea9-a825-f23f7669d132"
type: test_case
name: "nixgen_test.go — entrypoint 探索順序と installable 引数の 2 形式"
status: stable
resource: "internal/generator/nixgen/nixgen_test.go"
covers:
  - "TC-69537269-32d7-4e04-8722-806b58e7db7a"
  - "TC-b6d6cf76-16f9-4c9f-9b50-74104254ad6e"
  - "TC-4a39ca86-e866-41ec-93d2-663f8e6c1ce0"
follows:
  - "TP-e7c25263-6d2d-4a37-8275-26906889d912"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-15T06:08:07+09:00"
---
# nixgen_test.go — entrypoint 探索順序と installable 引数の 2 形式

## 対象

`internal/generator/nixgen/nixgen_test.go`（nix 生成器。旧 `cmd/layat/nix_test.go` から移設）

## 検証内容

- **探索順序** — flake の entrypoint を最優先とし、次に legacy の候補を順に見る順序が、
  ディレクトリを明示したときとカレントディレクトリから自動で探すときの双方で同一であること
- **明示指定のバリエーション** — ファイルを直接指定したときに名前から形が決まり、flake 形は
  ディレクトリを、legacy 形はファイルを参照先に持つこと。認識できないファイル名・存在しない
  パス・対象ファイルの無いディレクトリがいずれもエラーになること
- **installable 引数の 2 形式** — flake 形は entrypoint と attr path を 1 つの installable と
  して組み、attr path にシステムの次元を含むこと。legacy 形はファイル指定と attr path を
  分けて渡し、システムの次元を含まないこと。名前空間の引数とログ用のラベルも同様に対比される
  こと
- **発見した entrypoint の公開** — 発見前は空で、flake 形はディレクトリ、legacy 形はファイルの絶対パスを
  返すこと

- **失敗の構造** — 発見の失敗が Stage discover で原因（`fs.ErrNotExist` 等）を保持すること。
  `-f` のパスを権限で stat できない失敗が `fs.ErrPermission` を保ち、不在と報告されないこと。
  CWD に entrypoint が無い失敗は CWD のパスを含み、`fs.ErrNotExist` を包まないこと。nix eval と
  読み取り専用の build の失敗がそれぞれ Stage roots / build になり、失敗したサブコマンドと
  installable の全体を括弧で示す 1 行の Message と、キャプチャした stderr を持つこと。ロック内の
  build の失敗も Stage build で、nix の出力が生成器の writer へ流れつつ stderr がキャプチャされる
  こと。nix が成功した後の失敗（一括 eval の結果を解析できない・out-link を読めない・
  `--print-out-paths` が空）も `Failed` の `generator.Error` で、それぞれ Stage roots・build・build
  になること
- **失敗の分類と案内** — `error:` の行の、experimental feature が無効だという 1 種類の文言だけが
  検出され（`error:` 以外の行と他の文言は検出されない）、`PrerequisiteMissing` の `Guidance` に
  nix.conf と `NIX_CONFIG` の両手段が入ること。eval で問い合わせた attr path が `error:` の行の
  引用符の中にあるとき（flake の候補の列挙と legacy の selection path の 2 通りの文言）だけ
  `NotFound` で、config 名・名前空間が無い旨の Message と、単発は config 名の確認、一括は
  config の定義を促す `Guidance` を持つこと。config の中の評価エラー
  （`error: attribute 'foo' missing`）、build での属性の不在、それ以外は `Failed` になること。writer への書き込みは best-effort で、
  writer が失敗しても nix の成功は成功のまま、失敗時の `Stderr` には後続の出力まで全て残ること
- **診断の素通し** — 成功時にも eval と build のどちらの処理でも nix の stderr が writer へ届くこと。
  nix が行を分けて書いても writer へは行単位で渡り、終端の無い最後の断片は終了時に 1 回で
  渡ること
- **`--debug` の開示** — 実行する nix コマンドの引数が生成器に渡した開示用の関数へ渡り、writer へは
  何も書かれないこと

探索と引数の組み立ては一時ディレクトリに空ファイルを置いた実 FS の fixture に対して評価し、
nix のプロセスは起動しない。失敗の構造と `--debug` の開示は PATH の先頭に置いた nix のスタブで
確かめる。
