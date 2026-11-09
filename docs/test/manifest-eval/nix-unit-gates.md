---
id: "CASE-879a93da-d22f-4397-84da-3544f8249af1"
type: test_case
name: "nix-unit: gates.nix — 評価時に throw する拒否条件群"
status: stable
resource: "tests/nix-unit/gates.nix"
covers:
  - "TC-e7ff0e6d-32d7-4ed6-8c2f-449dba34b2f6"
  - "TC-f1da084f-c0f0-4aba-85d9-f273c04eb373"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:28:26+09:00"
---
# nix-unit gates.nix

## 対象

`tests/nix-unit/gates.nix`

## 検証内容

拒否されるべき入力が `ThrownError` になり、メッセージが**拒否されたフィールド（または理由の
語）を名指しする**ことをアサートする（`expectedError.type` + `expectedError.msg` の部分一致）。

`msg` は部分一致なので、**フィールド名を本文に含む他の拒否条件のメッセージとも一致しうる**。
`"target"` を期待する 2 件（絶対パス / `..` 脱出）は互いに区別されないだけでなく、
`copy` 併用・subpath・root 相対 src の各メッセージが `(target: …)` を含むため、それらが誤って
発火した場合にも通る。拒否条件の取り違えまで検出する粒度は持っていない。

- `method = "copy"` と out-of-store marker の併用 → `"out-of-store"`
- out-of-store marker の相対パス（`rel/path`）→ `"must be absolute"`
- root 相対 marker の絶対パス（`/abs`）と `..` 脱出（`../x`）→ `"invalid root-relative src"`
- `method = "copy"` と root 相対 marker の併用 → `"cannot be combined with a root-relative marker"`
- root 相対 marker の `<src>/<subpath>` が target の祖先・target が `<src>/<subpath>` の祖先・
  両者が同一（`./` や末尾 `/` の表記揺れを含む）・root 内の `..` を畳むと target の祖先になる src
  （`dev/../.claude`）・root 自身（`.`）→ `"overlaps target"`。subpath を省いた entry の
  メッセージは src をそのまま出し既定値を `/.` として付けないこと、subpath 付きの entry では
  src と subpath を `/` で繋いだ形で出すこと。先頭の文字列が一致する
  だけの兄弟パス（`.claude` と `.claude-x/skills`）は重なりとみなさず通ること
- 絶対パスの target（`/etc/x`）→ `"target"`
- `..` で root の外へ出る target（`../../etc/x`）→ `"target"`
- `..` で src の外へ出る subpath（`../escape`）→ `"subpath"`
- 別キーから同一 target を明示指定した衝突 → `"same target"`
- 未知キー（旧名 `source`）→ strict submodule で `"source"`
- 素の文字列 src → `"src"`
- `lib/types.nix` の `entriesType` を `evalModules` で直接評価し、未知キー `bogus` が
  モジュールという使い方（`modules/common.nix` が共有する型）でも eval エラーになること

配置元は [評価テストの store-backed な入力は固定 outPath を持つ fake flake-input で与える](../../test-plan/eval-test-double.md) に従う。

## 出典

`tests/nix-unit/gates.nix` の現行実装からの逆算（→ Issue #273「L1〜L4」節）。各拒否条件の
設計判断は
[配置元の中のパスを全層で subpath と呼ぶ](../../adr/subpath-naming-across-layers.md) / [entry の未知のキーを評価時に拒否する](../../adr/strict-unknown-keys-rejected.md) / [src の型で素の文字列を拒否する](../../adr/srctype-bare-string-rejected.md) / [copy と out-of-store の併用を eval 時に拒否する](../../adr/copy-with-out-of-store-eval-error.md) /
[絶対パスと .. エスケープは eval 時に拒否する](../../adr/path-safety-eval-rejection.md) / [manifest 内の target の衝突を評価時に検出する](../../adr/manifest-target-conflict-eval-error.md) が持つ。
