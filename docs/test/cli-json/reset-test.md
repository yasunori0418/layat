---
id: "CASE-020fa59e-9aaa-498b-ad5f-dd49b8b48f78"
type: test_case
name: "reset_test.go — 確認ポリシーと出力ストリームの排他"
status: stable
resource: "cmd/layat/reset_test.go"
covers:
  - "TC-b0e0cdb8-0c56-4eb8-8d4e-86569af85dec"
  - "TC-b4a365cd-0b1c-40ff-92c6-91f6cffe8e98"
  - "TC-7f30ff0f-a7bd-4a4b-ade9-7dc619207803"
follows:
  - "TP-e7c25263-6d2d-4a37-8275-26906889d912"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-19T00:08:45+09:00"
---
# reset_test.go — 確認ポリシーと出力ストリームの排他

## 対象

`cmd/layat/reset_test.go`

## 検証内容

- **確認ポリシー** — 同意フラグがあれば対話性を問わずプロンプトを出さないこと、対話環境で
  同意フラグが無ければプロンプトを出すこと、非対話環境で同意フラグが無ければ入力待ちで
  止まらず拒否すること
- **入力の解釈** — 明示的な肯定入力だけを肯定と読み、空行・未知の入力・入力の終端をいずれも
  否定として扱うこと
- **対話性の判定** — パイプやリダイレクトされた入力に対して非対話と判定されること。端末に
  接続されている側は自動テストの実行環境では作れないため対象外と明記されている
- **ストリームの排他** — 計画は stdout へ出して stderr へ漏らさないこと、結果レポートと
  削除対象の提示は stderr へ出して stdout へは出さないこと
- **暗黙の入力の記録** — `-f`・`--generator`・`--root` を省いた `reset --dryrun` が、探索で見つけた
  entrypoint と採用した generator をトップの `info` に、git の toplevel を `result.info.root` に出すこと。
  3 つとも与えた実行では、どちらの階層にも `info` が無いこと。`-f` だけを与えた実行の `info` が
  `generator` だけであること。entrypoint の探索に失敗した実行でも、トップの `info` が `{generator: nix}` で、
  `result.info` が無いこと

標準入力をパイプで差し替えて対話性を作り、stdout と stderr を同時に捕捉して排他を確かめる。
