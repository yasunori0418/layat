---
id: "REQ-c6891aeb-13c0-4ae7-9ad1-5c343735266a"
type: requirement
name: "HM モジュール経由でも名前つき config ごとに役割分離した独立 profile を取れる"
status: stable
derives_from:
  - "UC-d39c1994-f9a5-4860-80ba-f6e584adaf14"
  - "UC-1c280dce-7c72-44c0-95ea-d06344f62a47"
specification: |
  A user of the home-manager module SHALL also be able to hold several independent
  profiles separated by role, each with its own generations, and role separation SHALL NOT
  be confined to the standalone CLI path. One `<name>` of `layat.configs` SHALL yield one
  manifest and one profile; the atomicity of "one config = one profile = one manifest"
  SHALL be preserved and no helper for composing several configs SHALL be provided.
  Introducing the `<name>` dimension SHALL be non-breaking: a config named `default` SHALL
  go on resolving to the same profile as before, and no change to the layout of the
  profile directory SHALL be entailed. How the profile directory is keyed is stated by
  REQ-d5a2e289-40bc-45a9-9d44-21b8dc561b81, and the definition of the options themselves — that
  `layat.configs.<name>.entries` is the canonical form and the bare `layat.entries` a
  deprecated sugar onto `configs.default.entries` — by REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10, being common to every
  module; neither is restated here.
specification_ja: |
  home-manager モジュールの利用者も、役割分離された複数の独立 profile（それぞれが自分の
  世代を持つ）を取れなければならず、役割分離を CLI 単体での使い方に限ってはならない。
  `layat.configs` の `<name>` 1 つが 1 manifest = 1 profile に対応しなければならない。
  「1 config = 1 profile = 1 manifest」の atomic 性は保たなければならず、複数 config を
  合成するヘルパを提供してはならない。`<name>` 次元の導入は非破壊でなければならず、`default`
  という名の config は従来と同じ profile に解決され続け、profile ディレクトリのレイアウト
  変更を伴ってはならない。profile ディレクトリをどうキーするかは REQ-d5a2e289-40bc-45a9-9d44-21b8dc561b81、
  オプション自体の定義
  （canonical な形が `layat.configs.<name>.entries` であり、素の `layat.entries` が
  `configs.default.entries` への deprecated 糖衣であること）は全モジュール共通であり
  REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10 の担当で、いずれも本 item では規定しない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:31:01+09:00"
---
# REQ-c6891aeb-13c0-4ae7-9ad1-5c343735266a: HM モジュール経由でも名前つき config ごとに役割分離した独立 profile を取れる

## 仕様

- `layat.configs` の属性キー `<name>` が profile 名で、CLI 単体の entrypoint `layat.<name>` と
  同じ次元。`<name>` ごとに独立した link-farm を生成し、1 config = 1 profile = 1 manifest の
  atomic 性を保つ（合成ヘルパは提供しない）
- HM モジュール経由でも**役割分離ができる**。役割ごとに分けたいユーザーが CLI 単体での使い方へ
  回る必要は無い
- `<name>` 次元の導入は**非破壊**。`default` という名の config は従来と同じ profile へ解決され
  続け、`default` 以外の `<name>` が増えるだけで profile dir のレイアウト変更は無い
  （キーの体系そのものは [profileDir は home と system のみ name 直キーとし、fixed root と --root 上書きは roothash でキーする](profiledir-keying.md) の担当）

> **`docs/spec.md` 原文と異なる理由**: 原文の blockquote は「`layat.entries` は単一 attrset =
> 単一 manifest = 1 profile（固定名 `default`）で `<name>` 次元を持たず、**HM モジュール経由では
> 役割分離はできない**。役割ごとに分けたいユーザーは standalone CLI 経路を使う。HM の複数
> profile 化（`layat.configs.<name>.entries` 等）は将来拡張の seam として残す」と述べるが、
> この MVP 限定は **[HM モジュールは `layat.configs.<name>` ごとに独立した profile を持つ](../adr/hm-configs-independent-profiles.md)・[layat.entries を configs.default.entries への rename 糖衣として残す](../adr/hm-entries-renamed-sugar.md) が実装決定済み**（seam を実装へ進め、`layat.entries` を
> `configs.default.entries` への糖衣にする）で反転している。`docs/spec.md` 側がこの改訂に
> 追従できていないため、分割にあたって原文の MVP 限定を規範へ持ち込まない判断をした
> （[root は 3 マーカーと絶対パス文字列の union を取る](root-marker-kinds.md) / [意図が矛盾する組み合わせをクロスフィールドの確認で評価時に拒否する](normalize-manifest-cross-field-checks.md) で [rootKind に system を足しても schemaVersion を上げない](../adr/system-rootkind-without-version-bump.md)・[root は 3 つの marker で表す](../adr/three-root-markers.md) 由来の未追従を扱ったのと同じ扱い）。
> `docs/spec.md` の追従は本 item の担当範囲外。
>
> **原文の残る規範の所在**: `layat.configs` / `layat.entries` のオプション定義そのものは
> [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](module-common-options.md)（[HM モジュールは `layat.configs.<name>` ごとに独立した profile を持つ](../adr/hm-configs-independent-profiles.md) が定義先を `modules/common.nix` としているため全モジュール共通の
> 集合に属する）、profile dir レイアウトの物理形は [profileDir は config 専用ディレクトリとし、profile リンク・世代・pending out-link をその中に並べる](profile-on-disk-layout.md)、home mode が `<name>` 直キーで
> あることは [profileDir は home と system のみ name 直キーとし、fixed root と --root 上書きは roothash でキーする](profiledir-keying.md)、CLI 単体が `layat.<name>` で複数 profile を持てることは
> [entrypoint は `layat.<name>` に named manifest を公開し CLI は形ごとの attr path で build する](named-manifest-addressing.md)、1 起動あたりの activation 契約は [home-manager モジュールの engine kick は activation から全 config のビルド済み link-farm を 1 回の起動で渡し、失敗で switch を止める](hm-activation-contract.md)、module の使い方で rollback を host へ
> 一本化することは [module 時は rollback を host へ一本化し、layat profile は前進のみで追従する](module-mode-profile-internal.md) の担当。
>
> **configs を導入した旧決定の activation と衝突検出の項を本 item に含めない理由**: その旧決定は activation が `configs` を
> 走査して profile ごとに engine を kick すること（辞書順・1 profile の失敗で後続を止めず
> 最後に集約）、単一 HM config 内の `configs` 間の正規化後 target 衝突を eval 時
> assertion で停止することも決めている（[モジュールの activation は apply --all --manifest を 1 回だけ起動する](../adr/module-activation-single-apply-all.md) が前者を `apply --all` 1 回起動へ、[モジュールは configs 間の target 衝突を評価時に確認しない](../adr/modules-no-eval-conflict-assertion.md) が後者を
> activation の `apply --all` 前段の確認へ改めた）が、いずれも `docs/spec.md` に対応記述が無い
> （原文が旧決定に未追従のため）。本 item は原文 blockquote の担当範囲＝profile 粒度に
> 留め、**前者の規範は [HM の activation は全 config の link-farm を apply --all 1 回の起動に渡し、報告は辞書順・配置の並列度は --jobs に従う](hm-activation-per-config-kick.md)、後者の規範は [単一 HM config 内の configs 間 target 衝突は activation の apply --all 前段の確認で停止する](hm-configs-target-collision-assertion.md) が持つ**。なお後者の前段
> 検出は、別 entrypoint・別 manifest の cross-config を「eval では検出不可・実行時の
> 後勝ち + foreign warning」とする [target 衝突の検出処理を同一 manifest 内と cross-config で分ける](target-collision-eval-detection.md) / [同一 target を複数 config で狙うことによる振動はユーザー責任とし warning で可視化するに留める](cross-config-target-oscillation.md) に対する**例外**（HM の
> `configs` は全 config が activation の単一の `apply --all` 起動に載るため配置前に検出できる）
> であり、この関係は [単一 HM config 内の configs 間 target 衝突は activation の apply --all 前段の確認で停止する](hm-configs-target-collision-assertion.md) の注記が明示している。

## 出典

`docs/spec.md`「モジュールオプション仕様」→「共通オプション（全モジュール）」節の
blockquote「HM モジュールの profile 粒度（MVP）」。

決定の実体は [HM モジュールは `layat.configs.<name>` ごとに独立した profile を持つ](../adr/hm-configs-independent-profiles.md)「HM モジュールは layat.configs.<name> ごとに独立した profile を持つ」と
[layat.entries を configs.default.entries への rename 糖衣として残す](../adr/hm-entries-renamed-sugar.md)「layat.entries を configs.default.entries への rename 糖衣として残す」（MVP 限定の旧決定を改訂）。`<name>` 次元と
atomic 性・合成ヘルパ不提供は [HM モジュールは `layat.configs.<name>` ごとに独立した profile を持つ](../adr/hm-configs-independent-profiles.md)、`default` が従来と同じ profile に解決され続ける非破壊性は
[HM モジュールは `layat.configs.<name>` ごとに独立した profile を持つ](../adr/hm-configs-independent-profiles.md)（profile dir が `<name>` 直キーに乗ること）と [layat.entries を configs.default.entries への rename 糖衣として残す](../adr/hm-entries-renamed-sugar.md)（`layat.entries` を `configs.default.entries`
への糖衣とすること）の双方に依る。
