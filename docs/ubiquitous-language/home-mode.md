---
id: "UBQ-229fb5db-7865-4679-917c-2c067ea571c9"
type: ubiquitous_term
name: "home mode"
status: stable
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T01:55:15+09:00"
---

# home mode

## Definition

The placement mode where root = `$HOME`, selected explicitly with the `homeRoot` marker. Used by both standalone and modules such as home-manager. It commits a generation each time and exposes `--rollback` to the user.

## Avoid

"standalone-only", "the default when `root` is omitted".

## 定義

root = `$HOME` の配置モード。`homeRoot` マーカーで明示選択する。CLI 単体も home-manager 等の module もこのモードを使う。世代を毎回コミットし、`--rollback` をユーザーに公開する。

## 避ける表現

「CLI 単体専用」「`root` 省略時の既定」と説明すること。
