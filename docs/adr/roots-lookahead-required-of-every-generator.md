---
id: "ADR-02170bb7-d3b6-49ae-8426-de2113951637"
type: adr
name: "Roots の先読みを全ての生成器の前提条件にする"
status: stable
sources:
  - id: "issue-395"
    resource: "https://github.com/yasunori0418/layat/issues/395"
    title: "#395"
origin: "Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0055 §3 を書き直した"
justifies:
  - "REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729"
references:
  - "ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9"
  - "ADR-bac005fb-fe4d-4a3b-b9e1-517a62fe46c2"
  - "ADR-28605860-770e-41df-95dd-7dc05b20497a"
  - "ADR-51863e94-6a9d-4029-95e7-a7ad2bfa93c1"
  - "ADR-5caf1c14-a054-4c8d-95a2-c7f89e0c5246"
  - "ADR-63039f93-cd78-4356-8cd3-82c2f49a1aaa"
  - "ADR-6b7becf0-f029-41e3-895e-b672dce94d3d"
  - "ADR-c7ffc21a-4be1-4468-b63c-2f47eedd9ab3"
  - "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:41+09:00"
---
# ADR-02170bb7-d3b6-49ae-8426-de2113951637: Roots の先読みを全ての生成器の前提条件にする

- ステータス: stable
- 日付: 2026-11-03
- 関連: [eval を先に行い、flock を取ってから build する](eval-first-then-flock-then-build.md), [生成器の契約を Discover / Roots / Build / DryBuild の 4 操作にする](generator-four-operations.md), [apply --all は config ごとの情報を 1 回の eval でまとめて取る](apply-all-single-eval.md), [store path は out-link で受け取る](store-path-via-out-link.md), [内部コマンドの開示で透明性を保つ](internal-command-disclosure.md), [nix の experimental-features は利用者が有効にしておく前提にする](experimental-features-user-prerequisite.md), [CLI は cobra で組む](cobra-cli-framework.md), [root の必須性を別の確認を書かずに標準のエラーで担保する](root-required.md), [home mode の CLI と lib の版ずれを manifest の確認に任せて許容する](home-mode-skew-guarded-by-schemaversion.md), [1 config = 1 profile を全モードで自前に持つ](one-config-one-profile-all-modes.md), GitHub Issue #395
- 起点: Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0055 §3 を書き直した

## 背景

apply は eval を先に行って root の種別を得て、profileDir を確定し、flock を取ってから build する
（→ [eval を先に行い、flock を取ってから build する](eval-first-then-flock-then-build.md)）。生成器の契約では、build せずに root の種別を返すのが
Roots の役目である（→ [生成器の契約を Discover / Roots / Build / DryBuild の 4 操作にする](generator-four-operations.md)）。生成器によっては、build しないと
manifest の中身が分からないものもありうる。

## 決定

Roots は全ての生成器が build 前に答えなければならない契約の前提条件にする。先読みできる生成器だけが
Roots を持つ、という能力の宣言にはしない。Roots の規範は [CLI は manifest 生成器の契約（Discover / Roots / Build / DryBuild）越しに manifest を得る](../requirements/manifest-generator-contract.md) に置く。

## 根拠

eval 先行 → flock → build の順序は、build 前に profileDir を確定できることを前提にしている。先読みを
任意にすると、先読みできない生成器では flock の前に build する流れが生まれ、ロックの外での build に
よる out-link の競合が再び起きる。

## 棄却した代替案

### Roots を任意の能力にし、先読みできない生成器は build を先に行う

上記のとおり、ロックの外で build する流れができ、並行する apply が同じ pending out-link を奪い合う。
