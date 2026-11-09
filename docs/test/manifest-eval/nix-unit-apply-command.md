---
id: "CASE-b2455b29-6f7a-451b-a667-9d0719808eb5"
type: test_case
name: "nix-unit: apply-command.nix — mkApplyCommand の文字列と root の確認"
status: stable
resource: "tests/nix-unit/apply-command.nix"
covers:
  - "TC-4d477235-5917-469b-926b-119266048643"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:49+09:00"
---
# nix-unit apply-command.nix

## 対象

`tests/nix-unit/apply-command.nix`（`layat.mkApplyCommand` を fake の pkgs・layatPackage・farm で
純評価する。layatPackage の fake は `lib.getExe` の derivation の確認を通すため `type = "derivation"` を持つ）

## 検証内容

- **root の 4 値** — `"all"` は bare `--all`、`"project"` / `"home"` / `"system"` は対応する
  `--<root>-root` を付けた文字列に exact 一致
- **辞書順** — 宣言順と名前順が食い違う 2 config が名前順の `--manifest` に並ぶ
- **引用** — 空白を含む名前でも `<name>=<farm>` 全体が 1 語に引用される
- **実行ファイル** — 先頭の語が `layatPackage` の `bin/<mainProgram>` になる
- **空の manifests** — `--manifest` を 1 つも出さない（CLI は entrypoint を使う処理になる）
- **root の確認** — 省略と範囲外の値（文字列でない値を含む）が `ThrownError` で止まる
