---
id: "ADR-29c0838b-3e25-4b09-a479-9ebbbc4466c7"
type: adr
name: "root 相対マーカーの基準点を manifest の配置 root にする"
status: stable
sources:
  - id: "issue-362"
    resource: "https://github.com/yasunori0418/layat/issues/362"
    title: "#362"
origin: "Issue #362（src 側のプロジェクトルート実行時解決）の grilling（2026-10-12）。ADR-0059 §2 を書き直した"
justifies:
  - "REQ-81249072-56b8-42f5-807b-ea623c1efe99"
  - "REQ-a77181dd-a59c-40bb-855e-3430ad970eb6"
references:
  - "ADR-b7f9b972-69b7-4bbe-affc-97978e1a55f7"
  - "ADR-5e46ebd6-e44b-48dc-a93f-a8a3a3903967"
  - "ADR-6d2327f7-f513-42c3-89f8-808467db1914"
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
  - "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# root 相対マーカーの基準点を manifest の配置 root にする

- ステータス: stable
- 日付: 2026-11-03
- 関連: [root 相対マーカーのリンクを target 起点の相対パスで張る](root-relative-link-as-relative-path.md), [copy と out-of-store の併用を eval 時に拒否する](copy-with-out-of-store-eval-error.md), [out-of-store はマッピングだけを追跡する](out-of-store-tracks-mapping-only.md), [manifest.json を schemaVersion 1 の唯一の契約として固定する](manifest-json-contract-schemaversion-v1.md), [home mode の CLI と lib の版ずれを manifest の確認に任せて許容する](home-mode-skew-guarded-by-schemaversion.md), `docs/spec.md`
- 起点: Issue #362（src 側のプロジェクトルート実行時解決）の grilling（2026-10-12）。ADR-0059 §2 を書き直した

## 背景

root 相対マーカーの `src` は、どこかの root からの相対パスである。その基準点を、target と同じ
配置 root にするか、`src` の側で root の種別（project / home など）を選ばせるかを決める必要がある。
リンクは target の親から `src/subpath` への相対パスで張る。

## 決定

root 相対マーカーの `src` の基準点は、その manifest の配置 root とする。engine が実行時に解決した
値（root の種別を問わず、`--root` による上書きを含む）を使い、`src` の側で root の種別を選ぶ
引数は設けない。src の解決の時点は [out-of-store の src は評価時の絶対パスか配置 root 相対のどちらかで、root 相対は target と同じ root に対して実行時に解決する](../requirements/out-of-store-eval-time-path.md)、マーカーの規則は
[mkRootRelativeSymlink は配置 root 相対の src を表すマーカーを返し、engine が target 起点の相対 symlink として張る](../requirements/root-relative-symlink-marker.md) に置く。

## 根拠

相対リンクが持つのは target からの位置関係だけなので、target と別の root を基準にした `src` は
そもそも表せない。基準点を target と同じ root に固定すれば、選ぶ引数は構造的に要らなくなる。
root の種別を問わず同じ規則で使える。

## 棄却した代替案

### `src` の側で root の種別を選ばせる（`{ root = projectRoot; path = "..."; }`）

target と別の root を基準にした `src` は相対リンクで表せず、絶対パスで張る形に戻る。
