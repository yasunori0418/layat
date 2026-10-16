---
id: "TC-e54fe715-7b49-431c-a611-25e196b1d0a0"
type: test_condition
name: "apply --all が --manifest <name>=<link-farm> の集合に乗り、集合の形の不正を E_INPUT で拒否する"
mitigates:
  - "RISK-88088f4c-348f-4fa5-a2c1-dc42bb328b5d"
---
# TC-e54fe715-7b49-431c-a611-25e196b1d0a0: apply --all が --manifest <name>=<link-farm> の集合に乗り、集合の形の不正を E_INPUT で拒否する

## 条件

- **集合の適用** — `apply --all --manifest <name>=<link-farm>` を config ごとに繰り返すと、
  各組が config `<name>` として配置され、報告は指定順によらず辞書順になる。entrypoint の
  発見・nix の eval / build を一切行わない
- **名前による選択** — prebuilt 生成器の集合は config 名ごとに root（Targets は entries から
  導出）を返し、Roots / Build / DryBuild は名前で link-farm を選ぶ。集合に無い名前は NotFound
- **前段衝突検査** — 2 つの link-farm が同じ root の同じ正規化後 target を持つとき、配置前に
  止まり、メッセージに target と両 config 名が載る
- **形の不正** — `--all` なしの複数指定・`--all` なしの `<name>=` 形・`--all` と `<link-farm>` 形・
  空の `<name>`・`--all` での同じ `<name>` の重複・`-f` / `--generator` との併用は、nix を
  呼ぶ前に `E_INPUT` で止まる
- **単一形の不変** — `apply --manifest <link-farm> [name]` は従来どおり 1 つの link-farm を
  名前によらず適用する
