---
id: "TC-8435052a-5dcc-49e2-ac26-82f645cb6890"
type: test_condition
name: "out-of-store の配置前の確認がリンク先不在と確認の失敗を区別すること"
status: stable
mitigates:
  - "RISK-e8449214-7794-4d0d-b584-a3a152e2c1f9"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:53:44+09:00"
---
# TC-8435052a-5dcc-49e2-ac26-82f645cb6890: out-of-store の配置前の確認

out-of-store entry の配置直前に marker のリンク先の存在を確かめ、不在のときは dangling
symlink を作らずに engine 実行時エラーで停止することを検証する。

確認自体が失敗した場合（権限エラー等の非 ENOENT）が「リンク先が存在しない」と同一視されず、
別のエラーとして報告されることを条件に含む。両者を混同すると、実際は確認できていないだけの
状況をユーザーが「marker のパスが間違っている」と誤診する。

この条件が見るのは src 側（marker のリンク先）の実在の確認に限られる。target 側の現況分類に
よる停止（実体占有・祖先 symlink）は [実体が占有する target と祖先 symlink を上書きせず停止すること](no-overwrite-stop.md) の担当で、確認の層が異なる。
