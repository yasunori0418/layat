---
id: "TC-4f315cfc-446c-4c3f-8dc9-c36b10073e9d"
type: test_condition
name: "コピー中の syscall 失敗が伝播し、失敗した entry が結果へ記録されないこと"
status: stable
mitigates:
  - "RISK-bc7526c2-2a2d-427f-9490-92d7d262fce3"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:25:13+09:00"
---
# コピー失敗の伝播

コピー処理の各 syscall（mkdir / open / open-file / read-write / readlink / lstat）の
失敗がそれぞれエラーとして呼び出し元へ伝わることを検証する。エラーメッセージによる段の特定は、
診断上の要となる 2 つの処理（copy 配置時の親ディレクトリ作成・recopy の lstat）で確認し、残る段は
伝播そのものまでを見る。

失敗した entry が結果レコードへ「コピー済み」として残らないことも、結果を経由する 2 つの処理
（copy 配置・recopy）で条件に含む。非 ENOENT の lstat 失敗が「target 不在」と同一視されない
ことは、recopy の処理で特に重要になる。

foreign な実ファイルの skip が warning として可視化されることは、engine 側の配置処理で検証
する（`engine-core` の [internal/engine/engine_test.go — 実 tmpdir に対する apply の統合テスト](../engine-core/engine-test-go.md) が扱う。検証の所在を示すだけで relation は張らない）。
