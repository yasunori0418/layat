---
id: "ADR-2aaf72ab-1dfc-4cde-852d-4b1f9916780b"
type: adr
name: "モジュールの activation は lock を待ち、失敗したら switch を止める"
status: stable
sources:
  - id: "issue-585"
    resource: "https://github.com/yasunori0418/layat/issues/585"
    title: "#585"
origin: "ADR-0026 §3b を書き直した"
justifies:
  - "REQ-8085f194-c903-4ecb-abd8-c719fe7b3292"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T15:56:40+09:00"
---
# ADR-2aaf72ab-1dfc-4cde-852d-4b1f9916780b: モジュールの activation は lock を待ち、失敗したら switch を止める

- ステータス: stable
- 日付: 2026-11-03
- 関連: GitHub Issue #585, `docs/spec.md`
- 起点: ADR-0026 §3b を書き直した

## 背景

activation から engine を起動するとき、lock の待ち方と失敗の扱いを決めておく必要がある。CLI には
lock を取れなければ待たずに諦める `--no-wait` があり、engine は conflict などで非 0 終了する。

## 決定

activation は `--no-wait` を付けずに engine を起動し、lock を取れるまで待つ。engine が非 0 で
終わったら activation もそのまま失敗させ、switch を止める。規範は
REQ-8085f194-c903-4ecb-abd8-c719fe7b3292 に置く。

## 根拠

switch は利用者が意図して行う直列の操作で、lock を待っても困らない。lock を取れないときに配置せずに終えると、
switch は成功したのに layat だけが配置されていない状態が、利用者に知らされないまま生まれる。

conflict を握りつぶさずに switch を止めるのは、`home.file` が既存ファイルとの衝突で switch を止める
のと同じ振る舞いである。宣言した状態と実際の状態がずれたまま switch が成功することを防げる。

## 棄却した代替案

### `--no-wait` で起動し、lock を取れなければ配置せずに終える

switch の成功と layat の未配置が両立してしまい、利用者は気付けない。

### engine の失敗を警告に留めて switch を続ける

宣言と実体がずれたまま host の世代だけが進む。
