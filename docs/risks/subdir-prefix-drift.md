---
id: "RISK-e65844ea-a7b7-4bad-b15b-f60130b7e8fa"
type: risk
name: "subdir 適用形の前置が一部の target で抜ける、root 側へ混入する、または単体 marker の形が変わる"
status: stable
likelihood: low
impact: medium
level: low
threatens:
  - "REQ-08c8f639-790c-4970-a954-c9567538cf9b"
  - "REQ-37b56673-6e40-4a1b-a2a7-5d3c084e3e66"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:28:26+09:00"
---
# RISK-e65844ea-a7b7-4bad-b15b-f60130b7e8fa: subdir 適用形の前置が一部の target で抜ける、root 側へ混入する、または単体 marker の形が変わる

## リスク

root マーカーの subdir 適用形の規範は [root マーカーの subdir 適用形は全 target への前置と等価な評価時の糖衣で、root を動かさない](../requirements/root-marker-subdir-sugar.md) である。全 entry の target に
`<subdir>/` を前置し、root も `rootKind` も動かさない。前提として、単体の
マーカーは従来どおり `kind` を運ぶだけの入れ物で、root の値の型は [root は 3 マーカーと絶対パス文字列の union を取る](../requirements/root-marker-kinds.md) が定める。この糖衣が崩れる形は 3 つある。

- **前置の抜け・揺れ**: 明示 target や既定 target の一部にだけ前置が掛からない。または
  `"a/"` と `"./a"` のような表記揺れが正規化されず、`a//x` や `./a/x` のような target になる
- **root 側への混入**: `subdir` が `manifest.json` の `root` オブジェクトへ出る、または
  `rootKind` が変わる。engine は `subdir` を知らないので、出ても無視されるか解釈を誤る
- **単体 marker の形の変化**: `__functor` 化に伴い、単体の `projectRoot` / `homeRoot` /
  `systemRoot` が従来の `_layatMarker` / `kind` 以外のフィールドを持つ

拒否の条件側の破れ（絶対パス・`..` 脱出・空の `subdir` が通る、前置で target の `..` 脱出が
隠れる）は本 `risk` ではなく [評価時に止めるべき入力が止まらずに engine 実行時まで到達する](eval-gate-bypass.md) が扱う。

**顕在化したときに起きること**: 前置が抜けた entry は `$HOME/.config/nvim` のつもりが
`$HOME/nvim` に置かれ、宣言からは読み取れない場所へ配置される。root 側へ混入した場合は、
`subdir` を知らない engine が manifest を警告なく受け入れ、利用者は適用形が反映されていないことに
気付けない。単体 marker の形が変わると、marker を exact 比較する既存の利用側で比較が一致しなくなる。

## 評価

- likelihood: low — 前置は `normalizeManifest` の 1 箇所で全 entry へ一様に掛かり、
  正規化は既存の成分分解を流用する。root 側は `rootKind` だけを写す既存の写像のまま変えない
- impact: medium — 配置先がずれても root の外へは出ない（root 外への脱出は拒否の条件が
  止める）。ずれた配置は layat の記録に残るため、古くなった配置の除去で原状回復できる

## 対処

[subdir 適用形が全 target に正規化済みの前置を施し、root と単体 marker を変えず、不正な subdir を評価時に拒否する](../test/manifest-eval/root-subdir-sugar.md)（全 target への正規化済みの前置・root と単体 marker の不変・
不正な `subdir` の評価時拒否）で緩和する。

## 出典

[root マーカーに subdir の適用形を持たせる](../adr/root-marker-subdir-applied-form.md) / [fixed root には subdir を提供しない](../adr/no-subdir-for-fixed-root.md) / [subdir は全 target へ前置する評価時の糖衣にし、root を動かさない](../adr/subdir-as-target-prefix-sugar.md) / [root 相対で指定するパスすべてにパス安全性の確認を課す](../adr/path-safety-for-root-relative-inputs.md) と、`tests/nix-unit/root-subdir.nix` /
`tests/namaka/manifest-subdir/expr.nix` の実装。
