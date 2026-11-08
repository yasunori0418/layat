---
id: "ADR-37191429-2099-4217-a552-47b16a7e504c"
type: adr
name: "退避先が既にあれば conflict で止める"
status: stable
sources:
  - id: "issue-169"
    resource: "https://github.com/yasunori0418/layat/issues/169"
    title: "#169"
  - id: "issue-172"
    resource: "https://github.com/yasunori0418/layat/issues/172"
    title: "#172"
origin: "旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §3 を書き直した"
justifies:
  - "REQ-5dd5a4e9-6162-4fa5-b295-66844f5a4f3b"
  - "REQ-95e97d01-5c34-44b3-bc85-9ca53472bc3d"
references:
  - "ADR-65e41903-5eee-4a00-a154-be8c830cdca9"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T17:28:05+09:00"
---
# ADR-37191429-2099-4217-a552-47b16a7e504c: 退避先が既にあれば conflict で止める

- ステータス: stable
- 日付: 2026-11-03
- 関連: [退避先を target に suffix を付けた名前に固定する](backup-destination-suffix.md), GitHub Issue #169, #172
- 起点: 旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §3 を書き直した

## 背景

退避先は `<target>.<suffix>` に固定している（[退避先を target に suffix を付けた名前に固定する](backup-destination-suffix.md)）。
前回の `--backup` で退避したものが残っていれば、同じ名前がもう使われている。このときの
扱いを決める必要があった。

## 決定

退避先が既にあれば conflict にして apply 全体を止め、前の退避物を上書きも削除もしない。
この conflict は他の conflict と同じく全件を報告する。

conflict の種別と報告の形は [apply --backup は配置を塞ぐ記録外実体を rename 退避してから配置する](../requirements/apply-backup-rename-aside.md) と
[conflict で停止するときは全件を対処ガイダンス付きで列挙してから 1 本の集約エラーを返す](../requirements/conflict-full-report.md) に置く。

## 根拠

`--backup` は利用者のデータを守るための機能である。その退避物を利用者に知らせずに上書きすれば、機能
そのものがデータを失う別の原因になり、目的と矛盾する。同じコマンドを繰り返して通ることより
安全を優先する。

## 棄却した代替案

### 退避先が既にあっても上書きする

前回の退避物が失われ、守るための機能がデータを失わせることになる。

### 上書きを許す追加のフラグを設ける（`--force` のような二段階の opt-in）

操作が複雑になる割に、得られる便利さは小さい。古い退避物は利用者が確かめて消せば足りる。
