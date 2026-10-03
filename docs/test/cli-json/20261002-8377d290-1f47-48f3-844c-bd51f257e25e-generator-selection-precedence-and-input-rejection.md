---
id: "TC-8377d290-1f47-48f3-844c-bd51f257e25e"
type: test_condition
name: "生成器の選択が優先順位どおりに解決し、入力不正を E_INPUT で拒否し、--manifest 経路では環境を読まない"
mitigates:
  - "RISK-88088f4c-348f-4fa5-a2c1-dc42bb328b5d"
---
# TC-8377d290-1f47-48f3-844c-bd51f257e25e: 生成器の選択が優先順位どおりに解決し、入力不正を E_INPUT で拒否し、--manifest 経路では環境を読まない

## 条件

- **優先順位** — `--generator` > `LAYAT_GENERATOR` > `layat.toml` > `config.toml` > 既定 `nix`
  の各段が下の段に勝ち、値の決まった段より下の設定ファイルは読まない（壊れていても止まらない）
- **指定なし** — 空の環境変数・ファイルの不在・`generator` キーの無い設定ファイルは次の段へ進む
- **入力不正** — 未知の値（どの段でも）・TOML のパース失敗・未知キーは出どころを名指しした
  1 行の `inputError` になり、`--json` では `E_INPUT` に分類される。存在するが読めない設定
  ファイルは入力不正にせず読み込み失敗の分類を保つ
- **フラグの組み合わせ不正** — `--manifest` と `-f` / `--all` / `--generator` の同時指定が
  `E_INPUT` で止まり、nix を呼ばない
- **`--manifest` 経路の環境非依存** — `apply --manifest` は環境変数・設定ファイルが不正でも
  読まずに prebuilt で配置する
