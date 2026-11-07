---
id: "REQ-637599dc-a1ec-4af5-9e97-e882c7df56d0"
type: requirement
name: "CLI が持ち込む依存は許可した第三者ライブラリと pin した Go に限り、いずれも固定する"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  Everything the CLI brings in SHALL be confined to what has been explicitly allowed, and
  each SHALL be pinned, in contrast to the engine's stdlib-only constraint. A minimal set
  of third-party dependencies SHALL be allowed — cobra, for subcommands and help,
  `github.com/pelletier/go-toml/v2`, for the configuration file that selects the manifest
  generator, and `github.com/yasunori0418/outturn/go`, for the `--json` envelope types and
  item id derivation — while placement itself is done by importing the engine. Those dependencies SHALL be
  fixed by building with `buildGoModule` and a vendorHash string. The Go toolchain SHALL
  be pinned to the go of nixpkgs, and the `toolchain` directive SHALL NOT be used, so that
  the build does not fetch a toolchain of its own and stays reproducible under Nix.
specification_ja: |
  CLI が持ち込むものは明示的に許可したものだけに閉じ、いずれも固定しなければならない
  （engine の stdlib-only 制約とは対照的に扱う）。第三者依存は最小限だけ許可しなければ
  ならない（サブコマンド / help のための cobra、manifest 生成器を選ぶ設定ファイルのための
  `github.com/pelletier/go-toml/v2`、`--json` のエンベロープ型と item id 導出のための
  `github.com/yasunori0418/outturn/go`）。配置そのものは engine を import して行う。
  それらの依存は `buildGoModule` と vendorHash 文字列でビルドすることによって固定しなければ
  ならない。
  Go ツールチェーンは nixpkgs の go に pin しなければならず、`toolchain` ディレクティブを
  使ってはならない（ビルドが独自にツールチェーンを取得せず Nix の下で再現可能に保つため）。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# REQ-637599dc-a1ec-4af5-9e97-e882c7df56d0: CLI が持ち込む依存は許可した第三者ライブラリと pin した Go に限り、いずれも固定する

## 仕様

| コンポーネント | 依存 |
|---|---|
| `cmd/layat`（CLI = `packages.layat`）| 配置エンジンを import。最小依存を許可（**cobra** = サブコマンド / help）。entrypoint 発見と `nix`（build / eval）オーケストレーションを担う。`buildGoModule` + **vendorHash 文字列**でビルド。Go は nixpkgs の go に pin し `toolchain` ディレクティブ不使用 |

> **上は原文の写しで、規範は frontmatter が正**。CLI が engine を import して配置を駆動
> すること、および entrypoint 発見と `nix` のオーケストレーションを担う役割分担そのものは
> REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645、entrypoint の探索順は REQ-1cc080f6-ae91-4c1f-973e-b7054cfc0198、実行フローにおける eval / build の
> 順序は REQ-60c6b7ea-e936-4ce8-bd75-ad35e9c693b9、engine 側の stdlib-only 制約は REQ-b74a118a-1272-44eb-944c-7725163211c6 の担当。規範文が
> engine の import に触れるのは、第三者依存を最小限に抑えられる前提が「配置を自前で
> 持たず engine に委ねること」にあるためで、その役割分担自体を規範化するものではない。

> **2026-10-02 追記（ADR-aec6bd1d-f776-4d7c-9337-1ed78fee6180）**: 許可する第三者依存に `github.com/pelletier/go-toml/v2`
> （生成器を選ぶ設定ファイル `layat.toml` / `config.toml` のパーサ）を足した。import は
> `cmd/layat` に閉じ、engine の stdlib-only は変わらない。表の「`nix`（build / eval）
> オーケストレーション」は manifest 生成器（既定 nix）経由に読み替える（→ ADR-87a2cf34-9592-4c81-8980-246bca579ca0）。

## 出典

`docs/spec.md`「依存関係」節の表の `cmd/layat` 行。

決定の実体は ADR-6b7becf0-f029-41e3-895e-b672dce94d3d「CLI は cobra で組む」・ADR-97edd9af-befb-4a9f-a8ef-de038c76336a「layat は buildGoModule と vendorHash でビルドする」・ADR-2be5080a-1cba-4fea-bed8-66e0d0d7b8d0「Go toolchain は nixpkgs の go に pin する」で、CLI に cobra を許可し
`buildGoModule` + vendorHash・Go の pin でビルドを固定することを定めている。CLI を一次 UX と
する位置づけは ADR-88231ffc-4836-4252-be34-894944fd28c3「CLI が一次 UX で devShell に pin 版を同梱する」が定めるが、同 ADR は cobra / vendorHash / Go の pin のいずれにも触れておらず、
この item の規範を決めていないため、側面の根拠として `justifies` は張らない（位置づけそのものの
帰属は REQ-14f0aec9-abae-4621-82f3-40536a1ad904 / REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645 が担当する）。

go-toml/v2 の許可は ADR-aec6bd1d-f776-4d7c-9337-1ed78fee6180「CLI の依存に go-toml/v2 を許可する」が定めている。

outturn/go の許可は ADR-bd8a4c8a-5e7e-485a-bccf-20dab63125f2「CLI は outturn/go に依存する」が定めている。
