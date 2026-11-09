---
id: "TC-765858b6-d596-4428-b4da-5d432944a714"
type: test_condition
name: "lstat ドリフト修復が破損した entry だけを直し、ユーザーの編集には触れない"
status: stable
mitigates:
  - "RISK-67e96e59-005f-439d-a230-834b0228801c"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T13:42:21+09:00"
---
# lstat ドリフト修復が破損した entry だけを直し、ユーザーの編集には触れない

## テスト条件

世代スキップ時に実行される `lstat` ベースのドリフト修復について、「何を直し、何を直さないか」を
検証する。同一の link farm を 2 回 apply して `generationUnchanged` を成立させ、2 回目の
run が修復だけを行う状況を作る。

**直すもの**

- 消された symlink を再張りする
- 別物へ張り替えられた foreign symlink を記録どおりへ戻す。この修復は warning を伴う
  （他ツールとの競合をユーザーへ知らせる）
- 不在になった copy target を place-once で再マテリアライズする

**直さないもの**

- 記録どおりの entry（ドリフト無し）— 何もせず、warning も出さない
- 内容が編集された copy target — 存在するが異なる場合は触らない（home mode の place-once
  と振る舞いを揃える）。`--recopy` を明示した run では逆に上書きする

修復対象は symlink と copy の両方であること（片方だけの修復に縮退していないこと）を
条件に含む。

上位の規範は [エンジンとコマンド層の Go テストは nix を介さない実 FS 統合テストを主戦力とする](../../test-plan/go-test-layering.md)（`internal/engine/` を実 FS の tmpdir で駆動する統合レベル）。

## 対応する CASE

[drift_test.go — 世代スキップの発火条件と lstat ドリフト修復の対象範囲](drift-test-go.md)（`internal/engine/drift_test.go`）。`--backup` が同じ修復の処理で発火する
側は atomicity の [--backup の退避が opt-in で発動し、commit 後は残り、途中失敗では戻る](../atomicity/backup-aside-lifecycle.md) が担当する。
