---
id: "ADR-41d506b4-2c5c-4a17-9954-b0de37e97de5"
type: adr
name: "copy は foreign の実ファイルを上書きせず warning を出して見送る"
section: placement
status: stable
sources:
  - id: "issue-584"
    resource: "https://github.com/yasunori0418/layat/issues/584"
    title: "#584"
origin: "ADR-0022 §4（copy が foreign 実ファイルに当たったら warning を出して skip する）を書き直した"
justifies:
  - "REQ-07c3b735-3744-4778-a640-8c6fb66f4aa7"
references:
  - "ADR-c4ad9987-850e-4834-bbea-d8de3b073548"
  - "ADR-095220c9-5a3d-4fd4-94bd-8fa1a02dd36d"
  - "ADR-45380022-d959-4b22-aa05-5993ac1f3c87"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:41+09:00"
---
# copy は foreign の実ファイルを上書きせず warning を出して見送る

- ステータス: stable
- 日付: 2026-11-03
- 関連: [copy は世代の外に place-once で置く](copy-outside-generations-place-once.md), [cross-config の同一 target は後勝ちにし、foreign symlink の上書きで警告する](cross-config-last-wins-with-foreign-warning.md), [flock は明示実行で待ち、hook 実行でスキップする](flock-mode-explicit-wait-hook-skip.md)
- 起点: ADR-0022 §4（copy が foreign 実ファイルに当たったら warning を出して skip する）を書き直した

## 背景

copy は place-once で、target が既にあれば触らない（[copy は世代の外に place-once で置く](copy-outside-generations-place-once.md)）。
target にある実ファイルが layat の置いた物でなくても、warning を出さずに見送っていた。利用者には、layat が
中身を置いたように見えてしまう。

symlink の側では、記録に無い symlink を上書きするときに warning を出す
（[cross-config の同一 target は後勝ちにし、foreign symlink の上書きで警告する](cross-config-last-wins-with-foreign-warning.md)）。copy と symlink とで、foreign な物への扱いが揃って
いなかった。

## 決定

copy entry の target に、前世代の manifest に無い実ファイルが既にあるときは、上書きせずに copy を
見送り、warning を出す。apply 全体は止めない。layat が置いた物かどうかは、前世代の manifest に
その entry があるかで判別し、内容では判別しない。

規範は [copy が foreign 実ファイルを skip したときは warning で可視化する](../requirements/copy-foreign-skip-warning.md) に置く。

## 根拠

warning を出せば、layat が中身を置いていないことが利用者に見える。symlink の foreign 警告と
扱いが揃う。

copy は利用者のデータを上書きしない。既存のファイルが 1 つあるだけで apply 全体が止まると、
devShell の shellHook から呼ぶ project mode ではシェルに入るたびに失敗する。見送って warning に留めれば、
データを守りながら apply を止めずに済む。

前世代の manifest で判別すれば、内容を読まずに済む。copy の中身は利用者が編集してよいので、
内容で判別すると layat の置いた物を foreign と取り違える。

## 棄却した代替案

### エラーで止める

symlink の foreign 実ファイルと同じ扱いになるが、copy はもともと上書きしない方針である。
既存のファイル 1 つで apply 全体が止まり、shellHook から呼ぶと毎回失敗する。

### これまでどおり warning を出さずに見送る

実装は変わらないが、layat が中身を置いたという誤解が残る。
