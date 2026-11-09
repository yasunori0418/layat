---
id: "UBQ-dbbd91ee-1d41-438b-9c6f-6972643741cb"
type: ubiquitous_term
name: "project mode"
status: stable
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T01:45:24+09:00"
---

# project mode

## Definition

The placement mode where root = the project root, selected with the `projectRoot` marker. Its placements are ephemeral placement (not committed); generations and rollback are not exposed to the user. The root resolves to the git toplevel.

## Avoid

"relative to CWD", "relative to the config file".

In project mode, bundling `layat` in the devShell is canonical. `templates/project` puts a pinned `layat` in the devShell `packages`, so the CLI and `layat.lib` (manifest `schemaVersion`) match through the same flake input. A global install is a convenience for the CLI on its own (home mode).

## 定義

root = プロジェクトルートの配置モード。`projectRoot` マーカーで選ぶ。配置物は ephemeral placement（コミットしない）であり、世代・rollback はユーザーに公開しない。root は git toplevel に解決される。

## 避ける表現

「CWD 相対」「設定ファイル相対」と説明すること。

project mode の `layat` は devShell 同梱が canonical。`templates/project` の devShell `packages` に pin 版 `layat` を入れ、CLI と `layat.lib`（manifest `schemaVersion`）を同一 flake 入力で一致させる。グローバル install は CLI 単体（home mode）の利便。
