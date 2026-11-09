---
id: "ADR-3878ea60-4f4f-48a3-a389-3b7fbb672897"
type: adr
name: "--generator を prune と init では無視する"
status: stable
sources:
  - id: "issue-395"
    resource: "https://github.com/yasunori0418/layat/issues/395"
    title: "#395"
origin: "Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0056 §6 を書き直した"
justifies:
  - "REQ-badc7e10-0ba4-40d4-b334-6e40193119db"
references:
  - "ADR-36cb1e78-90c0-47d2-bbb1-4bfec5db2964"
  - "ADR-87a2cf34-9592-4c81-8980-246bca579ca0"
  - "ADR-ef9e4099-0cf4-409a-9e73-518866aea676"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# --generator を prune と init では無視する

- ステータス: stable
- 日付: 2026-11-03
- 関連: [生成器は明示の指定だけで選ぶ](generator-chosen-explicitly-only.md), [manifest の取得をプロセス内の Go インターフェイスの生成器越しにする](manifest-generator-in-process-interface.md), [Go の整形は gofmt、静的解析は go vet と golangci-lint で行う](gofmt-vet-golangci.md), GitHub Issue #395
- 起点: Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0056 §6 を書き直した

## 背景

生成器は `--generator` フラグ・環境変数・設定ファイルから選ぶ（→ [生成器は明示の指定だけで選ぶ](generator-chosen-explicitly-only.md)）。
サブコマンドの中には manifest を得ないものがある。prune は世代の系列だけを扱い、init は生成器を
経由しない（→ [manifest の取得をプロセス内の Go インターフェイスの生成器越しにする](manifest-generator-in-process-interface.md)）。status も manifest を得ずに動く。

## 決定

`--generator` は全サブコマンド共通のフラグとし、prune と init では無視する。値を検証せず、環境変数と
設定ファイルも読まない。manifest を得ないサブコマンドは同じ扱いにする（status も該当する）。

対象のサブコマンドの規範は [生成器は明示指定のみで選び、フラグ・環境変数・設定ファイル・既定の順に解決する](../requirements/generator-selection-precedence.md) に置く。

## 根拠

生成器を使わない操作が、未知の値や不正な設定ファイルで止まる理由は無い。共通のフラグ `-f` も同じ
扱いで、どのサブコマンドの起動行にも同じフラグを付けられる。

## 棄却した代替案

### 全サブコマンドで値と設定ファイルを検証する

不正な設定ファイルが置かれたディレクトリで、生成器と無関係な prune や init まで `E_INPUT` で止まる。

### prune と init では `--generator` を受け付けない

サブコマンドごとにフラグの定義が分かれ、共通の起動行をサブコマンドだけ変えて使えなくなる。
