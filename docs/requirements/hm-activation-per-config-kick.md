---
id: "REQ-c847d1af-a437-46bb-bd64-42083810d034"
type: requirement
name: "HM の activation は全 config の link-farm を apply --all 1 回の起動に渡し、報告は辞書順・配置の並列度は --jobs に従う"
status: stable
derives_from:
  - "UC-d39c1994-f9a5-4860-80ba-f6e584adaf14"
  - "UC-1c280dce-7c72-44c0-95ea-d06344f62a47"
specification: |
  Where `layat.configs` holds at least one config, the `home.activation.layat` of the
  home-manager module SHALL kick the engine exactly once per activation, as `layat apply --all --manifest <name>=<link-farm> … --jobs <n>`, passing
  one `<name>=<link-farm>` per config of `layat.configs` in the lexical order of `<name>`.
  `--jobs` SHALL always be given explicitly, with the value of `layat.jobs` even when it is
  0. Each `<name>=<link-farm>` SHALL be applied as its own profile with its own generation
  commit, so that the unit of atomicity stays one profile. Continuing past a partial
  failure, aggregating it and exiting non-zero SHALL be those of `apply --all` stated by
  REQ-4cbd9a0d-9f94-4747-8881-56020dc6d5af; the activation SHALL NOT loop over the configs
  nor aggregate failures in the shell, and SHALL only return the non-zero exit through
  `run` so that the switch stops. The report SHALL be presented in the lexical order of
  `<name>`, while the parallelism of the placement SHALL follow `--jobs`. Where
  `layat.configs` holds no config, the activation SHALL NOT invoke `layat`, since an
  `apply --all` without `--manifest` would take the entrypoint path. The contract of
  the kick is stated by REQ-8085f194-c903-4ecb-abd8-c719fe7b3292, and the correspondence of
  one config to one manifest and one profile by REQ-c6891aeb-13c0-4ae7-9ad1-5c343735266a;
  neither is restated here.
specification_ja: |
  `layat.configs` に config が 1 つ以上あるとき、home-manager モジュールの
  `home.activation.layat` は、activation 1 回につき engine を
  ちょうど 1 回、`layat apply --all --manifest <name>=<link-farm> … --jobs <n>` の形で起動
  しなければならない。`<name>=<link-farm>` は `layat.configs` の config ごとに 1 つ、`<name>`
  の辞書順に並べなければならない。`--jobs` は `layat.jobs` の値で、値が 0 でも常に明示で
  付けなければならない。各 `<name>=<link-farm>` はそれぞれ独立した profile として独立に世代を
  コミットしなければならない（atomic 性の単位を 1 profile に保つため）。部分失敗での続行・
  集約・非ゼロ終了は REQ-4cbd9a0d-9f94-4747-8881-56020dc6d5af が定める `apply --all` の
  ものでなければならない。activation は config をシェルでループしてはならず、失敗をシェルで
  集約してもならず、非ゼロ終了を `run` で返して switch を止めるだけにしなければならない。
  報告は `<name>` の辞書順で示さなければならず、配置の並列度は `--jobs` に従わなければ
  ならない。`layat.configs` に config が 1 つも無いときは、activation は `layat` を起動しては
  ならない（`--manifest` の無い `apply --all` は entrypoint を通る処理になるため）。起動の契約は REQ-8085f194-c903-4ecb-abd8-c719fe7b3292、1 config が
  1 manifest = 1 profile に対応することは REQ-c6891aeb-13c0-4ae7-9ad1-5c343735266a の担当で、
  いずれも本 item では規定しない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:31:01+09:00"
---
# REQ-c847d1af-a437-46bb-bd64-42083810d034: HM の activation は全 config の link-farm を apply --all 1 回の起動に渡し、報告は辞書順・配置の並列度は --jobs に従う

## 仕様

- `home.activation.layat` は全 config の link-farm を **1 回の起動**に渡す:
  `layat apply --all --manifest <name>=<link-farm> … --jobs <n>`。`<name>=<link-farm>` は
  config ごとに 1 つ、`<name>` の辞書順に並べる
- `--jobs` は `layat.jobs` の値を、0 でも常に明示で付ける
- 各 config は独立した profile で、profileDir 単位の flock・比較の基準との diff・古くなった
  配置の保守的な除去・世代コミットは config ごとに独立して実行される
- 部分失敗の続行・集約・非ゼロ終了は `apply --all` に任せ、activation はシェルでループも
  集約もしない。非ゼロ終了は `run` がそのまま返して switch を止める
- 報告は辞書順、配置の並列度は `--jobs` に従う
- config が 1 つも無いときは `layat` を起動しない

> **他 item との担当分界**: 起動 1 回の activation 契約（`entryAfter ["writeBoundary"]`・
> ビルド済み link-farm を渡すこと・activation が `nix eval` / `build` を行わないこと・engine
> error が非 0 終了で switch を止めること）は [home-manager モジュールの engine kick は activation から全 config のビルド済み link-farm を 1 回の起動で渡し、失敗で switch を止める](hm-activation-contract.md)。
> 1 config = 1 profile = 1 manifest の対応と `<name>` 次元そのものは
> [HM モジュール経由でも名前つき config ごとに役割分離した独立 profile を取れる](hm-named-configs-profiles.md)。`layat.configs` / `layat.jobs` オプションの定義は
> [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](module-common-options.md)。`--manifest <name>=<link-farm>` の繰り返しと
> `--all` の合成は [apply --manifest はビルド済み link-farm を engine へ直接適用し、--all と合成するときは name=link-farm を繰り返す](apply-manifest-direct.md)。並列適用と辞書順の集約・
> 部分失敗の続行は [apply --all は選択された config を並列に適用し、集約表示と results\[\] を辞書順にし、部分失敗でも続行して最後に集約する](apply-all-lexical-order-continue.md)、`--jobs` の解決規則は
> [apply --all は build を --jobs の worker pool で先行並列実行し build 失敗 config を部分失敗として集約する](apply-all-parallel-build-jobs.md)。configs 間の target 衝突の停止は
> [単一 HM config 内の configs 間 target 衝突は activation の apply --all 前段の確認で停止する](hm-configs-target-collision-assertion.md)。
>
> **本文箇条書き第 3 項に対応する規範文を持たない理由**: profile ごとに独立して実行される処理の
> 列挙は、`apply --all` が config ごとに engine を駆動することと、profileDir 単位の flock を
> 定める [flock の取得は既定 blocking とし --no-wait のときだけ try-lock でスキップする](flock-acquisition-mode.md)・profileDir のキーを定める
> [profileDir は home と system のみ name 直キーとし、fixed root と --root 上書きは roothash でキーする](profiledir-keying.md)・世代機構を定める
> [世代は link farm derivation を layat 自前 profile へコミットして積み、前世代 manifest から古くなった配置を除去する](generation-mechanism-link-farm-manifest.md) の合成から従うため、規範として再掲しない。

## 出典

[モジュールの activation は apply --all --manifest を 1 回だけ起動する](../adr/module-activation-single-apply-all.md)「モジュールの activation は apply --all --manifest を 1 回だけ起動する」。`layat.jobs` の定義は [layat.jobs で activation の並列度を指定する](../adr/module-jobs-option.md)。

`docs/spec.md` には対応記述が無いため、原文の写しは持たない（規範は frontmatter が正で、
上の箇条書きは ADR 本文の要約）。旧版（config ごとに起動し bash で集約する形）は
[モジュールの activation は apply --all --manifest を 1 回だけ起動する](../adr/module-activation-single-apply-all.md) で改められた。
