---
id: "REQ-8d965ca2-f8fd-44a4-87f3-94e850e9f85b"
type: requirement
name: "home mode の root は層ごとに定まった供給元から解決する"
status: stable
derives_from:
  - "UC-f2436d68-91ff-4c48-b1df-47acefe4f464"
  - "UC-d39c1994-f9a5-4860-80ba-f6e584adaf14"
specification: |
  In home mode (`root = homeRoot`) the root SHALL be resolved from the supplier determined
  by the layer. Standalone SHALL resolve it from `$HOME`, the shell environment variable
  at run time. The home-manager module SHALL resolve it from `$HOME` as resolved
  internally by home-manager, the module pinning `homeRoot`. Both SHALL work irrespective
  of the operating system.
specification_ja: |
  home mode（`root = homeRoot`）の root は、層ごとに定まった供給元から解決しなければならない。
  CLI 単体は実行時のシェル環境変数 `$HOME`、home-manager モジュールは HM が内部解決した
  `$HOME`（モジュールが `homeRoot` を pin する）から解決しなければならず、いずれも OS を
  問わず動作しなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:31:01+09:00"
---
# home mode の root は層ごとに定まった供給元から解決する

## 仕様

| 層 | root の解決方法 | 備考 |
|---|---|---|
| standalone | `$HOME`（実行時のシェル環境変数）| OS 問わず動作 |
| home-manager | `$HOME`（HM が内部解決・モジュールが `homeRoot` を pin）| OS 問わず動作 |

> **上は原文の写しを [NixOS / nix-darwin モジュールは root を systemRoot に固定する](../adr/system-modules-pin-systemroot.md) に合わせて改めたもので、規範は frontmatter が正**。`homeRoot` が marker であり `kind` を運ぶこと
> （実体パス解決は engine の実行時責務）は [root マーカーは `kind` を運ぶ入れ物でパス解決は engine が行う](root-markers-runtime-resolution.md) / [root は 3 マーカーと絶対パス文字列の union を取る](root-marker-kinds.md)、`--root` による全モード
> 共通の上書きは [--root は全モード共通で解決 root を明示上書きする](root-override-flag.md)、
> home mode の profileDir キーは [profileDir は home と system のみ name 直キーとし、fixed root と --root 上書きは roothash でキーする](profiledir-keying.md) の担当。**モジュールが `homeRoot` を pin する
> こと自体は [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](module-common-options.md) の担当**で、本 item の規範に現れる「モジュールが `homeRoot` を
> pin する」は、HM の `$HOME` が standalone の `$HOME` とは供給の処理が異なる（HM が内部
> 解決する）ことを説明する従属節であり、pin そのものを規範化するものではない。
>
> 原文の表にあった NixOS / nix-darwin 行（`config.users.users.${cfg.user}.home`）は規範に採らない。
> 両モジュールは home mode の root を供給せず `systemRoot` を pin する（→ [NixOS / nix-darwin モジュールは root を systemRoot に固定する](../adr/system-modules-pin-systemroot.md)・
> [全モジュールは共通オプションの同一集合を公開し、entries は configs 経由・root はモジュールが pin する](module-common-options.md)）。

## 出典

`docs/spec.md`「root の解決」→「home mode（`root = homeRoot`）」節の表。

決定の実体は root を 3 マーカーへ昇格し `$HOME` を `homeRoot` としたうえで各層が root を
供給すると定めた [root は 3 つの marker で表す](../adr/three-root-markers.md)と、各層を配線に徹させ root と activation タイミングだけを供給させる
[モジュールは root と起動の時機だけを供給して engine を起動する](../adr/modules-wire-engine-kick-only.md)。
