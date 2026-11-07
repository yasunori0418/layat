---
id: "ADR-fdb8003f-db36-424b-8484-6f2fd57a3027"
type: adr
name: "engine は JSON とエラーを標準ライブラリで扱う"
status: 採用
issues:
  - "#552"
origin: "engine / CLI の技術スタックの確定（2026-06-13）。ADR-0011 §7 の JSON とエラー処理を書き直した"
justifies:
  - "REQ-b74a118a-1272-44eb-944c-7725163211c6"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-01T23:13:39+09:00"
---
# ADR-fdb8003f-db36-424b-8484-6f2fd57a3027: engine は JSON とエラーを標準ライブラリで扱う

- ステータス: 採用
- 日付: 2026-11-01
- 関連: GitHub Issue #552
- 起点: engine / CLI の技術スタックの確定（2026-06-13）。ADR-0011 §7 の JSON とエラー処理を書き直した

## 背景

engine は `manifest.json` を読み、配置の失敗を呼び出し元へ返す。engine は第三者依存を持たない方針で、
JSON の扱いとエラー処理をどう実装するかを決める必要があった。

## 決定

engine は、JSON を `encoding/json` で扱う。エラーは `fmt.Errorf` の `%w` でラップし、`errors.Is` / `errors.As` で
判別する。どちらも標準ライブラリで完結させる。

engine の依存の規範は REQ-b74a118a-1272-44eb-944c-7725163211c6 に置く。

## 根拠

どちらも標準ライブラリで要件を満たせるので、第三者依存を入れる理由が無い。sentinel と型で判別できる
ラップ済みのエラーは、CLI の側でエラーコードを分類する手掛かりにもなる。

## 棄却した代替案

元の検討では、代替案を記録していない。
