---
id: "ADR-af37a636-a798-4c03-8fa4-e73ed6375a92"
type: adr
name: "apply --all の集約は config 名の辞書順に並べる"
status: stable
sources:
  - id: "issue-550"
    resource: "https://github.com/yasunori0418/layat/issues/550"
    title: "#550"
origin: "ドキュメント計画段階のレビュー（2026-06-13）で確定。ADR-0013 §3 と ADR-0016 §2 を書き直した"
justifies:
  - "REQ-4cbd9a0d-9f94-4747-8881-56020dc6d5af"
references:
  - "ADR-095220c9-5a3d-4fd4-94bd-8fa1a02dd36d"
  - "ADR-b3f3b68e-5096-46d1-bafd-53a9ab4711f4"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:39:38+09:00"
---
# apply --all の集約は config 名の辞書順に並べる

- ステータス: stable
- 日付: 2026-11-01
- 関連: [cross-config の同一 target は後勝ちにし、foreign symlink の上書きで警告する](cross-config-last-wins-with-foreign-warning.md), [配置の振動はユーザー責任とし、文書で注意するに留める](oscillation-is-user-responsibility.md), GitHub Issue #550
- 起点: ドキュメント計画段階のレビュー（2026-06-13）で確定。ADR-0013 §3 と ADR-0016 §2 を書き直した

## 背景

`apply --all` の結果の表示と、失敗の集約には、決まった並び順が要る。entrypoint の `layat.*` は
attrset で、Nix の attrset は定義順を保持しない。

## 決定

`apply --all` の集約表示と結果の並びは、config 名の辞書順（キーのソート順）にする。適用そのものの
順序は規範として決めない。

並び順の規範は [apply --all は選択された config を並列に適用し、集約表示と results\[\] を辞書順にし、部分失敗でも続行して最後に集約する](../requirements/apply-all-lexical-order-continue.md) に置く。

## 根拠

`builtins.attrNames` は辞書順を返すので、辞書順なら Nix 側でも Go 側でも追加の情報なしに決定的な
順序が得られる。各 config は独立した profile で atomic なので、config どうしが target を共有しなければ、
どの順で適用しても最終状態は変わらない。並びが要るのは、表示と失敗の集約を毎回同じ形にするためだけである。

[apply --all の配置を config 単位で並列に行う](apply-all-parallel-placement-per-config.md)。そのため、完了の順序は非決定になる。それでも集約を辞書順に
並べ直せば、出力は決定的なまま保てる。1 回の `apply --all` に載る config どうしの target 衝突の
うち、eval で検出できるものは適用前に止まる（[apply --all は選択した config 間の target 重複を build 前に止める](apply-all-preflight-target-conflict.md)）。検出できない衝突は実行時の後勝ち
（[cross-config の同一 target は後勝ちにし、foreign symlink の上書きで警告する](cross-config-last-wins-with-foreign-warning.md)）になり、どちらが残るかは適用順で変わる。
同じ target を複数の config で狙ったときは、[配置の振動はユーザー責任とし、文書で注意するに留める](oscillation-is-user-responsibility.md)。

## 棄却した代替案

### entrypoint の定義順に並べる

Nix の attrset が定義順を保持しないので実現できない。

### 完了した順に並べる

並列に適用すると実行のたびに順序が変わり、出力を比べたりテストで固定したりできなくなる。
