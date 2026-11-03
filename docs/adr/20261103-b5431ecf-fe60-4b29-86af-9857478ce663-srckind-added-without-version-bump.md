---
id: "ADR-b5431ecf-fe60-4b29-86af-9857478ce663"
type: adr
name: "srcKind の列挙値は schemaVersion を上げずに足す"
status: 採用
issues:
  - "#362"
origin: "Issue #362（src 側のプロジェクトルート実行時解決）の grilling（2026-10-12）。ADR-0059 §5a と §5b を書き直した"
justifies:
  - "REQ-0b0cd1e3-bfeb-45c1-978d-e2e11c568336"
  - "REQ-79ce0a09-f9bd-4e61-ba7f-45fb5643137b"
  - "DSG-fd39b9ee-a5d2-46c2-bdcd-fc399a1e0296"
---
# ADR-b5431ecf-fe60-4b29-86af-9857478ce663: srcKind の列挙値は schemaVersion を上げずに足す

- ステータス: 採用
- 日付: 2026-11-03
- 関連: `docs/spec.md`
- 起点: Issue #362（src 側のプロジェクトルート実行時解決）の grilling（2026-10-12）。ADR-0059 §5a と §5b を書き直した

## 背景

root 相対マーカーを manifest に載せるには、`srcKind` に新しい値 `"rootRelative"` が要る。
`manifest.json` は Nix と engine の間の唯一の安定契約で、`schemaVersion` は 1 である。当時の
engine の検証は `schemaVersion` と root だけを見ており、`srcKind` / `method` / `rootKind` の値は
検査していなかった。そのため未知の値を読むと、黙って誤った配置をしうる状態だった。

## 決定

`srcKind` に `"rootRelative"` を足し、`schemaVersion` は 1 のまま据え置く。代わりに engine は
manifest を読むとき、entry ごとに `srcKind` / `method` / `rootKind` の値を検査し、未知の値を
明示のエラーにする。列挙値の一覧は REQ-0b0cd1e3-bfeb-45c1-978d-e2e11c568336、版の固定は
REQ-79ce0a09-f9bd-4e61-ba7f-45fb5643137b、値の検査は DSG-fd39b9ee-a5d2-46c2-bdcd-fc399a1e0296
に置く。

## 根拠

CLI と lib は devShell に同梱されるので、古い engine が新しい manifest を読む経路は構造的に無い。
値の検査を足せば、仮にその経路ができても、未知の `srcKind` は誤配置されずに明示のエラーで止まる。
既存の manifest は全て既知の値なので、検査を足しても挙動は変わらない。

## 棄却した代替案

### `schemaVersion` を 2 に上げる

版ずれが起きない以上、全利用者の manifest の版を上げる代償に見合う保護が無い。未知の値を
明示のエラーにする保護は、値の検査で得られる。

### 値の検査を足さずに列挙値だけ足す

古い engine が新しい値を読んだとき、エラーにならずに誤った配置をしうる余地が残る。
