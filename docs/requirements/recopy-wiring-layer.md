---
id: "REQ-d37d8433-a0c0-4387-aee9-9526f5cef70c"
type: requirement
name: "layat.recopy は engine 起動の配線レイヤーのオプションで、true のときだけ activation の起動へ --recopy を付ける"
status: stable
derives_from:
  - "UC-d39c1994-f9a5-4860-80ba-f6e584adaf14"
specification: |
  The `layat.recopy` option, common to every module, SHALL wire `--recopy` onto the
  `layat apply --all` invocation performed at activation only where `layat.recopy` is true,
  overwriting every copy target from its src and discarding local edits being an explicit
  opt-in by the user; where it is false, the invocation SHALL NOT carry `--recopy`.
  `--recopy` SHALL apply to every config of that invocation alike, and the module SHALL NOT
  offer a per-config or per-entry override. `layat.recopy` SHALL be an option of the
  invocation-wiring layer, independent of `entries` (the manifest v1 contract), and SHALL
  NOT affect the manifest itself. The description of the option SHALL state that enabling
  it overwrites every copy target from src at each activation and discards local edits.
specification_ja: |
  全モジュール共通の `layat.recopy` は、値が true のときに限り、activation で行う
  `layat apply --all` 起動へ `--recopy` を配線しなければならない（全 copy target を src から
  上書きしローカル編集を捨てるのは、ユーザーの明示 opt-in であるため）。false のときは起動に
  `--recopy` を付けてはならない。`--recopy` はその起動に載る全 config へ一律に適用されなければ
  ならず、モジュールは config 単位・entry 単位の上書きを提供してはならない。`layat.recopy` は
  `entries`（manifest v1 契約）とは独立な起動配線レイヤーのオプションでなければならず、
  manifest 自体に影響してはならない。オプションの説明文には、有効にすると activation のたびに
  全 copy target を src から上書きしローカル編集を捨てることを明記しなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:04:55+09:00"
---
# layat.recopy は engine 起動の配線レイヤーのオプションで、true のときだけ activation の起動へ --recopy を付ける

## 仕様

- `layat.recopy = true` のときだけ、activation の `apply --all` 起動へ `--recopy` を付ける
- `--recopy` は起動に載る全 config へ一律に適用され、config 単位・entry 単位の上書きは無い
- 起動配線レイヤーのオプションで、manifest（entries）には影響しない
- オプションの説明文に「activation のたびに全 copy target を src から上書きし、ローカル編集を
  捨てる」ことを明記する

> **他 item との担当分界**: オプションの型とデフォルト値（`bool`・false）は
> [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](module-common-options.md)、`--recopy` が何を上書きしどう報告するかは
> [apply --recopy は config 内の全 copy target を src から無条件に上書き再コピーする](apply-recopy.md)、activation の起動形は [HM の activation は全 config の link-farm を apply --all 1 回の起動に渡し、報告は辞書順・配置の並列度は --jobs に従う](hm-activation-per-config-kick.md) /
> [home-manager モジュールの engine kick は activation から全 config のビルド済み link-farm を 1 回の起動で渡し、失敗で switch を止める](hm-activation-contract.md) の担当。`layat.backup` の配線を定める
> [layat.backup は engine 起動の配線レイヤーのオプションで manifest には影響しない](backup-wiring-layer.md) と対になる。

## 出典

[layat.recopy で activation の起動に --recopy を付ける](../adr/module-recopy-option.md)・[モジュールの起動修飾オプションを config ごとに持たない](../adr/module-launch-options-global.md)。`--recopy` そのものは [apply --recopy で全 copy target を src から無条件に上書きする](../adr/apply-recopy-overwrites-copies.md) が定めている。
