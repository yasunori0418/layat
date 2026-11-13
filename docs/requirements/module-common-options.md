---
id: "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
type: requirement
name: "全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する"
section: modules
status: stable
derives_from:
  - "UC-d39c1994-f9a5-4860-80ba-f6e584adaf14"
specification: |
  Every module (home-manager / NixOS / nix-darwin) SHALL expose the same set of common
  options: `layat.enable :: bool` defaulting to false, `layat.configs :: attrsOf (submodule
  { entries :: attrsOf entry; subdir :: nullOr str; identity :: nullOr str; })` defaulting to `{}`, with `entries`
  likewise defaulting to `{}` and `subdir` and `identity` defaulting to null, `layat.backup.enable :: bool` defaulting to false,
  `layat.backup.suffix :: str` defaulting to `"layat-backup"`, `layat.recopy :: bool`
  defaulting to false, and `layat.jobs :: ints.unsigned` defaulting to 0, where 0 means the
  logical CPU count. The bare `layat.entries` SHALL
  be kept as a sugar renaming onto `configs.default.entries`, non-breaking and deprecated.
  A module SHALL pass `configs.<name>.identity` unchanged to `mkManifest` as `identity`.
  These SHALL be defined jointly for every module, so that the set does not diverge per
  module. What the `<name>` dimension of `configs` means for the granularity of profiles
  is stated by [HM モジュール経由でも名前つき config ごとに役割分離した独立 profile を取れる](hm-named-configs-profiles.md) and is not restated here. A module SHALL pin the root by its
  own nature — home-manager to `homeRoot`, NixOS and nix-darwin to `systemRoot`, a devShell
  to `projectRoot` — and SHALL NOT
  expose a `root` option, so that a user of a module never restates it. This is a
  departure from `mkManifest` and the CLI entrypoint, where `root` is required to be
  stated explicitly ([root は明示必須で暗黙デフォルトを持たない](root-explicit-required.md)).
specification_ja: |
  全モジュール（home-manager / NixOS / nix-darwin）は共通オプションとして同一の集合を
  公開しなければならない: `layat.enable :: bool`（デフォルト false）・
  `layat.configs :: attrsOf (submodule { entries :: attrsOf entry; subdir :: nullOr str; identity :: nullOr str; })`
  （デフォルト `{}`・`entries` のデフォルトも `{}`・`subdir` と `identity` のデフォルトは null）・`layat.backup.enable :: bool`（デフォルト false）・
  `layat.backup.suffix :: str`（デフォルト `"layat-backup"`）・`layat.recopy :: bool`
  （デフォルト false）・`layat.jobs :: ints.unsigned`（デフォルト 0。0 は論理 CPU 数）。
  素の `layat.entries` は
  `configs.default.entries` への rename 糖衣として非破壊に残さなければならず、deprecated と
  しなければならない。モジュールは `configs.<name>.identity` をそのまま `mkManifest` の
  `identity` へ渡さなければならない。
  これらは全モジュール共通に定義し、モジュールごとに集合が分岐しないようにしなければ
  ならない。`configs` の `<name>` 次元が profile の粒度にとって何を意味するかは
  [HM モジュール経由でも名前つき config ごとに役割分離した独立 profile を取れる](hm-named-configs-profiles.md) の担当で、本 item では規定しない。モジュールは自分の性質で root を pin
  しなければならず（home-manager は `homeRoot`・NixOS / nix-darwin は `systemRoot`・
  devShell は `projectRoot`）、`root`
  オプションを公開してはならない。モジュール利用者が root を再指定しないようにするため
  である。これは `root` を明示必須とする `mkManifest` / CLI entrypoint の層
  （[root は明示必須で暗黙デフォルトを持たない](root-explicit-required.md)）との差分である。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T22:03:12+09:00"
---
# 全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する

## 仕様

```
layat.enable  :: bool          # デフォルト: false
layat.entries :: attrsOf entry # デフォルト: {}（属性キー = target）
                              # layat.configs.<name>.entries の
                              #   deprecated 糖衣へ降格済み（規範は frontmatter・下の注記を参照）
layat.backup.enable :: bool    # デフォルト: false
layat.backup.suffix :: str     # デフォルト: "layat-backup"
```

モジュールは自分の性質で root を pin する（HM → `homeRoot` / NixOS / nix-darwin → `systemRoot` / devShell → `projectRoot`）ため、
モジュール利用者は `root` を再指定しない。

> **上は原文の写しで、規範は frontmatter が正**。`entries` の属性キーが target であることと
> entry submodule のフィールド定義は [entry の識別子は属性キー = target とし一意性は Nix が担保する](entry-identity-by-target-key.md) / [entry submodule のフィールドは src / subpath / target / method の 4 つとする](entry-submodule-fields.md)、`root` が
> `mkManifest` / CLI 層で明示必須であることは [root は明示必須で暗黙デフォルトを持たない](root-explicit-required.md)、`--backup` の退避契約そのものは
> [apply --backup は配置を塞ぐ記録外実体を rename 退避してから配置する](apply-backup-rename-aside.md) の担当。`layat.backup` が manifest ではなく engine 起動の配線に働くことは
> [layat.backup は engine 起動の配線レイヤーのオプションで manifest には影響しない](backup-wiring-layer.md)、`layat.recopy` の配線は [layat.recopy は engine 起動の配線レイヤーのオプションで、true のときだけ activation の起動へ --recopy を付ける](recopy-wiring-layer.md) の担当。`layat.jobs` を activation の `--jobs` へ渡すことは
> [HM の activation は全 config の link-farm を apply --all 1 回の起動に渡し、報告は辞書順・配置の並列度は --jobs に従う](hm-activation-per-config-kick.md) の担当。
>
> **2026-10-15 追記**: 写しの表に無い `layat.recopy` / `layat.jobs` は [layat.recopy で activation の起動に --recopy を付ける](../adr/module-recopy-option.md)・[layat.jobs で activation の並列度を指定する](../adr/module-jobs-option.md) が
> 共通オプションへ足した（規範は frontmatter）。
>
> **2026-10-22 追記**: 写しの表に無い `layat.configs.<name>.subdir` は [`configs.<name>.subdir` で config ごとに target の基点を寄せる](../adr/module-config-subdir-option.md) が
> `configs` submodule へ足した（型とデフォルトの規範は frontmatter）。subdir から root の適用形を
> 組む HM の配線は [HM モジュールは `configs.<name>.subdir` が非 null の config を homeRoot の subdir 適用形で組み、全 target を `$HOME/<subdir>/` 配下に置く](hm-config-subdir-wiring.md) の担当で、root pin の条文は変わらない。
>
> **2026-10-24 追記（[NixOS / nix-darwin モジュールは root を systemRoot に固定する](../adr/system-modules-pin-systemroot.md)）**: root pin の条文に NixOS / nix-darwin → `systemRoot` を足した。両モジュールは
> `layat.user` を持たず、subdir から `systemRoot` の適用形を組む配線は [NixOS / nix-darwin モジュールは activation から全 config のビルド済み link-farm を systemRoot で 1 回の起動に渡す](system-module-activation-contract.md) の担当。
>
> **写しの `layat.entries` を canonical としない理由**: **[HM モジュールは `layat.configs.<name>` ごとに独立した profile を持つ](../adr/hm-configs-independent-profiles.md)・[layat.entries を configs.default.entries への rename 糖衣として残す](../adr/hm-entries-renamed-sugar.md) は
> `layat.configs.<name>.entries` を canonical とし、素の `layat.entries` を
> `configs.default.entries` への deprecated 糖衣へ降格することを決定済み**で、`docs/spec.md`
> の当該表はこれに未追従。[HM モジュールは `layat.configs.<name>` ごとに独立した profile を持つ](../adr/hm-configs-independent-profiles.md) は `configs` の定義先を `modules/common.nix`
> （全モジュール共通）と定めているため、これは HM 固有ではなく共通オプション集合の一部で
> あり、本 item の規範に含めた。[rootKind に system を足しても schemaVersion を上げない](../adr/system-rootkind-without-version-bump.md)・[root は 3 つの marker で表す](../adr/three-root-markers.md) 由来の未追従を [root は 3 マーカーと絶対パス文字列の union を取る](root-marker-kinds.md) / [意図が矛盾する組み合わせをクロスフィールドの確認で評価時に拒否する](normalize-manifest-cross-field-checks.md) で
> 扱ったのと同じ扱いである。`configs` の `<name>` 次元が profile の粒度にとって何を意味するか
> （役割分離が可能になること・`<name>` 次元の導入が非破壊であること）は [HM モジュール経由でも名前つき config ごとに役割分離した独立 profile を取れる](hm-named-configs-profiles.md)、
> profile dir をどうキーするかは [profileDir は home と system のみ name 直キーとし、fixed root と --root 上書きは roothash でキーする](profiledir-keying.md) の担当。`docs/spec.md` の追従は本 item の
> 担当範囲外。
>
> **`layat.configs` のデフォルトを `{}` とする根拠**: 旧決定（現在は [HM モジュールは `layat.configs.<name>` ごとに独立した profile を持つ](../adr/hm-configs-independent-profiles.md)）は型
> （`attrsOf (submodule { entries = entriesType; })`）のみを定めデフォルトに触れていなかったが、
> `attrsOf` の既定が空 attrset であることは Nix module system の仕組みから従うことであり、原文の表が他 3
> オプションのデフォルトを明示しているのに `configs` だけ欠けると非対称になるため規範に含めた。
> 原文が `layat.entries` のデフォルトとする `{}` は、糖衣を通じて `configs.default.entries` の
> デフォルトへ落ちる。

## 出典

`docs/spec.md`「モジュールオプション仕様」→「共通オプション（全モジュール）」節の
オプション表と、その直後の root pin の段落。

決定の実体は root 明示必須とモジュール側 pin を定めた [root の明示を mkManifest 層で必須にする](../adr/root-explicit-at-mkmanifest.md)と、配置ロジックを engine が所有し
モジュールは配線に徹すると定めた [モジュールは root と起動の時機だけを供給して engine を起動する](../adr/modules-wire-engine-kick-only.md)。`layat.backup` の追加は [モジュールに layat.backup を起動配線のオプションとして持つ](../adr/module-backup-option.md)、`layat.configs`
の新設と `layat.entries` の糖衣化は [HM モジュールは `layat.configs.<name>` ごとに独立した profile を持つ](../adr/hm-configs-independent-profiles.md)・[layat.entries を configs.default.entries への rename 糖衣として残す](../adr/hm-entries-renamed-sugar.md)、`layat.recopy` / `layat.jobs` の追加は [layat.recopy で activation の起動に --recopy を付ける](../adr/module-recopy-option.md)・[layat.jobs で activation の並列度を指定する](../adr/module-jobs-option.md)、NixOS / nix-darwin の `systemRoot` pin は [NixOS / nix-darwin モジュールは root を systemRoot に固定する](../adr/system-modules-pin-systemroot.md)。
