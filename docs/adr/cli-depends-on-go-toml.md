---
id: "ADR-aec6bd1d-f776-4d7c-9337-1ed78fee6180"
type: adr
name: "CLI の依存に go-toml/v2 を許可する"
status: stable
sources:
  - id: "issue-395"
    resource: "https://github.com/yasunori0418/layat/issues/395"
    title: "#395"
origin: "Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0056 §4 の依存の許可を書き直した"
justifies:
  - "REQ-637599dc-a1ec-4af5-9e97-e882c7df56d0"
references:
  - "ADR-e88f074e-95f4-4e04-97fc-eae2d0c695a3"
  - "ADR-ef9e4099-0cf4-409a-9e73-518866aea676"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# ADR-aec6bd1d-f776-4d7c-9337-1ed78fee6180: CLI の依存に go-toml/v2 を許可する

- ステータス: stable
- 日付: 2026-11-03
- 関連: [設定ファイルを strict な TOML で書く](config-file-strict-toml.md), [Go の整形は gofmt、静的解析は go vet と golangci-lint で行う](gofmt-vet-golangci.md), GitHub Issue #395
- 起点: Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0056 §4 の依存の許可を書き直した

## 背景

設定ファイルは strict な TOML で書く（→ [設定ファイルを strict な TOML で書く](config-file-strict-toml.md)）。Go の標準ライブラリには
TOML のパーサが無い。layat の Go の依存は層ごとの許可制で、engine は標準ライブラリだけに依存する。

## 決定

CLI の依存に TOML パーサ `github.com/pelletier/go-toml/v2` を許可する。engine は標準ライブラリだけの
ままにする。

依存の許可の規範は [CLI が持ち込む依存は許可した第三者ライブラリと pin した Go に限り、いずれも固定する](../requirements/cli-dependency-policy.md) に置く。

## 根拠

TOML の仕様を満たすパーサを自前で書いて保守するより、既存のパーサに任せるほうが誤読の余地が小さい。
設定ファイルを読むのは CLI だけなので、依存を CLI の層に留めれば engine の方針は崩れない。

## 棄却した代替案

### TOML のパーサを自前で書く

仕様の網羅と保守を layat が抱えることになり、strict に読むはずの設定で誤読を生みやすい。

### 標準ライブラリで読める JSON を設定の形式にする

依存は増えないが、コメントを書けず手書きの設定に向かない（→ [設定ファイルを strict な TOML で書く](config-file-strict-toml.md)）。
