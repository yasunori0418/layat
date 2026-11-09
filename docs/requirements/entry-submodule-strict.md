---
id: "REQ-3e446ad9-a6f4-4229-b5c5-184754b0ef51"
type: requirement
name: "entry submodule は strict とし未知キーと旧名を評価時エラーにする"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The entry submodule SHALL be strict: an unknown key SHALL be an evaluation-time error.
  A typo or an old field name (`name` / `source` / `dir` / `mode`) SHALL therefore fail
  at evaluation time instead of being silently ignored.
specification_ja: |
  entry submodule は strict（未知キー拒否）でなければならない。タイポや旧名
  （`name` / `source` / `dir` / `mode`）は利用者に知らせずに無視されるのではなく、評価時エラーと
  ならなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# entry submodule は strict とし未知キーと旧名を評価時エラーにする

## 仕様

entry submodule は **strict**（未知キー拒否）。タイポや旧名は評価時エラーになる。

旧名は改名を決めた ADR に対応する。

| 旧名 | 現在のフィールド | 改名を決めた ADR |
|---|---|---|
| `name` | （廃止・識別子は属性キー）| [entries を target をキーにした attrset にする](../adr/entries-keyed-by-target.md) |
| `source` | `subpath` | [配置元の中のパスを全層で subpath と呼ぶ](../adr/subpath-naming-across-layers.md) |
| `dir` | `subpath` | [配置元の中のパスを全層で subpath と呼ぶ](../adr/subpath-naming-across-layers.md) |
| `mode` | `method` | [entry の配置種別のフィールドを method と名付ける](../adr/mode-renamed-to-method.md) |

上の表は原文（`docs/spec.md` の旧名列挙と各 ADR 参照）を対応づけて整理したもので、
規範は frontmatter が正。

## 出典

`docs/spec.md`「entries スキーマ仕様」→「フィールド定義（entry submodule）」の
「submodule は **strict**（未知キー拒否）」。
