---
id: "ADR-87a2cf34-9592-4c81-8980-246bca579ca0"
type: adr
name: "manifest の取得をプロセス内の Go インターフェイスの生成器越しにする"
status: 採用
issues:
  - "#395"
origin: "Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0055 §1 を書き直した"
justifies:
  - "REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729"
  - "DSG-25ad3cce-f921-4db2-aae0-c0263d4b8295"
  - "REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645"
---
# ADR-87a2cf34-9592-4c81-8980-246bca579ca0: manifest の取得をプロセス内の Go インターフェイスの生成器越しにする

- ステータス: 採用
- 日付: 2026-11-03
- 関連: GitHub Issue #395
- 起点: Issue #395（epic: manifest 生成器のインターフェイス化）の grilling（2026-09-09）。ADR-0055 §1 を書き直した

## 背景

engine が受け取る契約は `manifest.json` 1 本で、engine は誰が manifest を生成したかに関知しない。
一方 `cmd/layat` は nix を前提に組まれていた。entrypoint の発見、`layat.<system>.<name>` の attr path、
rootKind の先読み eval、`nix build` による link-farm の取得、nix 固有のエラー文字列による失敗の分類が
CLI 層に散らばり、「manifest の生成は nix でなくてもよい」という位置付けを阻んでいた。

## 決定

manifest の取得は、プロセス内の Go インターフェイスとして定義した manifest 生成器越しに行う。
cmd 層に残すのは生成器の選択と engine への注入だけで、engine は既存の注入口をそのまま使い変更しない。

生成器を経由するサブコマンドの範囲と契約の規範は REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729、
パッケージ構成は DSG-25ad3cce-f921-4db2-aae0-c0263d4b8295、engine と CLI の 2 層構成は
REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645 に置く。

## 根拠

実際に使う生成器は nix 1 つしかない。Go インターフェイスなら、nix 実装を切り出し、`--manifest` の
経路とテストダブルを同じ契約の実装にまとめるだけで、cmd 層から nix への依存を外せる。プロセス間の
規約を決めるのは、2 つ目の実生成器が現れてからで足りる。

## 棄却した代替案

### 外部プログラム規約（`layat-manifest-<lang>` を exec する）

2 つ目の実生成器が無い段階で、wire 形式・発見規則・バージョニングを決めることになり、使われない
仕様が固まる。

### 生成器の抽象を設けず、cmd 層の nix 呼び出しを保つ

生成器を足すたびに cmd 層全体へ分岐が入る。`--manifest` の経路とテストダブルも、nix の経路とは
別の実装のまま残る。
