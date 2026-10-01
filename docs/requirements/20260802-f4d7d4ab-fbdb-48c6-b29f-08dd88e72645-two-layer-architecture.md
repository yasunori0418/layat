---
id: "REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645"
type: requirement
name: "layat は CLI とエンジンの 2 層で構成する"
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  layat SHALL be composed of two layers: the layat CLI (`packages.layat`, the primary UX
  installed on PATH) and the engine (a Go library). The CLI SHALL discover an entrypoint
  and obtain the store path of the named manifest through a manifest generator (nix by
  default), and import the engine to drive placement, stale removal and profile swap. The engine
  SHALL take `manifest.json` as its input. The boundary between the two layers SHALL be
  `manifest.json` alone, so that the engine does not depend directly on Nix evaluation
  results. Which files are discovered as entrypoints is stated by the CLI specification,
  and how the engine carries out placement is stated by REQ-6c4e174a-4d16-477a-96ff-17cb4eb5b564; neither is
  restated here.
specification_ja: |
  layat は layat CLI（`packages.layat`・PATH 常駐の一次 UX）と engine（Go ライブラリ）の
  2 層で構成しなければならない。CLI は生成器（既定 nix）経由で entrypoint を発見して
  named manifest の store path を得て、engine を import して配置・
  stale 除去・profile swap を駆動しなければならない。engine は `manifest.json` を入力に
  取らなければならない。
  2 層の境界は `manifest.json` だけとし、engine が Nix の評価結果へ直接依存しないように
  しなければならない。どのファイルを entrypoint として発見するかは CLI 仕様の担当、
  engine がどう配置を行うかは REQ-6c4e174a-4d16-477a-96ff-17cb4eb5b564 の担当で、いずれも本 item では規定しない。
---
# REQ-f4d7d4ab-fbdb-48c6-b29f-08dd88e72645: layat は CLI とエンジンの 2 層で構成する

## 仕様

layat は **2 層**で構成する。

```
[layat CLI]  packages.layat（PATH 常駐・一次 UX）
  ・entrypoint(flake.nix/shell.nix/default.nix)を発見（CWD 既定 / -f 上書き）
  ・内部で nix build/eval を回し named manifest(manifest.json + symlink farm)の store path を得る
  ・engine(ライブラリ)を import して配置・stale 除去・profile swap を駆動
   ↓ manifest.json in
[engine]  Go ライブラリ
  ・manifest.json を入力に取り nix(profile)/git(toplevel)のみ叩く
  ・ネイティブ FS 操作で place/replace/remove、保守的 stale 除去、nix-env --set
```

層の境界は `manifest.json` が担う。CLI と engine の間で受け渡すのはこの JSON 契約のみで、
engine は Nix の評価結果へ直接依存しない。

> **上の図は原文の写しで、規範は frontmatter が正**。図に含まれる次の 2 点は本 item の
> 規範ではない。
>
> - entrypoint として発見するファイルと発見規則（`flake.nix` / `shell.nix` /
>   `default.nix`・CWD 既定・`-f` 上書き）→ `docs/spec.md`「CLI 仕様」→
>   「entrypoint の発見」節の担当（後続 PR で item 化）
> - engine が叩く外部コマンドの限定とネイティブ FS 操作・`nix-env --set` → REQ-6c4e174a-4d16-477a-96ff-17cb4eb5b564
>
> **2026-10-02 追記（ADR-0055）**: 図の「内部で nix build/eval を回し」は、manifest 生成器
> （既定 nix）経由で named manifest の store path を得る、に読み替える。entrypoint の発見も
> 生成器の Discover 操作になる。生成器の契約は REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729、パッケージ構成は
> DSG-25ad3cce-f921-4db2-aae0-c0263d4b8295 の担当。層の境界が `manifest.json` だけであることは変わらない。

## 出典

`docs/spec.md`「アーキテクチャ概要」。層の境界が `manifest.json` だけであることは、
同節の構成図が層間の受け渡しを `↓ manifest.json in` の 1 本だけで描いていることによる
（`manifest.json` を「engine が読む唯一の安定契約」と明示するのは同「manifest.json
スキーマ（v1）」節で、当該節は後続 PR の担当）。

manifest の取得を生成器経由にしたのは ADR-0055「manifest 生成器の契約」。
