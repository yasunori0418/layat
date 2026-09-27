---
id: "CASE-a09f5137-2386-4bea-bcf9-a82504a988e6"
type: test_case
name: "apply_all_parallel_test.go — 完了順逆転時の辞書順集約・部分失敗と skip・カウントの並行安全性"
target: "cmd/layat/apply_all_parallel_test.go"
covers:
  - "TC-9ccb2ffe-7bf6-41ea-a9ce-df06cdda0509"
---
# CASE-a09f5137-2386-4bea-bcf9-a82504a988e6: apply_all_parallel_test.go — 完了順逆転時の辞書順集約・部分失敗と skip・カウントの並行安全性

## 対象

`cmd/layat/apply_all_parallel_test.go`

## 検証内容

- **完了順逆転時の辞書順** — 各 config が辞書順で次の config の完了を待ってから返る apply を
  注入して完了順を逆転させ（逆転が実際に起きたことも確かめる）、`-v` レポートの
  `layat: apply <name> done` と `results[]` が辞書順であること
- **部分失敗と skip** — 失敗・try-lock skip・成功を逆順に完了させ、stderr の failed /
  skipped / done 行が辞書順に並び、件数が (1, 1, 1)、失敗 config だけが error で、skip の
  subject は success、成功 config の subject が items ごと残ること
- **`--dryrun` の計画** — 同じ逆転の下で stdout の計画行（place / conflict）が辞書順で、
  conflict を含むと終了コード 2、`results[]` が辞書順であること
- **`--dryrun` の失敗行** — 失敗する config を逆順に完了させ、stderr の
  `layat: apply <name> --dryrun failed:` 行が辞書順に並び、終了コードが 1 であること
- **カウントの並行安全性** — 60 config を並列度 8 で成功・skip・失敗を 1/3 ずつ混ぜて流し、
  件数が (20, 20, 20) で正確なこと・同時実行数が 8 を超えないこと・`results[]` が選択順で
  あること（`go test -race` で競合を検出する）
