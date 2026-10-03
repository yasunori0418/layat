---
id: "REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729"
type: requirement
name: "CLI は manifest 生成器の契約（Discover / Roots / Build / DryBuild）越しに manifest を得る"
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  The CLI SHALL obtain manifests only through a manifest generator contract, and the CLI
  layer SHALL keep no more than the selection of a generator and its injection into the
  engine. The contract SHALL take a config `name` as its only config argument, without a
  system, and SHALL provide four operations: Discover, which finds the entrypoint; Roots,
  which returns `manifest.Root` (rootKind and the fixed root when fixed) for one config or
  for all configs without building anything; Build, which
  takes a name and a pending path; and DryBuild, which returns the store path of the
  link-farm without laying down a gcroot. The all-configs Roots and the prebuilt generator's
  Roots SHALL also return the normalized target list, while a single-config Roots MAY omit
  it. Every generator SHALL answer Roots before any
  build, because the apply flow resolves profileDir and takes the lock before it builds.
  Build SHALL lay down at the pending path, and return, a link-farm that can be committed
  as a generation, which means a store path. `--manifest` SHALL be served by a prebuilt
  generator whose Discover is the given path, whose Roots reads that link-farm's
  `manifest.json`, and whose Build returns that path.
specification_ja: |
  CLI は manifest を manifest 生成器の契約越しにのみ得なければならず、CLI 層に残すのは
  生成器の選択と engine への注入だけにしなければならない。契約が取る config の引数は
  `name` のみとし（system は取らない）、次の 4 操作を持たなければならない: entrypoint を
  発見する Discover、build せずに 1 config 分または全 config 分の `manifest.Root`
  （rootKind・fixed のときの root）を返す Roots、name と pending の
  パスを取る Build、gcroot を張らずに link-farm の store path を返す DryBuild。
  全 config 分の Roots と prebuilt 生成器の Roots は正規化後の target 一覧も返さなければ
  ならず、1 config 分の Roots はそれを省いてもよい。
  apply の実行フローは build の前に profileDir を確定してロックを取るため、全ての生成器は
  build の前に Roots に答えなければならない。Build は世代コミット可能な link-farm
  （= store path）を pending のパスに張って返さなければならない。`--manifest` は prebuilt
  生成器が担わなければならない（Discover = 与えられたパス、Roots = その link-farm の
  `manifest.json` から読む、Build = そのパスを返す）。
---
# REQ-194e4209-d804-4a4b-a2b8-3d39c6c33729: CLI は manifest 生成器の契約（Discover / Roots / Build / DryBuild）越しに manifest を得る

## 仕様

| 操作 | 入力 | 出力・事後条件 |
|---|---|---|
| Discover | `-f` の値（無ければ cwd） | entrypoint を発見する |
| Roots | `name`（`--all` 用の一括取得は全件） | build せずに `manifest.Root`（`RootKind` / `Root`）を返す。`--all` 用の一括取得と prebuilt は `Targets` も返す（単一 config では任意） |
| Build | `name` + pending のパス | 世代コミット可能な link-farm（= store path）を pending に張って返す |
| DryBuild | `name` | link-farm の store path を gcroot を張らずに返す |

- Roots の先読みは能力宣言ではなく契約の前提条件。apply の「eval 先行 → flock → build」
  （REQ-60c6b7ea-e936-4ce8-bd75-ad35e9c693b9）を全ての生成器で成り立たせるため。
- Build の事後条件が store path なのは、世代のバックエンドが nix profile で `nix-env --set` が
  store 外のパスを拒否するため。この条件を緩めるのは後継 epic #406 の担当。
- `--manifest` は prebuilt 生成器で、その外面（フラグ・`-f` / `--all` との排他）は
  REQ-dec58330-6dad-47f7-8f56-2402764a89c7 の担当。
- `Targets` は Go 構造体 `manifest.Root` 側のフィールドで、manifest.json スキーマ v1 は変えない。
  消費者は `apply --all` の cross-config target 衝突検査だけなので、課すのは一括取得と prebuilt に
  限る（単一 config の Roots で返すと nix 実装の eval が 1 本増える）。

生成器の選び方は REQ-badc7e10-0ba4-40d4-b334-6e40193119db、診断の扱いは
REQ-7a2f1ecf-4675-45aa-80c0-a8fc58db9edd の担当。

## 出典

ADR-0055「manifest 生成器の契約」§1〜§5。
