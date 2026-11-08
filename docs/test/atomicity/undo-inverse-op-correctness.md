---
id: "TC-9504e908-a4bc-4d2d-80d3-af07264284f8"
type: test_condition
name: "undo の逆操作が、対応する前方操作の直前状態を単体で復元する"
status: stable
mitigates:
  - "RISK-68e810c5-4e68-4b25-9bc0-6b2613022b49"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-05T00:26:53+09:00"
---
# TC-9504e908-a4bc-4d2d-80d3-af07264284f8: undo の逆操作が、対応する前方操作の直前状態を単体で復元する

## テスト条件

undo ジャーナルの各エントリ種別に対応する逆操作を、`Apply` を通さず個別に駆動し、
前方操作の直前状態が復元されることを検証する。

| 逆操作 | 対応する前方操作 | 復元すべき状態 |
|---|---|---|
| unlinkNew | 新規 symlink 配置（PlaceNew）| symlink が無い状態 |
| relinkOld | 張替え（PlaceReplace / PlaceForeign）・古くなった配置の除去・PreRemove の Unlink | 旧 dest を指す symlink |
| removeCopy | copy のマテリアライズ | copy tree が無い状態 |
| restoreRename | `--recopy` の rename 退避 | 退避元パスに退避前の内容 |
| mkdir | PreRemove の Rmdir | 空ディレクトリの再作成 |

undo ジャーナルは 6 種目として `--backup` の退避戻しを持つが、これは
`restoreRename` と実装（rename で戻す case 分岐）を共有するため単体駆動のテストを
持たず、`Apply` を通した end-to-end で覆う（[--backup の退避が opt-in で発動し、commit 後は残り、途中失敗では戻る](backup-aside-lifecycle.md) の担当）。

前方操作の対象が既に消えている場合（別プロセスに動かされた・そもそも配置に到達して
いなかった）に逆操作がエラーにならないことも条件に含む。「消すものが無い」「unlink すべき
現物が無いが再作成はできる」は復元の成功であり、失敗として扱えば [巻き戻し中の 1 件の失敗が残りの巻き戻しを道連れにし、復旧できない項目も報告されない](../../risks/unwind-best-effort-lost.md) の
「1 件の失敗が全体を道連れにする」を自ら誘発する。

上位の規範は [原子性は実 FS の条件で故障を誘発して不変条件ごとに検証する](../../test-plan/fault-injection-atomicity.md)（原子性の故障注入体系）と [エンジンとコマンド層の Go テストは nix を介さない実 FS 統合テストを主戦力とする](../../test-plan/go-test-layering.md)（`internal/engine/` の
実 FS 統合レベル）。

## 対応する CASE

[undo_test.go — 逆操作単体と unwind の順序・best-effort・journal 破棄](undo-test-go.md)（`internal/engine/undo_test.go`）。
