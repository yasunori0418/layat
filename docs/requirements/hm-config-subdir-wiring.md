---
id: "REQ-068f2fb3-5459-4325-9359-c931b5049e0d"
type: requirement
name: "HM モジュールは `configs.<name>.subdir` が非 null の config を homeRoot の subdir 適用形で組み、全 target を `$HOME/<subdir>/` 配下に置く"
description: "HM モジュールは subdir が非 null の config の manifest を homeRoot { subdir } で組み、subdir の確認は manifest の評価に委ね、root オプションを公開してはならない"
section: modules
status: stable
derives_from:
  - "UC-d39c1994-f9a5-4860-80ba-f6e584adaf14"
specification: |
  Where `layat.configs.<name>.subdir` is non-null, the home-manager module SHALL build the
  manifest of that config with the root `homeRoot { subdir = <subdir>; }`, so that every
  target of that config is placed under `$HOME/<subdir>/` while the root stays pinned to
  home. Where it is null, the module SHALL build the manifest with the plain `homeRoot`, and
  the subdir of one config SHALL NOT affect the targets of any other config. The module
  SHALL NOT validate, normalize or reinterpret the subdir itself — the empty string
  included — and SHALL leave rejecting an absolute path, an escape through `..` or an empty
  subdir to the evaluation of the manifest. The module SHALL NOT expose a `root` option.
specification_ja: |
  `layat.configs.<name>.subdir` が非 null のとき、home-manager モジュールはその config の
  manifest を root `homeRoot { subdir = <subdir>; }` で組まなければならない。これにより root を
  home に pin したまま、その config の全 target が `$HOME/<subdir>/` 配下に置かれる。null の
  ときは素の `homeRoot` で manifest を組まなければならず、ある config の subdir が他の config の
  target に影響してはならない。モジュールは subdir そのものを確認・正規化・読み替えしては
  ならず（空文字も含む）、絶対パス・`..` による脱出・空の subdir の拒否は manifest の評価に
  委ねなければならない。モジュールは `root` オプションを公開してはならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:29:58+09:00"
---
# HM モジュールは `configs.<name>.subdir` が非 null の config を homeRoot の subdir 適用形で組み、全 target を `$HOME/<subdir>/` 配下に置く

## 仕様

- `configs.<name>.subdir` が非 null の config は `homeRoot { subdir = <subdir>; }` を root にして
  manifest を組む。root は home のままで、全 target が `$HOME/<subdir>/` 配下になる
- null の config は素の `homeRoot` で組む。subdir はその config の target にだけ適用され、他の config の
  target は変わらない
- モジュールは subdir を確認・正規化しない。不正な subdir は manifest の評価で eval エラーになる
- `root` オプションは公開しない

> **他 item との担当分界**: オプションの型とデフォルト値（`nullOr str`・null）は
> [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](module-common-options.md)、subdir 適用形が target へどう前置され、どの値を
> 拒否するかは [root マーカーの subdir 適用形は全 target への前置と等価な評価時の糖衣で、root を動かさない](root-marker-subdir-sugar.md)、config ごとの
> profile の独立は [HM モジュール経由でも名前つき config ごとに役割分離した独立 profile を取れる](hm-named-configs-profiles.md) の担当。

## 出典

[`configs.<name>.subdir` で config ごとに target の基点を寄せる](../adr/module-config-subdir-option.md)・[モジュールは subdir の確認を lib に委ねる](../adr/module-delegates-subdir-validation.md)。subdir 適用形そのものは [root マーカーに subdir の適用形を持たせる](../adr/root-marker-subdir-applied-form.md) が定めている。
