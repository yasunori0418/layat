---
id: "RISK-5df2d02b-e5d4-40eb-86ad-e8bc96e4c34d"
type: risk
name: "manifest.json の形が engine の読む契約から気付かれずにずれる"
status: stable
likelihood: medium
impact: medium
level: medium
threatens:
  - "REQ-dedd2c28-bba3-4ecf-80c9-8c77347e8e1f"
  - "REQ-dd10d820-e453-4099-a47a-ffb9a7de02fb"
  - "REQ-0b0cd1e3-bfeb-45c1-978d-e2e11c568336"
  - "REQ-79ce0a09-f9bd-4e61-ba7f-45fb5643137b"
  - "REQ-250d936c-1df0-491d-a7af-708f38b61f33"
  - "REQ-b232ec98-af3b-41f3-a050-29d417322002"
  - "REQ-1dcc9a33-b0f2-43e0-8310-fc4b19e68fe7"
  - "REQ-77689c68-953c-4cbb-ab31-1ac1e4f5f2fe"
  - "REQ-eb363122-385a-499c-a074-c95efb949d07"
  - "REQ-97c1e088-a17e-46d9-a9a1-83d1757d0f7d"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:42:17+09:00"
---
# RISK-5df2d02b-e5d4-40eb-86ad-e8bc96e4c34d: manifest.json の形が engine の読む契約から気付かれずにずれる

## リスク

`manifest.json` の位置づけは [manifest.json が Nix と engine の唯一の安定契約であり schemaVersion は 1 に固定する](../requirements/manifest-single-stable-contract.md) と [MVP は manifest v1 のみを発行・受理しマイグレーション機構を持たない](../requirements/manifest-v1-only-no-migration.md) が定める。`normalizeManifest` が出す文書の形のうち、
トップレベルは [manifest.json のトップレベルは必須の schemaVersion / root / entries と任意の identity とする](../requirements/manifest-toplevel-fields.md)、`root` は [manifest.json の root は rootKind を持ち fixed のときだけ絶対パスを併記する](../requirements/manifest-root-object.md) が定める。
entry と辞書順は [manifest.json の entries は attrset を配列へ正規化し 5 フィールドを記録する](../requirements/manifest-entries-array.md)、marker の判別タグの扱いは [marker は判別タグで識別し manifest.json には漏らさない](../requirements/marker-tag-discrimination.md) が定める。
この形が、評価層の変更で気づかれずに変わりうる。

**顕在化したときに起きること**: engine 側は不正な文書を読んで失敗するか、より悪いことに
フィールドの欠落や順序変化を利用者に知らせずに受理し、意図しない配置を行う。`mkManifest` は
FS も nix daemon も介さない（[mkManifest は配置データを生成する純粋関数である](../requirements/mkmanifest-pure-function.md)）ため、この破れは実行するまで観測されず、
E2E まで落ちてこないと検出できない（純粋性そのものが脅かされるわけでは
ないので `threatens` は張らない。検出が難しい理由としてここに書くに留める）。`method` の既定や defaults の適用が変われば（規範は [method は配置方法を選び symlink は世代管理下・copy は世代管理外になる](../requirements/method-src-matrix.md) と [normalizeManifest が確認・デフォルト適用・marker 変換を行い mkManifest が derivation を組む](../requirements/normalize-manifest-defaults.md)）、世代管理下に置くつもりの entry が
世代外へ回る。`mkOutOfStoreSymlink` のマーカーの形が
変われば、manifest の `srcKind` と local path の写しがずれ、engine は out-of-store の entry を別の
種類として読む（マーカーの規範は [mkOutOfStoreSymlink は out-of-store symlink を表すマーカーを返す](../requirements/mkoutofstoresymlink-marker.md)）。`identity` を null のまま渡したときに manifest へ `identity` を出す
（→ [mkManifest の引数は必須の pkgs / entries / root と任意の identity とする](../requirements/mkmanifest-arguments.md)）のも、文書の形のずれに当たる。

## 評価

- likelihood: medium — 契約は 1 箇所（`lib/manifest.nix`）に閉じているが、フィールド追加・
  既定値変更は評価層の通常の改修で起こりうる
- impact: medium — engine の全処理が読む唯一の契約で、破れは配置結果そのものを狂わせる。
  ただし狂うのは layat が管理する配置であり、`lib/manifest.nix` を直したうえでの再 apply が
  正しい形へ収束させる。破れが実行するまで観測されない（上記のとおり `mkManifest` は純粋関数で
  E2E まで落ちないと気づけない）ことは沈黙性の問題で、回復可能性は下げない。
  **root の種別を取り違えて配置先そのものが別ツリーへ移る facet は本 item の射程ではない**。
  受理の段で止める側は [Nix と engine の唯一の契約である manifest の生成側と消費側がずれる](manifest-contract-skew.md) が、engine の解決規則そのものは [root 解決を誤り、意図しないディレクトリツリーへ配置する](root-resolution-wrong-tree.md) が持ち、
  どちらも「記録に残らない場所へ書くので回収できない」ことを理由に high を採っている。本 item が
  持つのは生成側が出す文書の形であり、下の残余第一（`fixed` の相対パスを受理する不足）も、
  拒否する規範が REQ 側に無いという仕様の未決事項として置いたもので、impact はその 2 item 側で
  計上する

## 対処

[manifest 文書の形（schemaVersion / root / entry 5 フィールド）を不変条件としてアサートする](../test/manifest-eval/manifest-shape-invariants.md)（文書構造の不変条件）・[fixed root で rootKind が fixed になり絶対パスを併記する（home marker が倒れないことを含む）](../test/manifest-eval/fixed-root-absolute-path.md)（fixed root の絶対パス併記）・[省略時の既定適用・明示上書き・target 辞書順での決定的配列化をアサートする](../test/manifest-eval/entry-normalization-defaults.md)
（既定適用と決定的順序）・[marker の判別と src 種別への変換を単一 entry の境界でアサートする](../test/manifest-eval/marker-resolution.md)（marker の判別と src 種別への変換）・[manifest 文書全体をスナップショットに固定し、どの不変条件も名指ししない変更を差分として出す](../test/manifest-eval/manifest-document-regression.md)
（文書全体のスナップショット回帰）で緩和する。

[manifest.json の root は rootKind を持ち fixed のときだけ絶対パスを併記する](../requirements/manifest-root-object.md) は project root の否定側を [manifest 文書の形（schemaVersion / root / entry 5 フィールド）を不変条件としてアサートする](../test/manifest-eval/manifest-shape-invariants.md)、fixed root の肯定側を [fixed root で rootKind が fixed になり絶対パスを併記する（home marker が倒れないことを含む）](../test/manifest-eval/fixed-root-absolute-path.md) が
見るので、評価層の分は両側が揃っている（→ Issue #288）。

**本区分の TC 群では緩和しきれない残余**が 5 つある。

第一に、`fixed` の `root` が実際に絶対パスであることを強制する確認がそもそも実装に無い。
`rootType` は marker でない文字列をすべて受理し（`lib.isString` のみ）、`normalizeManifest` は
それを確認せずに `rootKind = "fixed"` へ倒すため、相対パス文字列も fixed として通る。
[fixed root で rootKind が fixed になり絶対パスを併記する（home marker が倒れないことを含む）](../test/manifest-eval/fixed-root-absolute-path.md) は常に絶対パスを渡すので、この不足は評価テストでも補えない。拒否する規範が
REQ 側に無い（[manifest.json の root は rootKind を持ち fixed のときだけ絶対パスを併記する](../requirements/manifest-root-object.md) / [root は 3 マーカーと絶対パス文字列の union を取る](../requirements/root-marker-kinds.md) は「絶対パス」と述べるが拒否の条件を要求していない）
ので、仕様の未決事項として残余に置く。

第二に、`homeRoot` を HM 統合の配線として確かめる側は評価層の担当外で、`checks.hm-module`
（`integration` 区分）が持つ。ただし [manifest.json の root は rootKind を持ち fixed のときだけ絶対パスを併記する](../requirements/manifest-root-object.md) の home 側そのもの——`rootKind = "home"` で
`root` フィールドを持たないこと——は [fixed root で rootKind が fixed になり絶対パスを併記する（home marker が倒れないことを含む）](../test/manifest-eval/fixed-root-absolute-path.md) が exact 一致で担保済みで、
`checks.hm-module` の grep（`"rootKind":"home"` の部分一致）は余分なフィールドの混入を
検出しない。`integration` 区分が見るのは「HM モジュールが root に home を pin する配線」であり、
文書の形そのものではない。

第三に、`rootKind = "system"` は [manifest.json の root は rootKind を持ち fixed のときだけ絶対パスを併記する](../requirements/manifest-root-object.md) が定める 4 値の 1 つで、`normalizeManifest` は
`systemRoot` を `{ rootKind = "system"; }` へ変換して通す（→ [rootKind に system を足しても schemaVersion を上げない](../adr/system-rootkind-without-version-bump.md)）。この発行は
[nix-unit: structure.nix — manifest 構造の不変条件](../test/manifest-eval/nix-unit-structure.md) の変換テストがアサートするが、文書の形を条件とする [manifest 文書の形（schemaVersion / root / entry 5 フィールド）を不変条件としてアサートする](../test/manifest-eval/manifest-shape-invariants.md) の適用範囲は
project root に限られ、system の変換を受ける TC はまだ無い。`system` は規範側に所在を持つが、
条件側には持たない。

第四に、`fixed` の絶対パスが `mkManifest` の `passthru.root`（[mkManifest の返り値は passthru で root kind を露出する](../requirements/mkmanifest-passthru-rootkind.md)）へ写る側は
derivation 層の担当で、本区分の TC 群は見ない（[fixed root で rootKind が fixed になり絶対パスを併記する（home marker が倒れないことを含む）](../test/manifest-eval/fixed-root-absolute-path.md) が射程外と宣言している）。この
分岐は fixed のときだけ通り、CLI が build 前に `nix eval` で読む処理（`cli-json` 区分）の
入力になるが、現状どの risk も残余として引き取っておらず、CASE も無い。評価層が出す文書の
形とは別物なので本 risk の射程には入れず、`cli-json` 区分側で扱うべき不足としてここに記録する
に留める。解消は、同区分に passthru を threatens する risk を立てて残余を移送した時点。
それまでは本 risk が仮の置き場になる。

第五に、[MVP は manifest v1 のみを発行・受理しマイグレーション機構を持たない](../requirements/manifest-v1-only-no-migration.md) / [manifest.json が Nix と engine の唯一の安定契約であり schemaVersion は 1 に固定する](../requirements/manifest-single-stable-contract.md) は発行側と engine の受理側の双方を規範に含むが、本 risk が
射程に持つのは評価層が発行する側だけで、engine が新しい `schemaVersion` を拒否する側は
`cli-json` / `integration` 区分の担当になる。

## 出典

`tests/nix-unit/structure.nix` / `defaults.nix` / `resolve-marker.nix` と
`tests/namaka/manifest-project/expr.nix` の現行実装からの逆算（→ Issue #273「L1〜L4」節）。
manifest 契約の設計判断は [manifest.json へ marker を clean enum で写す](../adr/wire-format-clean-enum.md) / [manifest.json の entry は target を識別子にして name を持たない](../adr/manifest-entry-identity-is-target.md) が持つ。
