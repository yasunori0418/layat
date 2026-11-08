---
id: "TC-7d713505-5673-4e15-99bf-e98965e18892"
type: test_condition
name: "subdir 適用形が全 target に正規化済みの前置を施し、root と単体 marker を変えず、不正な subdir を評価時に拒否する"
status: stable
mitigates:
  - "RISK-e65844ea-a7b7-4bad-b15b-f60130b7e8fa"
  - "RISK-09df40d3-2752-433e-9ab0-2816fbd14969"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:49+09:00"
---
# TC-7d713505-5673-4e15-99bf-e98965e18892: subdir 適用形の前置・不変・拒否

## テスト条件

root マーカーの subdir 適用形（[root マーカーの subdir 適用形は全 target への前置と等価な評価時の糖衣で、root を動かさない](../../requirements/root-marker-subdir-sugar.md)）が、全 target への前置と
等価な糖衣として振る舞うことを確かめる。

- **marker の形**: 単体の 3 マーカーは `__functor` を除くと `_layatMarker` / `kind` の
  2 フィールドだけになる。適用形は `subdir` を足した marker で、`__functor` を持たない
- **適用形の引数**: `subdir` の欠落・未知キー・再適用が評価エラーになる
- **前置**: 既定 target と明示 target の全 entry に `<subdir>/` が前置され、`"a/"` と
  `"./a"` はどちらも `a/<target>` に正規化される。target 以外のフィールドは単体 marker の
  ときと変わらない
- **root 側の不変**: `rootKind` が単体 marker と同じ値で、`root` オブジェクトに `subdir` が
  現れず、`schemaVersion` も変わらない
- **拒否条件**: 絶対パス・`..` で root の外へ出る値・`""`・`"."`・正規化で空になる値の
  `subdir` が `"root subdir"` を含む `ThrownError` になる。entries が空でも拒否される。target の
  `..` 脱出は前置した後も従来どおり拒否される

前 4 項目は [subdir 適用形の前置が一部の target で抜ける、root 側へ混入する、または単体 marker の形が変わる](../../risks/subdir-prefix-drift.md)、拒否条件は [評価時に止めるべき入力が止まらずに engine 実行時まで到達する](../../risks/eval-gate-bypass.md) を
緩和する。文書全体の形は [manifest 文書全体をスナップショットに固定し、どの不変条件も名指ししない変更を差分として出す](manifest-document-regression.md) のスナップショットでも固定する。

## 覆う CASE

- [nix-unit: root-subdir.nix — root マーカーの subdir 適用形](nix-unit-root-subdir.md)（`tests/nix-unit/root-subdir.nix`）
- [namaka: manifest-subdir — subdir 適用形で生成した manifest 文書のスナップショット回帰](namaka-manifest-subdir.md)（`tests/namaka/manifest-subdir/expr.nix`）

## 出典

[root マーカーの subdir 適用形は全 target への前置と等価な評価時の糖衣で、root を動かさない](../../requirements/root-marker-subdir-sugar.md) の規範と [root マーカーに subdir の適用形を持たせる](../../adr/root-marker-subdir-applied-form.md) / [fixed root には subdir を提供しない](../../adr/no-subdir-for-fixed-root.md) / [subdir は全 target へ前置する評価時の糖衣にし、root を動かさない](../../adr/subdir-as-target-prefix-sugar.md) / [root 相対で指定するパスすべてにパス安全性の確認を課す](../../adr/path-safety-for-root-relative-inputs.md)。
