---
id: "ADR-e88f074e-95f4-4e04-97fc-eae2d0c695a3"
type: adr
name: "設定ファイルを strict な TOML で書く"
section: cli-commands
status: stable
sources:
  - id: "issue-395"
    resource: "https://github.com/yasunori0418/layat/issues/395"
    title: "#395"
origin: "Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0056 §4 の形式を書き直した"
justifies:
  - "REQ-badc7e10-0ba4-40d4-b334-6e40193119db"
references:
  - "ADR-36cb1e78-90c0-47d2-bbb1-4bfec5db2964"
  - "ADR-ef9e4099-0cf4-409a-9e73-518866aea676"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# 設定ファイルを strict な TOML で書く

- ステータス: stable
- 日付: 2026-11-03
- 関連: [生成器は明示の指定だけで選ぶ](generator-chosen-explicitly-only.md), [Go の整形は gofmt、静的解析は go vet と golangci-lint で行う](gofmt-vet-golangci.md), GitHub Issue #395
- 起点: Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0056 §4 の形式を書き直した

## 背景

生成器の選択はプロジェクト設定とユーザー設定からも読む（→ [生成器は明示の指定だけで選ぶ](generator-chosen-explicitly-only.md)）。
設定ファイルは利用者が手で書くもので、項目は当面 `generator` だけである。

## 決定

設定ファイルは TOML で書き、未知のキーをエラーにする strict な読み方をする。version の項目は持たない。
存在するが読めない設定ファイル（権限が無い・ディレクトリである）は入力の不正にせず、読み込み失敗の
分類を保つ。

キーと失敗の分類の規範は [生成器は明示指定のみで選び、フラグ・環境変数・設定ファイル・既定の順に解決する](../requirements/generator-selection-precedence.md) に置く。

## 根拠

項目が 1 つの段階で前方互換の仕組みを持っても使われない。未知のキーをエラーにしておけば、将来
項目を足したときに古い layat は無視せずエラーで止まる。読めない設定ファイルは内容の誤り
ではなく環境の問題なので、入力の不正とは分けて報告する。

## 棄却した代替案

### JSON

コメントを書けず、手書きの設定に向かない。

### YAML

暗黙の型変換があり、strict に読んでも値の誤読が残る。

### 未知のキーを無視する

キー名の打ち間違いがエラーにならずに無視され、意図しない生成器で実行される。
