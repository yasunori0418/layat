---
id: "REQ-d37d8433-a0c0-4387-aee9-9526f5cef70c"
type: requirement
name: "layat.recopy は engine 起動の配線レイヤーのオプションで、true のときだけ activation の起動へ --recopy を付ける"
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
---
# REQ-d37d8433-a0c0-4387-aee9-9526f5cef70c: layat.recopy は engine 起動の配線レイヤーのオプションで、true のときだけ activation の起動へ --recopy を付ける

## 仕様

- `layat.recopy = true` のときだけ、activation の `apply --all` 起動へ `--recopy` を付ける
- `--recopy` は起動に載る全 config へ一律に適用され、config 単位・entry 単位の上書きは無い
- 起動配線レイヤーのオプションで、manifest（entries）には影響しない
- オプションの説明文に「activation のたびに全 copy target を src から上書きし、ローカル編集を
  捨てる」ことを明記する

> **他 item との担当分界**: オプションの型とデフォルト値（`bool`・false）は
> REQ-fc1c7ce6-dc9d-4dd3-98f5-7877d9f99d10、`--recopy` が何を上書きしどう報告するかは
> REQ-7cc32a2b-eee4-4a29-8dc1-a1dc23e7a065、activation の起動形は REQ-c847d1af-a437-46bb-bd64-42083810d034 /
> REQ-8085f194-c903-4ecb-abd8-c719fe7b3292 の担当。`layat.backup` の配線を定める
> REQ-e1e1114b-ba07-4d57-8e04-6e30e39a5da3 と対になる。

## 出典

ADR-0061「モジュールに `layat.recopy` オプションを足し、`layat.jobs` とともに共通オプションへ
載せる」§1・§3。`--recopy` そのものは ADR-0020 §1 が定めている。
