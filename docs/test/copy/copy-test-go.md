---
id: "CASE-dcd5e7e4-9190-4b13-9521-9b8801ec85fb"
type: test_case
name: "internal/engine/copy_test.go — コピー処理の syscall 失敗注入"
status: stable
resource: "internal/engine/copy_test.go"
covers:
  - "TC-4f315cfc-446c-4c3f-8dc9-c36b10073e9d"
  - "TC-d1eb1814-ac5e-4576-b092-7db4929fba43"
follows:
  - "TP-e7c25263-6d2d-4a37-8275-26906889d912"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-15T00:45:16+09:00"
---
# copy_test.go

## 対象

`internal/engine/copy_test.go`

コピー処理の各 syscall を実 FS の状態（パスの途中を通常ファイルで塞いで ENOTDIR を誘発する・
ディレクトリを読み取り元にする等、root でも成立する条件）で失敗させ、エラーが伝播することを
確認する故障注入テスト。結果レコードが汚れないことは `Result` を経由する 2 つの処理（copy 配置・
recopy）で併せて確認する。残る 6 本は `copyFile` / `copyTree` / `copySymlink` のプリミティブを
直接呼ぶため結果レコードを持たない。

## 主な検証内容

- **mkdir 失敗**: copy 配置時の親ディレクトリ作成失敗、ツリーコピー中のディレクトリ作成失敗
- **ファイル open 失敗**: 読み取り側 open の失敗、書き込み側 open-file の失敗
- **転送失敗**: read-write 中のコピー失敗
- **readlink 失敗**: symlink 複製時の readlink 失敗
- **既存の複製先の削除の失敗**: symlink 複製時に複製先が空でないディレクトリだと削除に
  失敗し、その失敗が捨てられずに返ること
- **recopy の lstat 失敗**: 非 ENOENT（パスの途中にある通常ファイルによる ENOTDIR 等）が「target
  不在」と同一視されずエラーになること、およびこのとき `Recopied` / `Copied` のいずれにも
  記録が残らないこと
