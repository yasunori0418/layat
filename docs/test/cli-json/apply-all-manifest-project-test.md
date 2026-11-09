---
id: "CASE-c4ffa8b1-4058-4cba-b911-520ee970b2ae"
type: test_case
name: "apply_all_manifest_project_test.go — project と home の link-farm を apply --all --manifest で配置し root フィルタで選別する"
status: stable
resource: "cmd/layat/apply_all_manifest_project_test.go"
covers:
  - "TC-4d477235-5917-469b-926b-119266048643"
  - "TC-b4a365cd-0b1c-40ff-92c6-91f6cffe8e98"
follows:
  - "TP-e7c25263-6d2d-4a37-8275-26906889d912"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-05T00:26:57+09:00"
---
# apply_all_manifest_project_test.go

## 対象

`cmd/layat/apply_all_manifest_project_test.go`（`runApplyAll` を `--manifest` の集合で実 engine まで
駆動する。link-farm は rootKind を指定できる `writeRootKindLinkFarm` で書く）

## 検証内容

`git init` した一時ディレクトリのサブディレクトリへ移り、rootKind が project と home の link-farm を
1 つずつ渡す。

- **フィルタなし** — project 側が git toplevel 配下に、home 側が隔離した `$HOME` 配下に置かれる
- **`--project-root` / `--home-root`** — 各 root の config だけが置かれる
- **`--system-root` で該当なし** — どの config も置かれず終了コードは 0。指定が無いときは stderr が
  空で、`-v` のときだけ `layat: apply --all: no matching configs` の 1 行が出る
- **配置先** — project 側はサブディレクトリ（cwd）にも `$HOME` にも置かれない
- **nix 非依存** — どの実行も nix を呼ばない

`--root` は rootKind の解決を迂回するため使わない。
