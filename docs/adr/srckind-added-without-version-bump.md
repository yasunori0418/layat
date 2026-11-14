---
id: "ADR-b5431ecf-fe60-4b29-86af-9857478ce663"
type: adr
name: "srcKind の列挙値は schemaVersion を上げずに足す"
section: manifest-generator
status: stable
sources:
  - id: "issue-362"
    resource: "https://github.com/yasunori0418/layat/issues/362"
    title: "#362"
origin: "Issue #362（src 側のプロジェクトルート実行時解決）の grilling（2026-10-12）。ADR-0059 §5a と §5b を書き直した"
justifies:
  - "REQ-0b0cd1e3-bfeb-45c1-978d-e2e11c568336"
  - "REQ-79ce0a09-f9bd-4e61-ba7f-45fb5643137b"
  - "DSG-fd39b9ee-a5d2-46c2-bdcd-fc399a1e0296"
references:
  - "ADR-88231ffc-4836-4252-be34-894944fd28c3"
  - "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
  - "ADR-5e46ebd6-e44b-48dc-a93f-a8a3a3903967"
  - "ADR-6d2327f7-f513-42c3-89f8-808467db1914"
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# srcKind の列挙値は schemaVersion を上げずに足す

- ステータス: stable
- 日付: 2026-11-03
- 関連: [CLI が一次 UX で devShell に pin 版を同梱する](cli-primary-ux-devshell-pinned.md), [home mode の CLI と lib の版ずれを manifest の確認に任せて許容する](home-mode-skew-guarded-by-schemaversion.md), [copy と out-of-store の併用を eval 時に拒否する](copy-with-out-of-store-eval-error.md), [out-of-store はマッピングだけを追跡する](out-of-store-tracks-mapping-only.md), [manifest.json を schemaVersion 1 の唯一の契約として固定する](manifest-json-contract-schemaversion-v1.md), `docs/spec.md`
- 起点: Issue #362（src 側のプロジェクトルート実行時解決）の grilling（2026-10-12）。ADR-0059 §5a と §5b を書き直した

## 背景

root 相対マーカーを manifest に載せるには、`srcKind` に新しい値 `"rootRelative"` が要る。
`manifest.json` は Nix と engine の間の唯一の安定契約で、`schemaVersion` は 1 である。当時の
engine の検証は `schemaVersion` と root だけを見ており、`srcKind` / `method` / `rootKind` の値は
確認していなかった。そのため未知の値を読むと、エラーにならずに誤った配置をしうる状態だった。

## 決定

`srcKind` に `"rootRelative"` を足し、`schemaVersion` は 1 のまま据え置く。代わりに engine は
manifest を読むとき、entry ごとに `srcKind` / `method` / `rootKind` の値を確認し、未知の値を
明示のエラーにする。列挙値の一覧は [manifest.json の entries は attrset を配列へ正規化し 5 フィールドを記録する](../requirements/manifest-entries-array.md)、版の固定は
[manifest.json が Nix と engine の唯一の安定契約であり schemaVersion は 1 に固定する](../requirements/manifest-single-stable-contract.md)、値の確認は [rootRelative の LinkDest は target の親から src/subpath への相対パスを Entry だけから計算し、root が要るのは配置前の確認だけにする](../design/root-relative-linkdest.md)
に置く。

## 根拠

project mode では CLI と lib が devShell に同梱されるので、古い engine が新しい manifest を読むことは
構造的に起きない。home mode ではグローバルな CLI と flake が pin する lib が別の入力になり、版がずれうる。
この場合も値の確認があれば、未知の `srcKind` は誤配置されずに明示のエラーで止まる。
既存の manifest は全て既知の値なので、確認を足しても挙動は変わらない。

## 棄却した代替案

### `schemaVersion` を 2 に上げる

版ずれは project mode では起きず、home mode で起きても値の確認が誤配置を止める。全利用者の
manifest の版を上げる代償に見合う保護が無い。

### 値の確認を足さずに列挙値だけ足す

古い engine が新しい値を読んだとき、エラーにならずに誤った配置をしうる余地が残る。
