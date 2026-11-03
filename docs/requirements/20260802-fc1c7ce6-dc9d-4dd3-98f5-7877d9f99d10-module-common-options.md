---
id: "REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10"
type: requirement
name: "全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する"
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
  is stated by REQ-c6891aeb-13c0-4ae7-9ad1-5c343735266a and is not restated here. A module SHALL pin the root by its
  own nature — home-manager to `homeRoot`, NixOS and nix-darwin to `systemRoot`, a devShell
  to `projectRoot` — and SHALL NOT
  expose a `root` option, so that a user of a module never restates it. This is a
  departure from `mkManifest` and the CLI entrypoint, where `root` is required to be
  stated explicitly (REQ-4ec3accc-8bb6-461f-9024-dcf0027849e4).
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
  REQ-c6891aeb-13c0-4ae7-9ad1-5c343735266a の担当で、本 item では規定しない。モジュールは自分の性質で root を pin
  しなければならず（home-manager は `homeRoot`・NixOS / nix-darwin は `systemRoot`・
  devShell は `projectRoot`）、`root`
  オプションを公開してはならない。モジュール利用者が root を再指定しないようにするため
  である。これは `root` を明示必須とする `mkManifest` / CLI entrypoint の層
  （REQ-4ec3accc-8bb6-461f-9024-dcf0027849e4）との差分である。
---
# REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10: 全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する

## 仕様

```
layat.enable  :: bool          # デフォルト: false
layat.entries :: attrsOf entry # デフォルト: {}（属性キー = target・→ ADR-c494e86e-ceb9-4380-b983-4e09553b01d6）
                              # ← ADR-3447db4a-cdec-4573-bbad-05bbffc34a65 により layat.configs.<name>.entries の
                              #   deprecated 糖衣へ降格済み（規範は frontmatter・下の注記を参照）
layat.backup.enable :: bool    # デフォルト: false（→ ADR-5af374dd-61a7-4a13-b2e6-fb5136673dd3）
layat.backup.suffix :: str     # デフォルト: "layat-backup"（→ ADR-5af374dd-61a7-4a13-b2e6-fb5136673dd3）
```

モジュールは自分の性質で root を pin する（HM → `homeRoot` / NixOS / nix-darwin → `systemRoot` / devShell → `projectRoot`）ため、
モジュール利用者は `root` を再指定しない。

> **上は原文の写しで、規範は frontmatter が正**。`entries` の属性キーが target であることと
> entry submodule のフィールド定義は REQ-cb77ea05-bab8-4ccf-b09e-d23d8f71cdc7 / REQ-a33a11e3-830d-4142-88ed-4c1fc35e7f74、`root` が
> `mkManifest` / CLI 層で明示必須であることは REQ-4ec3accc-8bb6-461f-9024-dcf0027849e4、`--backup` の退避契約そのものは
> REQ-5dd5a4e9-6162-4fa5-b295-66844f5a4f3b の担当。`layat.backup` が manifest ではなく engine 起動の配線に効くことは
> REQ-e1e1114b-ba07-4d57-8e04-6e30e39a5da3、`layat.recopy` の配線は REQ-d37d8433-a0c0-4387-aee9-9526f5cef70c の担当。`layat.jobs` を activation の `--jobs` へ渡すことは
> REQ-c847d1af-a437-46bb-bd64-42083810d034 の担当。
>
> **2026-10-15 追記（ADR-9506276e-2108-4962-9f90-efe82fec7f66 / ADR-e8bb6153-2528-419e-8be0-c862a3eb4e05）**: 写しの表に無い `layat.recopy` / `layat.jobs` は ADR-9506276e-2108-4962-9f90-efe82fec7f66・ADR-e8bb6153-2528-419e-8be0-c862a3eb4e05 が
> 共通オプションへ足した（規範は frontmatter）。
>
> **2026-10-22 追記（ADR-5dad12a8-ada3-4a04-a3cf-1795a5b44c98）**: 写しの表に無い `layat.configs.<name>.subdir` は ADR-5dad12a8-ada3-4a04-a3cf-1795a5b44c98 が
> `configs` submodule へ足した（型とデフォルトの規範は frontmatter）。subdir から root の適用形を
> 組む HM の配線は REQ-068f2fb3-5459-4325-9359-c931b5049e0d の担当で、root pin の条文は変わらない。
>
> **2026-10-24 追記（ADR-c2ea1648-5e7a-47a1-a2c6-30169f93e22c）**: root pin の条文に NixOS / nix-darwin → `systemRoot` を足した。両モジュールは
> `layat.user` を持たず、subdir から `systemRoot` の適用形を組む配線は REQ-a220bb28-89de-42b0-a854-e7468755ecb1 の担当。
>
> **写しの `layat.entries` を canonical としない理由**: **ADR-01abcd5a-464f-4c44-8afd-30950631487c・ADR-3447db4a-cdec-4573-bbad-05bbffc34a65 は
> `layat.configs.<name>.entries` を canonical とし、素の `layat.entries` を
> `configs.default.entries` への deprecated 糖衣へ降格することを決定済み**で、`docs/spec.md`
> の当該表はこれに未追従。ADR-01abcd5a-464f-4c44-8afd-30950631487c は `configs` の定義先を `modules/common.nix`
> （全モジュール共通）と定めているため、これは HM 固有ではなく共通オプション集合の一部で
> あり、本 item の規範に含めた（REQ-37b56673-6e40-4a1b-a2a7-5d3c084e3e66 / REQ-16faf428-77f3-492f-b858-222c5274cbf7 で ADR-9bdf2604-c7b3-468a-b1e4-2478ef37477c 由来の未追従を
> 扱ったのと同じ扱い）。`configs` の `<name>` 次元が profile の粒度にとって何を意味するか
> （役割分離が可能になること・`<name>` 次元の導入が非破壊であること）は REQ-c6891aeb-13c0-4ae7-9ad1-5c343735266a、
> profile dir をどうキーするかは REQ-d5a2e289-40bc-45a9-9d44-21b8dc561b81 の担当。`docs/spec.md` の追従は本 item の
> 担当範囲外。
>
> **`layat.configs` のデフォルトを `{}` とする根拠**: 旧決定（現在は ADR-01abcd5a-464f-4c44-8afd-30950631487c）は型
> （`attrsOf (submodule { entries = entriesType; })`）のみを定めデフォルトに触れていなかったが、
> `attrsOf` の既定が空 attrset であることは Nix module system の帰結であり、原文の表が他 3
> オプションのデフォルトを明示しているのに `configs` だけ欠けると非対称になるため規範に含めた。
> 原文が `layat.entries` のデフォルトとする `{}` は、糖衣を通じて `configs.default.entries` の
> デフォルトへ落ちる。

## 出典

`docs/spec.md`「モジュールオプション仕様」→「共通オプション（全モジュール）」節の
オプション表と、その直後の root pin の段落。

決定の実体は root 明示必須とモジュール側 pin を定めた ADR-f85f1afd-9b87-4f7a-922e-97c62554f198「root の明示を mkManifest 層で必須にする」と、配置ロジックを engine が所有し
モジュールは配線に徹すると定めた ADR-ad5bd91d-e061-4a1e-9eef-651f4c9ce880。`layat.backup` の追加は ADR-5af374dd-61a7-4a13-b2e6-fb5136673dd3、`layat.configs`
の新設と `layat.entries` の糖衣化は ADR-01abcd5a-464f-4c44-8afd-30950631487c・ADR-3447db4a-cdec-4573-bbad-05bbffc34a65、`layat.recopy` / `layat.jobs` の追加は ADR-9506276e-2108-4962-9f90-efe82fec7f66・ADR-e8bb6153-2528-419e-8be0-c862a3eb4e05、NixOS / nix-darwin の `systemRoot` pin は ADR-c2ea1648-5e7a-47a1-a2c6-30169f93e22c。
