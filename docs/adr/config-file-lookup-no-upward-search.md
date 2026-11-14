---
id: "ADR-5908eb28-40a5-49d5-a09e-15b35082b970"
type: adr
name: "設定ファイルは -f のディレクトリか cwd だけで探す"
section: cli-commands
status: stable
sources:
  - id: "issue-395"
    resource: "https://github.com/yasunori0418/layat/issues/395"
    title: "#395"
origin: "Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0056 §3 を書き直した"
justifies:
  - "REQ-badc7e10-0ba4-40d4-b334-6e40193119db"
references:
  - "ADR-36cb1e78-90c0-47d2-bbb1-4bfec5db2964"
  - "ADR-ef9e4099-0cf4-409a-9e73-518866aea676"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# 設定ファイルは -f のディレクトリか cwd だけで探す

- ステータス: stable
- 日付: 2026-11-03
- 関連: [生成器は明示の指定だけで選ぶ](generator-chosen-explicitly-only.md), [Go の整形は gofmt、静的解析は go vet と golangci-lint で行う](gofmt-vet-golangci.md), GitHub Issue #395
- 起点: Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0056 §3 を書き直した

## 背景

生成器の選択にはプロジェクト設定 `layat.toml` の段がある（→ [生成器は明示の指定だけで選ぶ](generator-chosen-explicitly-only.md)）。
layat はこれまで設定ファイルを読まず、config を設定ファイルから発見する機構も持たない。entrypoint の
発見は生成器の Discover が行うので、設定ファイルを探すために entrypoint を使うと、生成器を決める前に
生成器が要ることになる。

## 決定

プロジェクト設定は、`-f` で指定したディレクトリ（`-f` がファイルならそのファイルのあるディレクトリ）、
`-f` が無ければ cwd だけで探し、親ディレクトリへは遡らない。設定ファイルは生成器の選択にだけ使い、
config の発見には使わない。

`-f` のパスを stat できないときの扱いを含む探索の規範は [生成器は明示指定のみで選び、フラグ・環境変数・設定ファイル・既定の順に解決する](../requirements/generator-selection-precedence.md) に置く。

## 根拠

探す場所を `-f` と cwd に限れば、entrypoint を見つける前に設定を読め、「設定を読むには entrypoint が
要り、entrypoint の発見には生成器が要る」循環が起きない。どの設定が使われたかも cwd の深さに左右され
ない。config の発見に使わないのは、config は entrypoint から評価して得るという既存の方針を保つため
である。

## 棄却した代替案

### 親ディレクトリへ遡って探す

entrypoint の発見との循環を生み、どの設定が使われたかが cwd の深さで変わる。

### 設定ファイルで config も指定できるようにする

config を設定ファイルからも得られるようになり、entrypoint から評価して得る config と食い違いうる。
