---
id: "ADR-36cb1e78-90c0-47d2-bbb1-4bfec5db2964"
type: adr
name: "生成器は明示の指定だけで選ぶ"
section: manifest-generator
status: stable
sources:
  - id: "issue-395"
    resource: "https://github.com/yasunori0418/layat/issues/395"
    title: "#395"
origin: "Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0056 §1 を書き直した"
justifies:
  - "REQ-badc7e10-0ba4-40d4-b334-6e40193119db"
references:
  - "ADR-87a2cf34-9592-4c81-8980-246bca579ca0"
  - "ADR-ef9e4099-0cf4-409a-9e73-518866aea676"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:41+09:00"
---
# 生成器は明示の指定だけで選ぶ

- ステータス: stable
- 日付: 2026-11-03
- 関連: [manifest の取得をプロセス内の Go インターフェイスの生成器越しにする](manifest-generator-in-process-interface.md), [Go の整形は gofmt、静的解析は go vet と golangci-lint で行う](gofmt-vet-golangci.md), GitHub Issue #395
- 起点: Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0056 §1 を書き直した

## 背景

manifest の取得は生成器越しに行う（→ [manifest の取得をプロセス内の Go インターフェイスの生成器越しにする](manifest-generator-in-process-interface.md)）。利用者が選べる実生成器は
当面 nix だけだが、選ぶ仕組みが無いと、2 つ目の生成器を足すたびに CLI の使い方を決め直すことになる。

選び方には、各生成器の Discover を順に試して最初に entrypoint を見つけたものを使う暗黙の方式（claim
方式）と、利用者が明示する方式がある。

## 決定

生成器は明示の指定だけで選び、claim 方式は採らない。指定は `--generator` フラグ・環境変数
`LAYAT_GENERATOR`・プロジェクト設定・ユーザー設定・既定の `nix` の順に解決し、上の段で値が決まれば
下の段は読まない。

各段の細目（空の値やキーの無い設定ファイルの扱い、ユーザー設定のパス）は
[生成器は明示指定のみで選び、フラグ・環境変数・設定ファイル・既定の順に解決する](../requirements/generator-selection-precedence.md) に置く。

## 根拠

生成器の切り替えは配置結果を変える操作で、ファイルの置き方から暗黙に決まると原因を追えない。

段を分けるのは置き場所ごとに用途が違うためで、CI や一時的な切り替えは環境変数、リポジトリ単位の
固定はプロジェクト設定、マシン単位の既定はユーザー設定に置ける。フラグが常に勝つので、どの層の
設定も 1 回の実行で上書きできる。

## 棄却した代替案

### claim 方式（Discover を順に試す）

entrypoint の置き方で生成器が切り替わる。同じディレクトリに複数の生成器の entrypoint があると、
結果が試行順に左右される。

### フラグだけで選ぶ

リポジトリやマシン単位で生成器を固定できず、既定以外の生成器を使う利用者は毎回フラグを付ける
ことになる。
