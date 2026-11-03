---
id: "SDR-2457aa00-fab7-4c99-af53-2e6643c21c79"
type: sara_adr
name: "model.yaml で組み込みモデルを全面置換し、汎用の型名を使う"
status: 採用
issues:
  - "#203"
origin: "epic #203 の grilling セッション（2026-07-31・2026-08-01）で確定。ADR-0048 §2 を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
supersedes:
  - "ADR-0048"
references:
  - "SDR-c15921a9-7aea-407c-a34b-7465783b1af1"
  - "SDR-e093c6b2-75fd-4c91-a8f0-5a89e005e4ed"
---
# SDR-2457aa00-fab7-4c99-af53-2e6643c21c79: model.yaml で組み込みモデルを全面置換し、汎用の型名を使う

- ステータス: 採用
- 日付: 2026-11-01
- 関連: SDR-c15921a9-7aea-407c-a34b-7465783b1af1, SDR-e093c6b2-75fd-4c91-a8f0-5a89e005e4ed, GitHub Issue #203, `docs/model.yaml`
- 起点: epic #203 の grilling セッション（2026-07-31・2026-08-01）で確定。ADR-0048 §2 を書き直した

## 背景

sara は組み込みのモデル（型と relation）を持ち、カスタムスキーマで置き換えられる。置き換えは
全面で、部分的なマージはできない。組み込みには `hardware_requirement` や `system_architecture` の
ように、layat に対応する実体の無い型が含まれる。

## 決定

`docs/model.yaml` で組み込みモデルを全面置換し、layat の文書構造に合わせた型だけを定義する。
ハードウェア系の型と、対応する中間層の無い `scenario` / `system_architecture` は定義しない。

型名とフィールド名は layat 固有の語を混ぜず、汎用の名前にする（例: `requirement` / `design` / `adr`）。
型の内訳と relation は `docs/model.yaml` に置き、その変遷は SDR-34916a83-0cff-4524-a8ff-5a3195af52c3・SDR-49733f2e-ed79-480d-bfb9-6361a7f95aed などが持つ。

## 根拠

使わない型が並ぶと、読む人が「ここを埋めるべきか」を毎回判断することになる。全面置換しかできない
以上、必要な型だけを書くのが素直である。

型名を汎用にしておけば、モデルを別のリポジトリへ切り出したときにそのまま使える。

## 棄却した代替案

### 組み込みモデルをそのまま使う

対応する実体の無い型が残る。system / software の 2 段の要求も、layat には区別が無い。

### layat 固有の型名を使う

切り出し先で型名を付け直すことになる。
