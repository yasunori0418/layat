---
id: "ADR-04cc2a03-78dd-4bf3-9cc7-3b8ed7faf429"
type: adr
name: "apply --all に root モードのフィルタを持たせる"
status: 採用
issues:
  - "#550"
origin: "実装前レビュー第 3 巡（2026-06-14）で確定。ADR-0017 §1 を書き直した"
justifies:
  - "REQ-d95b814f-aa7a-470e-9320-c14f9c14da7b"
supersedes:
  - "ADR-0017"
references:
  - "ADR-0036"
---
# ADR-04cc2a03-78dd-4bf3-9cc7-3b8ed7faf429: apply --all に root モードのフィルタを持たせる

- ステータス: 採用
- 日付: 2026-11-01
- 関連: ADR-0036, GitHub Issue #550
- 起点: 実装前レビュー第 3 巡（2026-06-14）で確定。ADR-0017 §1 を書き直した

## 背景

1 つの entrypoint に home mode と project mode の config が混在することがある。devShell の
shellHook から素の `apply --all` を打つと、project の config だけでなく home mode の config も
`$HOME` に配置されてしまう。

## 決定

`apply --all` に、root マーカーの名前に揃えた `--project-root` / `--home-root` / `--system-root` の
フィルタを持たせる。フィルタを付けると、その root モードの config だけを適用する。素の `--all` は
全 config を適用し、フィルタは opt-in の修飾にとどめる。

フィルタの規範（名指し apply との関係、root 権限で絞らずに実行したときの warning を含む）は
REQ-d95b814f-aa7a-470e-9320-c14f9c14da7b に置く。

## 根拠

フィルタを opt-in にすれば、混在した entrypoint で home mode の config まで適用してしまうのを
防ぎつつ、素の `--all` で全件を適用する使い方も残る。devShell は `apply --all --project-root` を
使えば、home mode の config に触れない。

フラグ名を root マーカーと揃えると、config の書き方と CLI の指定が 1 対 1 に対応する。
`--system-root` は system mode（ADR-0036）の config を選ぶ。

## 棄却した代替案

### `--all` を root モードに依らないまま残し、文書で注意する

誤適用を防ぐのを利用者の注意だけに頼ることになる。

### 混在した entrypoint で `--all` をエラーにする

誤適用は確実に防げるが、home と project を 1 つのリポジトリで管理する正当な使い方を妨げる。

### 値付きの `--all=<mode>` にする

フラグの数は減るが、真偽値の `--all` と値付きの形が混在し、cobra での表現が複雑になる。
