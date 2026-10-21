---
id: "RISK-e65844ea-a7b7-4bad-b15b-f60130b7e8fa"
type: risk
name: "subdir 適用形の前置が一部の target で抜ける、root 側へ漏れる、または単体 marker の形が変わる"
likelihood: low
impact: medium
level: low
threatens:
  - "REQ-08c8f639-790c-4970-a954-c9567538cf9b"
  - "REQ-37b56673-6e40-4a1b-a2a7-5d3c084e3e66"
---
# RISK-e65844ea-a7b7-4bad-b15b-f60130b7e8fa: subdir 適用形の前置が一部の target で抜ける、root 側へ漏れる、または単体 marker の形が変わる

## リスク

root マーカーの subdir 適用形（REQ-08c8f639-790c-4970-a954-c9567538cf9b）は、全 entry の target に
`<subdir>/` を前置する評価時の糖衣で、root も `rootKind` も動かさない。前提として、単体の
マーカーは従来どおり kind を運ぶだけの入れ物で、root の値はマーカーと絶対パス文字列の union を
取る（REQ-37b56673-6e40-4a1b-a2a7-5d3c084e3e66）。この糖衣が崩れる形は 3 つある。

- **前置の抜け・揺れ**: 明示 target や既定 target の一部にだけ前置が掛からない。または
  `"a/"` と `"./a"` のような表記揺れが正規化されず、`a//x` や `./a/x` のような target になる
- **root 側への漏れ**: `subdir` が `manifest.json` の `root` オブジェクトへ出る、または
  `rootKind` が変わる。engine は `subdir` を知らないので、出ても無視されるか解釈を誤る
- **単体 marker の形の変化**: `__functor` 化に伴い、単体の `projectRoot` / `homeRoot` /
  `systemRoot` が従来の `_layatMarker` / `kind` 以外のフィールドを持つ

拒否ゲート側の破れ（絶対パス・`..` 脱出・空の `subdir` が通る、前置で target の `..` 脱出が
隠れる）は本 risk ではなく RISK-09df40d3-2752-433e-9ab0-2816fbd14969 が扱う。

**顕在化したときに起きること**: 前置が抜けた entry は `$HOME/.config/nvim` のつもりが
`$HOME/nvim` に置かれ、宣言からは読み取れない場所へ配置される。root 側へ漏れた場合は、
`subdir` を知らない engine が manifest を黙って受け入れ、利用者は適用形が効いていないことに
気付けない。単体 marker の形が変わると、marker を exact 比較する既存の利用側が壊れる。

## 評価

- likelihood: low — 前置は `normalizeManifest` の 1 箇所で全 entry へ一様に掛かり、
  正規化は既存の成分分解を流用する。root 側は `rootKind` だけを写す既存の写像のまま変えない
- impact: medium — 配置先がずれても root の外へは出ない（root 外への脱出は拒否ゲートが
  止める）。ずれた配置は layat の記録に残るため、stale 除去で原状回復できる

## 対処

TC-7d713505-5673-4e15-99bf-e98965e18892（全 target への正規化済みの前置・root と単体 marker の不変・
不正な `subdir` の評価時拒否）で緩和する。

## 出典

ADR-0040 の決定 §1 / §2 と、`tests/nix-unit/root-subdir.nix` /
`tests/namaka/manifest-subdir/expr.nix` の実装。
