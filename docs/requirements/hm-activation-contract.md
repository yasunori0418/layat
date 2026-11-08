---
id: "REQ-8085f194-c903-4ecb-abd8-c719fe7b3292"
type: requirement
name: "home-manager モジュールの engine kick は activation から全 config のビルド済み link-farm を 1 回の起動で渡し、失敗で switch を止める"
status: stable
derives_from:
  - "UC-d39c1994-f9a5-4860-80ba-f6e584adaf14"
specification: |
  Where `layat.configs` holds at least one config, the home-manager module SHALL kick the
  engine from `home.activation.layat`
  (`entryAfter ["writeBoundary"]`) with a single
  `layat apply --all --manifest <name>=<link-farm> …` invocation, each link-farm being built
  by `mkManifest` from the entries of a config at module evaluation time and its store path
  passed from the activation script. The activation SHALL NOT perform `nix eval` or
  `nix build`, not being on the entrypoint path. The flock SHALL be taken blocking, the
  placement report SHALL be visible, and an engine error such as a conflict SHALL stop the
  switch by exiting non-zero, matching the clobber error of `home.file` under a declarative
  switch. Because the pinned layat CLI (`packages.layat`) and `mkManifest` come of the same
  flake input, a schemaVersion skew SHALL NOT arise structurally. This states the contract
  of the kick; the correspondence of one config to one manifest and one profile is stated
  by REQ-c6891aeb-13c0-4ae7-9ad1-5c343735266a, and the definition of the option supplying
  its entries by REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10, neither being restated here. The
  order of the `<name>=<link-farm>` pairs, `--jobs`, how a partial failure is aggregated,
  and not invoking `layat` when no config is present are stated by REQ-c847d1af-a437-46bb-bd64-42083810d034 and are likewise not
  restated here.
specification_ja: |
  `layat.configs` に config が 1 つ以上あるとき、home-manager モジュールは
  `home.activation.layat`（`entryAfter ["writeBoundary"]`）から
  `layat apply --all --manifest <name>=<link-farm> …` の 1 回の起動で engine を kick しなければ
  ならない。各 link-farm はモジュール評価時に config の entries から `mkManifest` でビルドし、
  その store パスを activation script から渡す。activation は entrypoint を通る処理ではないため
  `nix eval` / `nix build` を行ってはならない。flock は blocking で取り、配置レポートを可視とし、
  engine error（conflict 等）は非 0 終了で switch を止めなければならない（宣言的 switch・
  `home.file` の clobber エラーと同型）。pin 版 layat CLI（`packages.layat`）と `mkManifest` が
  同一 flake input 由来のため、schemaVersion skew は構造的に起きてはならない。本 item が
  規定するのは起動の契約であり、1 config が 1 manifest = 1 profile に対応することは
  REQ-c6891aeb-13c0-4ae7-9ad1-5c343735266a、その entries を供給するオプションの定義は
  REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10 の担当で、いずれも本 item では規定しない。
  `<name>=<link-farm>` の並び・`--jobs`・部分失敗の集約・config が無いときに `layat` を起動しないことは
  REQ-c847d1af-a437-46bb-bd64-42083810d034 の担当で、これも本 item では規定しない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:31:01+09:00"
---
# REQ-8085f194-c903-4ecb-abd8-c719fe7b3292: home-manager モジュールの engine kick は activation から全 config のビルド済み link-farm を 1 回の起動で渡し、失敗で switch を止める

## 仕様

| method | `src` | 動作 |
|---|---|---|
| `"symlink"` | path / set | engine が store link をネイティブ symlink |
| `"symlink"` | marker | engine が out-of-store symlink をネイティブ作成（HM の mkOutOfStoreSymlink には委譲しない）|
| `"copy"` | path / set | engine が place-once ネイティブコピー |

- `home.activation.layat`（`entryAfter ["writeBoundary"]`）から engine を起動する。配置ロジックは
  HM に依存しない。root は `homeRoot` を pin
- engine kick は **`layat apply --manifest <link-farm>`**。モジュール評価時に `layat.entries` から
  `mkManifest` で link-farm をビルドし、その store パスを activation script から渡す。
  activation は `nix eval` / `build` を行わない（entrypoint 経路ではない）
- **blocking flock・配置レポート可視・engine error（conflict 等）は非 0 終了で switch を止める**
  （宣言的 switch・`home.file` の clobber エラーと同型）
- pin 版 layat CLI（`packages.layat`）と `mkManifest` が同一 flake input 由来のため
  schemaVersion skew は構造的に起きない
- 世代は layat 自前 profile（内部機構・前世代マニフェスト + stale 追跡）に乗る。ユーザー向け
  rollback は host（`home-manager --rollback`）に一本化（`layat rollback` 非公開）

> **上は原文の写しで、規範は frontmatter が正**。原文が参照する次の規範は本 item の担当では
> ない。
>
> - method と `src` の組み合わせが選ぶ配置方法 → [method は配置方法を選び symlink は世代管理下・copy は世代管理外になる](method-src-matrix.md)。表が HM でも engine が扱う
>   （HM の `mkOutOfStoreSymlink` へ委譲しない）と述べる点は [mkOutOfStoreSymlink は out-of-store symlink を表すマーカーを返す](mkoutofstoresymlink-marker.md) / [全モジュールと devShell は engine をキックするだけの配線とし、ネイティブ機構へ翻訳しない](modules-are-engine-wiring.md)
> - `apply --manifest` の契約そのもの（entrypoint 発見・eval・build を行わないこと・
>   `-f` との併用エラー・`--all` と合成するときの `<name>=<link-farm>` の繰り返し）→ [apply --manifest はビルド済み link-farm を engine へ直接適用し、--all と合成するときは name=link-farm を繰り返す](apply-manifest-direct.md)
> - モジュールが root を pin し利用者が再指定しないこと → [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](module-common-options.md)。`homeRoot` の
>   層ごとの解決 → [home mode の root は層ごとに定まった供給元から解決する](home-mode-root-resolution.md)
> - manifest を供給するオプションの定義（canonical は `layat.configs.<name>.entries`・
>   `layat.entries` はその糖衣）→ [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](module-common-options.md)、1 config が 1 manifest = 1 profile に対応する
>   ことと profile 粒度 → [HM モジュール経由でも名前つき config ごとに役割分離した独立 profile を取れる](hm-named-configs-profiles.md)。本 item の写しにある `layat.entries` は原文逐語であり、
>   規範文では供給元を「config の entries」と述べて [HM モジュールは `layat.configs.<name>` ごとに独立した profile を持つ](../adr/hm-configs-independent-profiles.md) と衝突しない形にしている。
>   activation が全 config を 1 回の `apply --all` 起動へ渡す規律（[モジュールの activation は apply --all --manifest を 1 回だけ起動する](../adr/module-activation-single-apply-all.md)・`<name>` の
>   辞書順・`--jobs`・部分失敗の集約）は原文に対応記述が無いため本 item の写しには現れないが、
>   規範は [HM の activation は全 config の link-farm を apply --all 1 回の起動に渡し、報告は辞書順・配置の並列度は --jobs に従う](hm-activation-per-config-kick.md) が持つ
> - flock を既定 blocking で取ること → [flock の取得は既定 blocking とし --no-wait のときだけ try-lock でスキップする](flock-acquisition-mode.md)。レポートと warning を stderr へ出す
>   ストリーム規律 → [stdout は機械可読出力を専有しレポートと warning は stderr へ出す](stream-discipline.md)。終了コードの体系 → [終了コードは 0 = 成功 / 1 = 一般エラー / 2 = dryrun の conflict とする](exit-codes.md)
> - 世代が layat 自前 profile に乗ること → [世代は link farm derivation を layat 自前 profile へコミットして積み、前世代 manifest から古くなった配置を除去する](generation-mechanism-link-farm-manifest.md)。profile 名の次元 → [HM モジュール経由でも名前つき config ごとに役割分離した独立 profile を取れる](hm-named-configs-profiles.md)、
>   profileDir のキー → [profileDir は home と system のみ name 直キーとし、fixed root と --root 上書きは roothash でキーする](profiledir-keying.md)。rollback を host へ一本化すること → [module 時は rollback を host へ一本化し、layat profile は前進のみで追従する](module-mode-profile-internal.md)
> - `manifest.json` が唯一の安定契約で `schemaVersion` が 1 に固定であること →
>   [manifest.json が Nix と engine の唯一の安定契約であり schemaVersion は 1 に固定する](manifest-single-stable-contract.md) / [MVP は manifest v1 のみを発行・受理しマイグレーション機構を持たない](manifest-v1-only-no-migration.md)。本 item はそれを前提に「module の使い方では skew が構造的に
>   起きない」ことだけを規定する
>
> 原文が世代の箇条書きに併記する「MVP は固定名 `default` の単一 profile
> （`<state>/nix/profiles/layat/default`）」は、**[HM モジュールは `layat.configs.<name>` ごとに独立した profile を持つ](../adr/hm-configs-independent-profiles.md) が `layat.configs.<name>` の実装を
> 決定済み**で反転しているため写しから落とした（[HM モジュール経由でも名前つき config ごとに役割分離した独立 profile を取れる](hm-named-configs-profiles.md) / [profileDir は home と system のみ name 直キーとし、fixed root と --root 上書きは roothash でキーする](profiledir-keying.md) と同じ扱い）。
> `docs/spec.md` の追従は本 item の担当範囲外。
>
> **2026-10-15 追記（[モジュールの activation は apply --all --manifest を 1 回だけ起動する](../adr/module-activation-single-apply-all.md)）**: 上の写しの `layat apply --manifest <link-farm>` は、全 config の
> link-farm を渡す `layat apply --all --manifest <name>=<link-farm> …` の 1 回起動と読み替える。
> blocking flock・レポート可視・非 0 終了で switch を止める点は変わらない。

## 出典

`docs/spec.md`「モジュール別動作仕様」→「home-manager モジュール」節の表と箇条書き。

決定の実体は [モジュールの activation は apply --all --manifest を 1 回だけ起動する](../adr/module-activation-single-apply-all.md)と
[モジュールの activation は lock を待ち、失敗したら switch を止める](../adr/module-activation-waits-and-fails-switch.md)。モジュールを配線に徹させることは [モジュールは root と起動の時機だけを供給して engine を起動する](../adr/modules-wire-engine-kick-only.md) が
定めている。全 config を `apply --all` の 1 回の起動で渡す形は [モジュールの activation は apply --all --manifest を 1 回だけ起動する](../adr/module-activation-single-apply-all.md)。
