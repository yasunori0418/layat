---
id: "RISK-09df40d3-2752-433e-9ab0-2816fbd14969"
type: risk
name: "評価時に止めるべき入力が止まらずに engine 実行時まで到達する"
status: stable
likelihood: low
impact: high
level: medium
threatens:
  - "REQ-6911eab6-12b4-457c-9db4-d7430a9e9b3f"
  - "REQ-d1b5b3f5-10a0-400d-9f03-ba00c63d1c34"
  - "REQ-3e446ad9-a6f4-4229-b5c5-184754b0ef51"
  - "REQ-a33a11e3-830d-4142-88ed-4c1fc35e7f74"
  - "REQ-16faf428-77f3-492f-b858-222c5274cbf7"
  - "REQ-5c6b07da-3d06-414d-8770-4f438234b322"
  - "REQ-99ca5381-6c53-426c-b145-7b4297c53868"
  - "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
  - "REQ-08c8f639-790c-4970-a954-c9567538cf9b"
  - "REQ-4ec3accc-8bb6-461f-9024-dcf0027849e4"
  - "REQ-97c1e088-a17e-46d9-a9a1-83d1757d0f7d"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-05T00:26:05+09:00"
---
# RISK-09df40d3-2752-433e-9ab0-2816fbd14969: 評価時に止めるべき入力が止まらずに engine 実行時まで到達する

## リスク

`mkManifest` は `evalModules` で入力を確認する単一の箇所（[mkManifest 自身が evalModules で入力を確認する単一の箇所になる](../requirements/mkmanifest-input-validation.md)）で、パス安全性
（[target / subpath のパス安全性を評価時に確認する](../requirements/path-safety-validation.md)）・未知キーと旧名の拒否（[entry submodule は strict とし未知キーと旧名を評価時エラーにする](../requirements/entry-submodule-strict.md)）・entry のフィールドを 4 つに限ること
（[entry submodule のフィールドは src / subpath / target / method の 4 つとする](../requirements/entry-submodule-fields.md)）・意図が矛盾する組み合わせ（[意図が矛盾する組み合わせをクロスフィールドの確認で評価時に拒否する](../requirements/normalize-manifest-cross-field-checks.md)）・同一 manifest 内の target 衝突
（[target 衝突の検出処理を同一 manifest 内と cross-config で分ける](../requirements/target-collision-eval-detection.md)）・素の文字列 src の拒否（[src は path / set / marker の 3 種を取り store link を既定として out-of-store は marker で opt-in する](../requirements/src-kinds-store-default.md)）を全てここで止める。この確認は
`modules/common.nix` が共有する entriesType にも適用される（[全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](../requirements/module-common-options.md)）。
root マーカーの subdir 適用形（[root マーカーの subdir 適用形は全 target への前置と等価な評価時の糖衣で、root を動かさない](../requirements/root-marker-subdir-sugar.md)）の拒否（絶対パス・`..` 脱出・空の
`subdir`、前置で target の `..` 脱出を隠さないこと）も同じ条件で止める。

`root = systemRoot`（system mode）は `rootKind = "system"` として評価を通り、評価時に止める
入力ではない（→ [rootKind に system を足しても schemaVersion を上げない](../adr/system-rootkind-without-version-bump.md)、[root は 3 マーカーと絶対パス文字列の union を取る](../requirements/root-marker-kinds.md)）。root 省略時のエラー（[root は明示必須で暗黙デフォルトを持たない](../requirements/root-explicit-required.md) の明示必須）と、`pkgs` などの必須引数を
省いた呼び出しのエラー（[mkManifest の引数は必須の pkgs / entries / root と任意の identity とする](../requirements/mkmanifest-arguments.md)）も、評価時に止める入力に当たる。
どちらも型定義の条件ではなく関数の必須引数として止まり、省略を見るアサートは評価テストに存在しない。
この 2 本へは緩和の無い残余として `threatens` を張る（[mkManifest の引数は必須の pkgs / entries / root と任意の identity とする](../requirements/mkmanifest-arguments.md) のうち
`identity` の扱いは [manifest.json の形が engine の読む契約から気付かれずにずれる](manifest-contract-drift.md) 側の [manifest 文書の形（schemaVersion / root / entry 5 フィールド）を不変条件としてアサートする](../test/manifest-eval/manifest-shape-invariants.md) が見る）。

**顕在化したときに起きること**: この条件が緩むと、`..` や絶対パスで root の外を指す target が
engine まで届き、意図しない場所を書き換える。タイポした旧名は利用者に知らせずに無視され、ユーザーは
宣言したはずの entry が配置されない理由を追えない。target 衝突は最後に評価された側が勝ち、
どちらが配置されたか宣言からは読み取れない。いずれも「評価時に止まる」ことが安全の前提に
なっており、実行時まで到達した時点で被害は FS に出る。

## 評価

- likelihood: low — 条件は型定義とクロスフィールドの確認の 1 箇所に集約されており、
  境界条件（`..` の深さ・0 ちょうど）まで含めて nix-unit が対で覆っている。本 risk が
  扱う「止めるべき入力が止まらずに通る」方向、すなわち throw しなくなる退行はこの確認が
  確実に捕まえる（どの条件が throw したかの取り違えは [nix-unit: gates.nix — 評価時に throw する拒否条件群](../test/manifest-eval/nix-unit-gates.md) の部分一致では
  区別できないが、それは止まらずに通る失敗とは別の失敗モードで [評価スイート自身の盲点でアサートが空虚に緑になる](eval-suite-blind-spot.md) の担当）。
  ただし root と必須引数の省略はこの確認の外にあり、上の残余に当たる
- impact: high — パス安全性の破れ、すなわち `..` や絶対パスが root の外を指したまま engine へ
  届く場面では、書き込み先が layat の記録に残らず原状回復の対象にすらならない。ここで継承則を
  援用する必要は無い。本 item が扱う確認の箇所は `lib/manifest.nix` の `mkManifest` で、`lib` として公開
  される出荷物そのものである。テストスイートのように開発内へ閉じないため、high 行が直接
  当たる（継承則は開発内に閉じる条件が low 行に取られるのを防ぐための例外で、本 item は
  その例外を必要としない）

## 対処

[拒否すべき入力が評価時に throw することをアサートする](../test/manifest-eval/rejection-gates.md)（各条件が評価時に `ThrownError` を投げること。主な処理は `normalizeManifest`
だが、型共有の 1 件は `entriesType` を直接評価する）・[パス脱出判定の深さ境界を private helper 越しに内側・外側の対でアサートする](../test/manifest-eval/path-escape-boundary.md)（private helper を単体で
叩いてパス脱出判定の深さ 0 境界を内側・外側の対で見ること）・[subdir 適用形が全 target に正規化済みの前置を施し、root と単体 marker を変えず、不正な subdir を評価時に拒否する](../test/manifest-eval/root-subdir-sugar.md)（subdir 適用形の
拒否の条件）で緩和する。1 つ目はエラーの到達性、2 つ目は判定ロジック単体、3 つ目は subdir
適用形が足した拒否の条件という軸の違いで 3 つの TC に分ける。

## 出典

`tests/nix-unit/gates.nix` / `tests/nix-unit/escapes-base.nix` の現行実装からの逆算
（→ Issue #273「L1〜L4」節）。各条件の設計判断は
[配置元の中のパスを全層で subpath と呼ぶ](../adr/subpath-naming-across-layers.md) / [entry の未知のキーを評価時に拒否する](../adr/strict-unknown-keys-rejected.md) / [src の型で素の文字列を拒否する](../adr/srctype-bare-string-rejected.md) /
[copy と out-of-store の併用を eval 時に拒否する](../adr/copy-with-out-of-store-eval-error.md) / [絶対パスと .. エスケープは eval 時に拒否する](../adr/path-safety-eval-rejection.md) / [manifest 内の target の衝突を評価時に検出する](../adr/manifest-target-conflict-eval-error.md) が持つ。
