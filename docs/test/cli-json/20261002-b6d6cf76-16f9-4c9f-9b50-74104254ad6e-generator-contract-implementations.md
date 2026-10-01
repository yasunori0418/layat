---
id: "TC-b6d6cf76-16f9-4c9f-9b50-74104254ad6e"
type: test_condition
name: "生成器の各実装が nix 無しで同じ契約（操作・失敗の構造）を満たす"
mitigates:
  - "RISK-31b8346c-7f1d-406f-aa74-6e1024edee83"
---
# TC-b6d6cf76-16f9-4c9f-9b50-74104254ad6e: 生成器の各実装が nix 無しで同じ契約（操作・失敗の構造）を満たす

## 条件

nix を起動せず、契約の実装（prebuilt・テストダブル）と失敗型を直接確かめる。

- **失敗の構造** — `generator.Error` の文面は Message そのままで、原因の error を Unwrap で
  辿れる（`errors.Is` / `errors.As` が原因まで届く）。`%w` で包み直されてもマーカーとして
  取り出せる
- **prebuilt の操作** — Discover は与えられたパスを絶対パスにし、Build / DryBuild はその
  パスを返す。Roots は manifest.json の root を返し、Targets を entries から導く。読めない
  manifest.json は prebuilt・Stage roots の失敗になり、原因の `fs.ErrNotExist` を保持する
- **テストダブル** — 各操作は差し替えた戻り値（未設定なら零値）を返し、呼び出しを順に記録する
