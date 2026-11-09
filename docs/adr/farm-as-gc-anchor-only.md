---
id: "ADR-30f105c7-85a8-4fab-9dea-3277d932b524"
type: adr
name: "symlink farm を GC アンカー専用にする"
status: stable
origin: "実装前のレビューと横断検証で詰めた GC アンカーの細目。ADR-0006 の決定節（GC アンカーの項）・ADR-0013 §1・ADR-0016 §5・ADR-0019 §2 を書き直した"
justifies:
  - "REQ-60e6b49c-9ba1-4552-a0ec-d340421ec281"
  - "REQ-b12fc3c0-d7fe-4003-922c-f3ac0d969b66"
  - "REQ-62eda895-efd4-4eaf-a58b-600e8637da75"
references:
  - "ADR-f90509f2-d75a-4162-baa5-a5cfdb8a7c30"
  - "ADR-521765d7-7e31-42f4-b548-979185a7586a"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
  - "ADR-86a77232-dc99-4c04-bd6d-2127410e8828"
  - "ADR-15a358f4-1c81-49d5-b658-3a1fe39f0f75"
  - "ADR-74bf4a1f-bb74-44a7-a3cb-3b3e87c24f95"
  - "ADR-2119433e-de69-4447-85a5-575d58246007"
  - "ADR-45380022-d959-4b22-aa05-5993ac1f3c87"
  - "ADR-74423bb8-1453-4689-a25e-30bf46536afc"
  - "ADR-a51897c5-f709-4b7a-8db2-61e6ebb0040c"
  - "ADR-f607b605-c69e-4c7e-ac89-75b7bce5bee1"
  - "ADR-88231ffc-4836-4252-be34-894944fd28c3"
  - "ADR-d15a216f-b5d6-4df9-9825-b23a341bec1d"
  - "ADR-2f472fc4-23ed-4169-aa49-8bded9e336d9"
  - "ADR-dfdcc588-2e23-4d3b-ae17-b508a8d85ff0"
  - "ADR-553eea60-6ab4-4da7-93a9-3acf44b2fae3"
  - "ADR-cbac86c1-d774-4cf3-abe6-ca8a7c2f1995"
  - "ADR-c4ad9987-850e-4834-bbea-d8de3b073548"
  - "ADR-6d2327f7-f513-42c3-89f8-808467db1914"
  - "ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93"
  - "ADR-9f2cc068-8edb-4e3d-958b-0dff8f6ed90c"
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
  - "ADR-ee31b8d0-f7c5-442f-ae90-dcd1dd16b7a0"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:41+09:00"
---
# symlink farm を GC アンカー専用にする

- ステータス: stable
- 日付: 2026-11-01
- 関連: [世代は link farm と nix profile に乗せる](generations-on-link-farm-and-nix-profile.md), [lib は純データ生成に徹する](lib-produces-data-only.md), [1 config = 1 profile を全モードで自前に持つ](one-config-one-profile-all-modes.md), [mkActivationScript で config ごとのラッパー derivation を提供する](mkactivationscript-wrapper.md), [module と project では世代操作を公開しない](module-and-project-hide-generation-ops.md), [CLI 単体では世代を常時有効にする](standalone-generations-always-on.md), [engine を単一の固定 Go バイナリにする](engine-fixed-go-binary.md), [flock は明示実行で待ち、hook 実行でスキップする](flock-mode-explicit-wait-hook-skip.md), [新規配置と張替えを先に、世代のコミットを最後に行う](placement-order-commit-last.md), [独立した操作はサブコマンドにし、操作の修飾はフラグにする](subcommand-set.md), [終了コードは 0 / 1 / 2 の 3 値にする](exit-code-scheme.md), [CLI が一次 UX で devShell に pin 版を同梱する](cli-primary-ux-devshell-pinned.md), [templates と layat init は固定 ref で導入する](templates-and-init-fixed-ref.md), [eval を先に行い、flock を取ってから build する](eval-first-then-flock-then-build.md), [テストを sandbox 可否で flake check と E2E に分ける](test-strategy-layers.md), [engine は cmd/ と internal/ に置きライブラリ化しない](engine-internal-cmd-layout.md), [一部の entry だけを適用する --only を提供する](partial-apply-only-flag.md), [copy は世代の外に place-once で置く](copy-outside-generations-place-once.md), [out-of-store はマッピングだけを追跡する](out-of-store-tracks-mapping-only.md), [古くなった配置の除去は manifest の差分で行う](stale-removal-by-manifest-diff.md), [古くなった配置の除去は記録どおりの symlink だけに限る](conservative-removal-invariant.md), [manifest.json を schemaVersion 1 の唯一の契約として固定する](manifest-json-contract-schemaversion-v1.md), [manifest 内の target の衝突を評価時に検出する](manifest-target-conflict-eval-error.md), `docs/spec.md`
- 起点: 実装前のレビューと横断検証で詰めた GC アンカーの細目。ADR-0006 の決定節（GC アンカーの項）・ADR-0013 §1・ADR-0016 §5・ADR-0019 §2 を書き直した

## 背景

layat は世代を profile に積み、rollback で前の世代へ戻せるようにする。戻った先の配置元が
`nix-collect-garbage` で消えていては rollback が成り立たないので、各世代は自分が使う store パスを
GC から守る必要がある。

## 決定

`lib.mkManifest` が返す derivation は、`manifest.json` と並べて、配置元を指す symlink farm を持つ。
farm は GC アンカー専用とし、engine は配置に farm を使わず、`manifest.json` の解決済み `src` の
文字列を使う。アンカーを持つのは store 由来かつ `method = "symlink"` の entry に限る。アンカーの
名前は `target` のハッシュにする。

アンカーの対象と名前の規範は [symlink farm は GC アンカー専用でアンカーは store-backed な symlink entry に限る](../requirements/farm-gc-anchor-only.md)・
[symlink farm の GC アンカー名は target のハッシュとする](../requirements/farm-anchor-name-hash.md) に、derivation の中身は
[mkManifest は manifest.json と symlink farm を含む store オブジェクトを返す](../requirements/mkmanifest-return-value.md) に置く。

## 根拠

store パスを文字列として `manifest.json` に書くだけでは、GC の参照は Nix による内容のスキャンに頼る
ことになり、パスの表現を変えた途端に、警告もなく参照が消える。明示の symlink で張れば、profile の世代が
GC root として、アンカーを持つ配置元を確実に保持する。

out-of-store の配置元は store の外にあるので保持する対象が無い。copy の entry は一度だけ実体化した後は
store から独立し、世代の外に置かれるので、配置元を保持し続ける理由が無い。保持すると不要な store が
残り、GC を妨げる。

farm は配置に使わないので、アンカー名は読める必要が無い。`target` をサニタイズした名前は別の
`target` が同じ名前になり得るが、ハッシュなら衝突しない。

## 棄却した代替案

### `manifest.json` だけを持ち symlink farm を持たない

GC の参照が store パスの内容スキャンに頼り、表現を変えると警告もなく参照が消える。

### copy の entry もアンカーする

扱いは一貫するが、世代の外にある copy のために store を保持し続け、GC を妨げる。

### アンカー名を `target` のサニタイズにし、衝突したら連番を付ける

読めるが、連番が並び順に依存して名前が安定せず、衝突の回避もハッシュより脆い。

### アンカーを store パスごとに 1 つへ集約する

GC の目的には足りるが、`target` との対応が失われる。
