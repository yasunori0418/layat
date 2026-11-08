---
id: "ADR-ca24e802-cc57-40c5-ad03-a1c5f0b6e498"
type: adr
name: "mkApplyCommand の root を既定値のない必須の enum にする"
status: stable
sources:
  - id: "issue-530"
    resource: "https://github.com/yasunori0418/layat/issues/530"
    title: "#530"
origin: "Issue #530 の grilling（2026-10-25）。ADR-0068 §2 を書き直した"
justifies:
  - "REQ-6c51facc-fa7a-4b5e-939b-44e7578d7a4f"
references:
  - "ADR-04cc2a03-78dd-4bf3-9cc7-3b8ed7faf429"
  - "ADR-db293e0c-4789-45b7-93a9-adedc021344f"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T17:34:07+09:00"
---
# ADR-ca24e802-cc57-40c5-ad03-a1c5f0b6e498: mkApplyCommand の root を既定値のない必須の enum にする

- ステータス: stable
- 日付: 2026-11-03
- 関連: [apply --all に root モードのフィルタを持たせる](apply-all-root-mode-filter.md), [lib.mkApplyCommand で事前展開した apply コマンドを組む](lib-mkapplycommand.md), GitHub Issue #530
- 起点: Issue #530 の grilling（2026-10-25）。ADR-0068 §2 を書き直した

## 背景

`mkApplyCommand`（[lib.mkApplyCommand で事前展開した apply コマンドを組む](lib-mkapplycommand.md)）が組むのは `apply --all` の行である。
素の `--all` は全 config を適用し、root モードで絞るには `--project-root` などのフィルタを付ける
（[apply --all に root モードのフィルタを持たせる](apply-all-root-mode-filter.md)）。

## 決定

`mkApplyCommand` の `root` は既定値を持たない必須の引数にし、`"all"` / `"project"` / `"home"` /
`"system"` だけを受ける。省略や範囲外の値は評価時に throw する。

各値が付けるフィルタの規範は [lib.mkApplyCommand は事前展開した apply --all --manifest のコマンド行を返す](../requirements/lib-apply-command.md) に置く。

## 根拠

devShell から素の `--all` を実行すると、home mode の config まで `$HOME` に置かれる。既定値を
持たせなければ、どの root を置くかを呼び出し側が必ず書くことになり、意図しない配置に書き手が
気付かないままにはならない。

値を CLI の root フィルタと同じ名前の enum にすれば、引数と付くフラグが 1 対 1 に対応する。
誤った値は CLI を実行する前の評価時に止まる。

## 棄却した代替案

### `"all"` を既定値にする

モジュールの呼び出しは短くなるが、devShell で省略すると home mode の config まで置かれる。

### `"project"` を既定値にする

devShell での典型的な使い方には合うが、省略したときにどの root が置かれるかが呼び出し側の記述から
読み取れなくなる。
