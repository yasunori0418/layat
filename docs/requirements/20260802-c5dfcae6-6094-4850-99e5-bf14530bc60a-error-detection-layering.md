---
id: "REQ-c5dfcae6-6094-4850-99e5-bf14530bc60a"
type: requirement
name: "設定の誤りは評価時に、実体の不整合は engine 実行時に検出する層分けを守る"
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  Errors SHALL be detected in the layer that can decide them. An error that is decidable
  from the configuration alone (an omitted `root`, an ill-typed value, an unknown key, a
  contradictory combination, an absolute path or a `..` escape) SHALL be detected at Nix
  evaluation time by `evalModules` / `lib.throwIf` inside `mkManifest`, so that the engine
  is never started. An error that is only decidable against the actual filesystem SHALL be
  detected by the engine at run time. A property that is guaranteed natively by the data
  model SHALL fall into neither layer, no check for it being placed at all. This item
  states the layering alone; the conditions named above are given as instances of it, and
  each condition, its behaviour, and the native guarantee of identifier uniqueness are
  stated by the items responsible for them and SHALL NOT be restated here.
specification_ja: |
  エラーは、それを判定できる層で検出しなければならない。設定だけで判定できるもの
  （`root` 省略・型不正・未知キー・矛盾する組み合わせ・絶対パス / `..` エスケープ）は
  `mkManifest` の `evalModules` / `lib.throwIf` により Nix 評価時に検出し、engine を
  起動させてはならない。実際のファイルシステムと突き合わせて初めて判定できるものは
  engine 実行時に検出しなければならない。データモデルによって native に担保される性質は
  どちらの層にも載せてはならず、検査そのものを置いてはならない。本 item は層分けのみを
  規定する。上に挙げた条件は
  その事例として示すもので、個々の条件とその動作、および識別子の一意性が native に
  担保されること自体は各担当 item の規範であって、ここでは再掲しない。
---
# REQ-c5dfcae6-6094-4850-99e5-bf14530bc60a: 設定の誤りは評価時に、実体の不整合は engine 実行時に検出する層分けを守る

## 仕様

評価時エラー（`root` 省略・不正な型・未知キー・copy+marker・絶対パス / `..` エスケープ）は
`mkManifest` の `evalModules` / `lib.throwIf` が検出する。識別子の一意性は entries が
attrset のため Nix のキー重複不可で担保され、**重複 `name` という評価時エラーは存在しない**。

判定に実体（ファイルシステム）を要する条件は評価時には決まらないため、engine 実行時の
検出に回る。

> **上は原文の写しで、規範は frontmatter が正**。原文が列挙する個々の条件は、それぞれの
> 担当 item が規範を持つ。
>
> - `root` の明示必須 → REQ-4ec3accc-8bb6-461f-9024-dcf0027849e4
> - 未知キー・旧名の strict 拒否 → REQ-3e446ad9-a6f4-4229-b5c5-184754b0ef51
> - `src` が素の文字列を取らないこと → REQ-99ca5381-6c53-426c-b145-7b4297c53868
> - 絶対パス / `..` エスケープの拒否 → REQ-6911eab6-12b4-457c-9db4-d7430a9e9b3f
> - copy + out-of-store marker の矛盾拒否 → REQ-16faf428-77f3-492f-b858-222c5274cbf7
> - 識別子 = 属性キーと一意性の native 担保 → REQ-cb77ea05-bab8-4ccf-b09e-d23d8f71cdc7
> - 同一 manifest 内の正規化後 target 重複 → REQ-5c6b07da-3d06-414d-8770-4f438234b322
> - `mkManifest` が入力の確認を担う単一の条件であること → REQ-d1b5b3f5-10a0-400d-9f03-ba00c63d1c34
>
> `root = systemRoot`（system mode）は `rootKind = "system"` として評価を通り、評価時エラーには
> ならない（→ ADR-9bdf2604-c7b3-468a-b1e4-2478ef37477c）。

## 出典

`docs/spec.md`「エラー仕様」節の導入文と、同節の表のうち評価時エラー行・engine 実行時
エラー行の別。

決定の実体は ADR-2e454c1f-6cad-41d8-8311-2c8e5b8efdf6「mkManifest の入力を evalModules で検査する」（設定の
誤りを `mkManifest` の評価時に閉じる）と、native に担保される性質へ確認を置かない根拠と
なる ADR-c494e86e-ceb9-4380-b983-4e09553b01d6「entries を target をキーにした attrset にする」。

実体依存の判定を engine 実行時へ回すことは、ADR-5e46ebd6-e44b-48dc-a93f-a8a3a3903967「copy と out-of-store の併用を eval 時に拒否する」などの個別の決定に現れる層の分かれ方を、節の導入文が
一般則としてまとめたものである。この一般則そのものを表明した ADR は無いため、これらの ADR からは
側面の根拠として `justifies` は張らない（個別の決定の帰属は REQ-16faf428-77f3-492f-b858-222c5274cbf7 が担当する）。
