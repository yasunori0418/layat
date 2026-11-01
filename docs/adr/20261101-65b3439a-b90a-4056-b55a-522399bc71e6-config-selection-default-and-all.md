---
id: "ADR-65b3439a-b90a-4056-b55a-522399bc71e6"
type: adr
name: "apply の config は名前省略で default を選び、--all で全件を選ぶ"
status: 採用
issues:
  - "#4"
  - "#550"
origin: "CLI 化と entrypoint の再設計（Issue #2・#3・#4、2026-06-13）で確定。ADR-0007 §4 を書き直した"
justifies:
  - "REQ-c2d44626-d8f4-446a-a80a-319a500129b4"
---
# ADR-65b3439a-b90a-4056-b55a-522399bc71e6: apply の config は名前省略で default を選び、--all で全件を選ぶ

- ステータス: 採用
- 日付: 2026-11-01
- 関連: GitHub Issue #4, #550
- 起点: CLI 化と entrypoint の再設計（Issue #2・#3・#4、2026-06-13）で確定。ADR-0007 §4 を書き直した

## 背景

entrypoint は `layat.<name>` に複数の config を公開できる。1 つのプロジェクトに役割の違う config
（skills・vim-plugins など）を置き、それぞれを独立した profile で管理するためである。CLI は、
どの config を適用するかを利用者の指定から決める必要がある。

## 決定

`layat apply` は、名前を省略すると `layat.default` を、名前を指定するとその config を、`--all` を
付けると `layat.*` の全 config を適用する。`default` が無いのに名前を省略したらエラーにする。

選択の規範は REQ-c2d44626-d8f4-446a-a80a-319a500129b4 に置く。

## 根拠

flake の `default` の慣例に倣えば、config が 1 つのプロジェクトでは名前を書かずに済む。`default` が
無いときに別の config を推測して選ぶと、意図しない配置が起きうるのでエラーにする。

全件の適用は、名前を並べさせずに明示のフラグで受ける。省略時の既定を全件にしないのは、名前を
省いただけで全 config が配置されると、混在した entrypoint で影響範囲が広がりすぎるためである。

## 棄却した代替案

### 名前を省略したら全 config を適用する

既定の操作の影響範囲が最も広くなる。home mode と project mode が混在する entrypoint では、名前を
省いた apply が `$HOME` にも配置する。
