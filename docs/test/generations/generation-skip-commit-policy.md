---
id: "TC-1ee7a729-5629-4c92-b02a-4f6e11c0b57d"
type: test_condition
name: "世代スキップは project / system mode で derivation 同一のときだけ発火し、それ以外は新世代を積む"
status: stable
mitigates:
  - "RISK-67e96e59-005f-439d-a230-834b0228801c"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:49+09:00"
---
# TC-1ee7a729-5629-4c92-b02a-4f6e11c0b57d: 世代スキップは project / system mode で derivation 同一のときだけ発火し、それ以外は新世代を積む

## テスト条件

世代を積むか積まないかの判定を、mode と derivation の同一性の掛け合わせで検証する。
commit の呼び出し記録を観測点にして「新世代が積まれたか」を判定する。

| mode | link farm | 期待 |
|---|---|---|
| project | 前回と同一 | 世代を積まない（`shellHook` の高頻度実行で無限増殖させない）|
| project | 前回と異なる | 新世代を積む |
| system | 前回と同一 | 世代を積まない（boot ごとの activation で増殖させない）|
| home | 前回と同一 | 新世代を積む（世代スキップは project / system 限定）|

スキップ時も完全な no-op にはならず、`lstat` によるドリフトの確認だけは実行されること
（確認の中身は [lstat ドリフト修復が破損した entry だけを直し、ユーザーの編集には触れない](drift-repair-scope.md) が担当）。逆にスキップしない run では、修復ではなく通常の
plan / place の処理を通ること。

## 対応する CASE

[drift_test.go — 世代スキップの発火条件と lstat ドリフト修復の対象範囲](drift-test-go.md)（`internal/engine/drift_test.go`）。
system mode の root 再 apply で世代が増えないことは [e2e 12-system — user namespace の root で system mode の配置・世代・権限の条件を往復する](e2e-12-system.md)（e2e `12-system`）も確かめる。
