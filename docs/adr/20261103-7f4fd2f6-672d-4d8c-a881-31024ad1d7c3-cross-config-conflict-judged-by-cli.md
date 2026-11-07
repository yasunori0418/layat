---
id: "ADR-7f4fd2f6-672d-4d8c-a881-31024ad1d7c3"
type: adr
name: "config 間の target 衝突は CLI が判定し engine に持ち込まない"
status: 採用
issues:
  - "#584"
origin: "ADR-0038 §2a（データ算出は nix・検出判定は Go CLI・engine は不変）を書き直した"
justifies:
  - "REQ-535b811d-dfc5-4eac-92db-737e70eb5415"
references:
  - "ADR-75df92ae-0694-452b-8fb4-350994e31b25"
  - "ADR-28605860-770e-41df-95dd-7dc05b20497a"
  - "ADR-095220c9-5a3d-4fd4-94bd-8fa1a02dd36d"
  - "ADR-af37a636-a798-4c03-8fa4-e73ed6375a92"
  - "ADR-b3f3b68e-5096-46d1-bafd-53a9ab4711f4"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:30+09:00"
---
# ADR-7f4fd2f6-672d-4d8c-a881-31024ad1d7c3: config 間の target 衝突は CLI が判定し engine に持ち込まない

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-75df92ae-0694-452b-8fb4-350994e31b25, ADR-28605860-770e-41df-95dd-7dc05b20497a, ADR-095220c9-5a3d-4fd4-94bd-8fa1a02dd36d, ADR-af37a636-a798-4c03-8fa4-e73ed6375a92, ADR-b3f3b68e-5096-46d1-bafd-53a9ab4711f4
- 起点: ADR-0038 §2a（データ算出は nix・検出判定は Go CLI・engine は不変）を書き直した

## 背景

`apply --all` は、選択した config 間の target 重複を build の前に止める
（ADR-75df92ae-0694-452b-8fb4-350994e31b25）。この確認に要る正規化後の target をどこで算出し、
重複の判定をどの層に置くかを決める必要があった。

`apply --all` は build の前に、config ごとの情報を 1 回の eval でまとめて取っている
（ADR-28605860-770e-41df-95dd-7dc05b20497a）。どの config を対象にするかを決める root モードの
フィルタは、CLI にしかない。engine は 1 回の起動で 1 つの manifest だけを扱う。

## 決定

正規化後の target は lib が manifest の passthru として算出し、CLI が `apply --all` の一括 eval で
rootKind と一緒に受け取る。`--manifest` で渡されたビルド済みの集合では、CLI が
各 link-farm の `manifest.json` から target を読む。重複の判定は CLI が行う。engine には手を入れず、config 間の衝突を
engine へ持ち込まない。

一括 eval で取る値と式は REQ-535b811d-dfc5-4eac-92db-737e70eb5415 に置く。

## 根拠

確認の対象になる集合は、CLI のフィルタで選んだ config に依存する。選択の処理が CLI にある以上、判定も
CLI に置くのが素直である。エラーメッセージを組み立てるのも CLI のほうが制御しやすい。

target を passthru で取れば、rootKind と同じく build せずに安い eval で先取りできる。一括 eval の
返り値が増えるだけで、eval の回数は増えない。

engine は 1 つの manifest しか見ないことで層を分けている。config 間の衝突を engine で扱うには、
複数の manifest を engine に渡すようインターフェイスを変えなければならない。

## 棄却した代替案

### engine で検出する

1 回の起動で 1 つの manifest だけを扱うという層の分け方が崩れ、複数の manifest を渡すインターフェイスの変更が要る。

### 一括 eval の `--apply` 式の中で throw する

フィルタによる選択を nix 式に持ち込むことになり、エラーメッセージも制御しにくい。データの算出と
判定を分けたほうが層として素直である。
