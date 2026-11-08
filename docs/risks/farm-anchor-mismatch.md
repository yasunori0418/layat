---
id: "RISK-f67a0883-950a-4458-8b6d-1f95cb039cb1"
type: risk
name: "GC アンカーの対象と名前を取り違えて世代が回収不能になる"
status: stable
likelihood: medium
impact: high
level: high
threatens:
  - "REQ-62eda895-efd4-4eaf-a58b-600e8637da75"
  - "REQ-b12fc3c0-d7fe-4003-922c-f3ac0d969b66"
  - "REQ-60e6b49c-9ba1-4552-a0ec-d340421ec281"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:28:26+09:00"
---
# RISK-f67a0883-950a-4458-8b6d-1f95cb039cb1: GC アンカーの対象と名前を取り違えて世代が回収不能になる

## リスク

GC アンカーの対象と名前は、それぞれ [symlink farm は GC アンカー専用でアンカーは store-backed な symlink entry に限る](../requirements/farm-gc-anchor-only.md) と [symlink farm の GC アンカー名は target のハッシュとする](../requirements/farm-anchor-name-hash.md) が定める。farm derivation は
`mkManifest` の返り値に含まれる（[mkManifest は manifest.json と symlink farm を含む store オブジェクトを返す](../requirements/mkmanifest-return-value.md)）。

**顕在化したときに起きること**: 抽出条件が緩めば copy や out-of-store の entry までアンカーを
持ち、世代外・置き切りであるはずの配置が store の生存に紐づく。逆に厳しすぎれば store-backed
な symlink がアンカーを失い、世代が積まれた後に前世代の store パスが GC で回収されて
ロールバック先が破損する。アンカー名が非決定的になれば、同じ target が世代ごとに別名のアンカーを
持ち、間引きの対象を特定できなくなる。いずれも配置直後には観測されず、GC が実行された後に
初めて破損した形で現れる。

**この破れが `requirement` に張る理由**: 検証手段を変えても、GC アンカーの取り違えという破れ
自体は残る。`docs/agents/sara-graph.md` の判別規約に従い `requirement` 側へ張る。

## 評価

- likelihood: medium — 抽出条件は method × srcKind の組み合わせで分岐し、条件の追加・変更で
  取りこぼしが起こりうる
- impact: high — アンカーの欠落は前世代の store パス回収 = ロールバック不能を招く

## 対処

[GC アンカー対象の抽出条件と farm の組み立て結果をアサートする](../test/manifest-eval/farm-anchor-derivation.md)（アンカー対象の抽出と farm の組み立て）・[GC アンカー名の形式・決定性・特殊文字耐性をアサートする](../test/manifest-eval/anchor-name-determinism.md)（アンカー名の形式と決定性）で
緩和する。

ただし本区分が見るのは抽出条件とアンカー名、および derivation のビルドスクリプトへの配線
までで、**[mkManifest は manifest.json と symlink farm を含む store オブジェクトを返す](../requirements/mkmanifest-return-value.md) の「返り値が symlink farm を含む store オブジェクトである」ことそのもの
は残余**として残る。[GC アンカー対象の抽出条件と farm の組み立て結果をアサートする](../test/manifest-eval/farm-anchor-derivation.md) は fake な pkgs 越しにスクリプト本文を取り出すだけで実ビルドは
行わないため、配置後の farm derivation の実体は評価テストの射程外にあり、`integration` 区分の
E2E が担当する（→ Issue #289）。

## 出典

`tests/nix-unit/farm-entries.nix` / `tests/nix-unit/anchor-name.nix` の現行実装からの逆算
（→ Issue #273「L1〜L4」節）。GC アンカーの設計判断そのものは [symlink farm を GC アンカー専用にする](../adr/farm-as-gc-anchor-only.md) が持つ。
