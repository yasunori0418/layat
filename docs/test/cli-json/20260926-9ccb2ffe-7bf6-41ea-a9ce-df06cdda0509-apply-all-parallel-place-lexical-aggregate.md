---
id: "TC-9ccb2ffe-7bf6-41ea-a9ce-df06cdda0509"
type: test_condition
name: "配置並列の集約が完了順に依らず辞書順で、カウントが並行安全である"
mitigates:
  - "RISK-b7159827-3d29-405f-87ef-cdce7bb9a72f"
---
# TC-9ccb2ffe-7bf6-41ea-a9ce-df06cdda0509: 配置並列の集約が完了順に依らず辞書順で、カウントが並行安全である

## 条件

`apply --all` の stage 2（配置の並列実行）が、config の完了順に依らず決定的に集約する
ことを確かめる。

- **辞書順集約** — 完了順を逆転させても `results[]`・`-v` レポート・`--dryrun` の計画・
  失敗 / skip の行が辞書順に出る
- **カウントの並行安全性** — 多数の config が成功・skip・失敗を混ぜて並行に終わっても
  applied / skipped / failures が正確で、`-race` で検出されるデータ競合が無い。同時実行数は
  `--jobs` を超えない
- **意味論の維持** — 部分失敗でも成功 config の subject が結果ごと残り、try-lock の skip は
  失敗に数えず subject も success。`--dryrun` の終了コードは error > conflict > 0 のまま
