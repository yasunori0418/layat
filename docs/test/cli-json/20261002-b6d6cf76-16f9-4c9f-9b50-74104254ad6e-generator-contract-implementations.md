---
id: "TC-b6d6cf76-16f9-4c9f-9b50-74104254ad6e"
type: test_condition
name: "生成器の各実装が実 nix 無しで同じ契約（操作・失敗の構造）を満たす"
mitigates:
  - "RISK-31b8346c-7f1d-406f-aa74-6e1024edee83"
---
# TC-b6d6cf76-16f9-4c9f-9b50-74104254ad6e: 生成器の各実装が実 nix 無しで同じ契約（操作・失敗の構造）を満たす

## 条件

実 nix を起動せず、契約の実装（prebuilt・テストダブル・nix 生成器）と失敗型を確かめる。nix
生成器は PATH の先頭に置いた nix のスタブで駆動する。

- **失敗の構造** — `generator.Error` の文面は Message そのままで、原因の error を Unwrap で
  辿れる（`errors.Is` / `errors.As` が原因まで届く）。`%w` で包み直されてもマーカーとして
  取り出せる
- **nix 生成器の失敗** — 発見の失敗は Stage discover で原因を保持し、eval の失敗は Stage roots、
  build の失敗は Stage build になる。発見の失敗の文面は生成器化の前と同じ
- **prebuilt の操作** — Discover は与えられたパスを絶対パスにし、Build / DryBuild はその
  パスを返す。Roots は manifest.json の root を返し、Targets を entries から導く。読めない
  manifest.json は prebuilt・Stage roots の失敗になり、原因の `fs.ErrNotExist` を保持する。
  link-farm は 1 config なので AllRoots は失敗する
- **テストダブル** — 各操作は差し替えた戻り値（未設定なら零値）を返し、呼び出しを順に記録する
