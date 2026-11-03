---
id: "ADR-29c0838b-3e25-4b09-a479-9ebbbc4466c7"
type: adr
name: "root 相対マーカーの基準点を manifest の配置 root にする"
status: 採用
issues:
  - "#362"
origin: "Issue #362（src 側のプロジェクトルート実行時解決）の grilling（2026-10-12）。ADR-0059 §2 を書き直した"
justifies:
  - "REQ-81249072-56b8-42f5-807b-ea623c1efe99"
  - "REQ-a77181dd-a59c-40bb-855e-3430ad970eb6"
references:
  - "ADR-b7f9b972-69b7-4bbe-affc-97978e1a55f7"
---
# ADR-29c0838b-3e25-4b09-a479-9ebbbc4466c7: root 相対マーカーの基準点を manifest の配置 root にする

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-b7f9b972-69b7-4bbe-affc-97978e1a55f7, `docs/spec.md`
- 起点: Issue #362（src 側のプロジェクトルート実行時解決）の grilling（2026-10-12）。ADR-0059 §2 を書き直した

## 背景

root 相対マーカーの `src` は、どこかの root からの相対パスである。その基準点を、target と同じ
配置 root にするか、`src` の側で root の種別（project / home など）を選ばせるかを決める必要がある。
リンクは target の親から `src/subpath` への相対パスで張る。

## 決定

root 相対マーカーの `src` の基準点は、その manifest の配置 root とする。engine が実行時に解決した
値（root の種別を問わず、`--root` による上書きを含む）を使い、`src` の側で root の種別を選ぶ
引数は設けない。src の解決の時点は REQ-81249072-56b8-42f5-807b-ea623c1efe99、マーカーの規則は
REQ-a77181dd-a59c-40bb-855e-3430ad970eb6 に置く。

## 根拠

相対リンクが持つのは target からの位置関係だけなので、target と別の root を基準にした `src` は
そもそも表せない。基準点を target と同じ root に固定すれば、選ぶ引数は構造的に要らなくなる。
root の種別を問わず同じ規則で使える。

## 棄却した代替案

### `src` の側で root の種別を選ばせる（`{ root = projectRoot; path = "..."; }`）

target と別の root を基準にした `src` は相対リンクで表せず、絶対パスで張る形に戻る。
