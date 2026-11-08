---
id: "REQ-3f541d39-da41-4ef8-858b-707f54cf6a29"
type: requirement
name: "root マーカーは kind を運ぶ入れ物でパス解決は engine が行う"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  `lib.projectRoot` / `lib.homeRoot` / `lib.systemRoot` SHALL be pure functions returning
  a marker attrset that carries the kind of runtime resolution. They SHALL NOT be sugar
  that returns a path string. `mkManifest` SHALL record the kind in `manifest.json`, and
  resolving the concrete path SHALL be the engine's runtime responsibility.
specification_ja: |
  `lib.projectRoot` / `lib.homeRoot` / `lib.systemRoot` は、実行時解決の種別（kind）を
  運ぶマーカー attrset を返す純粋関数でなければならない。パス文字列を返す糖衣であっては
  ならない。`mkManifest` は kind を `manifest.json` に記録しなければならず、実体パスの
  解決は engine の実行時責務としなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:36:11+09:00"
---
# REQ-3f541d39-da41-4ef8-858b-707f54cf6a29: root マーカーは kind を運ぶ入れ物でパス解決は engine が行う

## 仕様

`root` 引数に渡す **root マーカー**は、`mkOutOfStoreSymlink` と同じ「マーカーを渡して
挙動を opt-in する」パターンに従う。

```
projectRoot :: marker   # 実行時に git rev-parse --show-toplevel（--root で上書き可）
homeRoot    :: marker   # 実行時に $HOME
systemRoot  :: marker   # /（distro 構想・将来）
```

- core lib（nixpkgs のみ依存）では **kind を運ぶマーカー attrset を返す純粋関数**。
  実体パス解決は engine の実行時責務。
- マーカーは「**実行時解決の種別を運ぶ入れ物**」であってパス文字列を返す糖衣ではない。
  `homeRoot` / `projectRoot` は評価時にパスへ展開できない（`$HOME` / git toplevel は
  実行環境依存）。
- `mkManifest` は kind を `manifest.json` に記録し、エンジンが実行時に解決する。

上のシグネチャブロックは原文の写しで、規範は frontmatter が正。`systemRoot` の
「distro 構想・将来」の限定は規範に含めない（理由は下の注記）。原文が同じ箇所で述べる
「暗黙デフォルトは無い・`root` を省略すると評価時エラー」は [root は明示必須で暗黙デフォルトを持たない](root-explicit-required.md) の担当。

## 出典

`docs/spec.md`「lib API」→「`lib.projectRoot` / `lib.homeRoot` / `lib.systemRoot`」、
および「`root` の値」節の注記。

> **`systemRoot` を「distro 構想・将来」としない理由**: system mode は **[rootKind に system を足しても schemaVersion を上げない](../adr/system-rootkind-without-version-bump.md)・[root は 3 つの marker で表す](../adr/three-root-markers.md) が
> 実装を決定済み**（`rootKind = "system"` を正規値として通し、engine は root = `/` へ
> 解決する）で、`docs/spec.md` がこの改訂に追従できていない。原文の限定を規範に採らない
> 扱いは [root は 3 マーカーと絶対パス文字列の union を取る](root-marker-kinds.md) と同じ。ただし本 item の規範は「marker が kind を運ぶこと」に
> 留まり、`systemRoot` が実行時に `/` へ解決することまでは主張しない。そのため
> [root は 3 つの marker で表す](../adr/three-root-markers.md) の決定のうち `/` への解決は本 item の `justifies` の根拠に含めない（当該規範は [root は 3 マーカーと絶対パス文字列の union を取る](root-marker-kinds.md) が
> 持つ）。
>
> **marker の関数適用形（`subdir`）**: 適用形 `homeRoot { subdir = ".config"; }` は本 item の
> 担当範囲外で、その規範は [root マーカーの subdir 適用形は全 target への前置と等価な評価時の糖衣で、root を動かさない](root-marker-subdir-sugar.md) が持つ。本 item は原文の範囲
> （marker が kind を運ぶこと）に留まる。
