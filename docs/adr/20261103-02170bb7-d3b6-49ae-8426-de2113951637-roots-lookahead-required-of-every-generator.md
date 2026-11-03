---
id: "ADR-02170bb7-d3b6-49ae-8426-de2113951637"
type: adr
name: "Roots の先読みを全ての生成器の前提条件にする"
status: 採用
issues:
  - "#395"
origin: "Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0055 §3 を書き直した"
justifies:
  - "REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729"
supersedes:
  - "ADR-0055"
references:
  - "ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9"
  - "ADR-bac005fb-fe4d-4a3b-b9e1-517a62fe46c2"
---
# ADR-02170bb7-d3b6-49ae-8426-de2113951637: Roots の先読みを全ての生成器の前提条件にする

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9, ADR-bac005fb-fe4d-4a3b-b9e1-517a62fe46c2, GitHub Issue #395
- 起点: Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0055 §3 を書き直した

## 背景

apply は eval を先に行って root の種別を得て、profileDir を確定し、flock を取ってから build する
（→ ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9）。生成器の契約では、build せずに root の種別を返すのが
Roots の役目である（→ ADR-bac005fb-fe4d-4a3b-b9e1-517a62fe46c2）。生成器によっては、build しないと
manifest の中身が分からないものもありうる。

## 決定

Roots は全ての生成器が build 前に答えなければならない契約の前提条件にする。先読みできる生成器だけが
Roots を持つ、という能力の宣言にはしない。Roots の規範は REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729 に置く。

## 根拠

eval 先行 → flock → build の順序は、build 前に profileDir を確定できることを前提にしている。先読みを
任意にすると、先読みできない生成器では flock の前に build する流れが生まれ、ロックの外での build に
よる out-link の競合が戻る。

## 棄却した代替案

### Roots を任意の能力にし、先読みできない生成器は build を先に行う

上記のとおり、ロックの外で build する流れができ、並行する apply が同じ pending out-link を奪い合う。
