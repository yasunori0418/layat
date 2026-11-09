---
id: "UBQ-e6466d28-6ab6-43be-a370-cca82eb1cfdb"
type: ubiquitous_term
name: "root"
status: stable
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-07T01:45:24+09:00"
---

# root

## Definition

The base path for placement. Chosen explicitly via the `root` argument of the public API (there is no implicit default). Its type is a union of `string` (absolute path, fixed at evaluation time) and `marker` (resolved at runtime). The markers are projectRoot, homeRoot, and systemRoot. Each marker also accepts the applied form `homeRoot { subdir = "..."; }`. The applied form does not move root; it is evaluation-time sugar equivalent to prefixing every target with `<subdir>/`.

## Avoid

"`$HOME` is fixed", "defaults to home mode", "the default when `root` is omitted"; describing a marker as "sugar that returns a path string" (it is a runtime-resolution kind, not expanded to a path at evaluation time); describing the applied form as "root actually resolves to the subdir" (`$HOME/.config` becoming root). Root and `rootKind` stay the same as the bare marker; only the targets change.

## 定義

配置の基準パス。公開 API の `root` 引数で明示的に選ぶ（暗黙のデフォルトは持たない）。型は `string`（絶対パス・評価時に固定）と `marker`（実行時に解決）の union。マーカーは projectRoot / homeRoot / systemRoot。各マーカーは適用形 `homeRoot { subdir = "..."; }` も受け付ける。適用形でも root は動かず、全 target へ `<subdir>/` を前置するのと等価な評価時の糖衣である。

## 避ける表現

「`$HOME` 固定」「既定は home mode」「`root` 省略時の既定」と説明すること、マーカーを「パス文字列を返す糖衣」と説明すること（実行時解決の種別であって、評価時にパスへ展開しない）、適用形を「root が実際に subdir へ解決される」と説明すること（root と `rootKind` は単体形と同じで、変わるのは target だけ）。
