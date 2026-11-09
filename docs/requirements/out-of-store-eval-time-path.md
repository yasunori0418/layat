---
id: "REQ-81249072-56b8-42f5-807b-ea623c1efe99"
type: requirement
name: "out-of-store の src は評価時の絶対パスか配置 root 相対のどちらかで、root 相対は target と同じ root に対して実行時に解決する"
status: stable
derives_from:
  - "UC-01b896b4-04b9-40d0-bf9e-966eaf64c3d4"
specification: |
  The local path of an out-of-store entry SHALL take one of two forms. For
  `srcKind: "outOfStore"` it SHALL be an absolute path settled at Nix evaluation time
  ([mkOutOfStoreSymlink は out-of-store symlink を表すマーカーを返す](mkoutofstoresymlink-marker.md)) and SHALL be hard-coded into the manifest as
  it stands. Because the root of a `target` is resolved at run time by the engine, that
  absolute path SHALL NOT affect the resolution on the `target` side; the two SHALL belong
  to separate points in time and SHALL be independent. For `srcKind: "rootRelative"` it
  SHALL be a path relative to the placement root, recorded in the manifest as it stands,
  and SHALL be resolved at run time against the same root as the `target`
  ([mkRootRelativeSymlink は配置 root 相対の src を表すマーカーを返し、engine が target 起点の相対 symlink として張る](root-relative-symlink-marker.md)). Since the convention for the home directory
  differs between macOS and Linux, an absolute local path SHOULD be resolved by
  discriminating the operating system within the flake.
specification_ja: |
  out-of-store entry のローカルパスは次の 2 形のどちらかでなければならない。
  `srcKind: "outOfStore"` では Nix 評価時に確定する絶対パス（[mkOutOfStoreSymlink は out-of-store symlink を表すマーカーを返す](mkoutofstoresymlink-marker.md)）とし、
  そのまま manifest へハードコードされなければならない。`target` の root 解決は engine が
  実行時に行うため、その絶対パスは `target` 側の解決に影響してはならない（両者は別の時点に
  属し独立である）。`srcKind: "rootRelative"` では配置 root 相対のパスとし、そのまま
  manifest へ記録され、実行時に `target` と同じ root に対して解決されなければならない
  （[mkRootRelativeSymlink は配置 root 相対の src を表すマーカーを返し、engine が target 起点の相対 symlink として張る](root-relative-symlink-marker.md)）。macOS / Linux でホームの慣例が異なるため、絶対パスの
  ローカルパスは flake 内で OS を判別して解決するべきである。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# out-of-store の src は評価時の絶対パスか配置 root 相対のどちらかで、root 相対は target と同じ root に対して実行時に解決する

## 仕様

`mkOutOfStoreSymlink` の引数は Nix 評価時に確定する絶対パス。`$HOME` は使えない。
`target` の root 解決は実行時に行われるため `target` 側には影響しない。
`mkRootRelativeSymlink` の引数は配置 root 相対のパスで、実行時に `target` と同じ root に対して
解決される。

```
outOfStore    Nix 評価時:  mkOutOfStoreSymlink "/path/to/dotfiles"  →  manifest にハードコード
              実行時:      root を engine が解決                       →  target を絶対パス化
rootRelative  Nix 評価時:  mkRootRelativeSymlink "dev/skills/x"       →  manifest に相対のまま記録
              実行時:      root を engine が解決                       →  target と src を同じ root で解決
```

macOS / Linux でホームの慣例が異なるため、絶対パスのローカルパスは flake 内で OS 判別して
解決するのが推奨。`builtins.getEnv "HOME"`（`--impure` 必要）や flake の `specialArgs` 注入も
使えるが、通常は OS 判別で十分。

> **規範は frontmatter が正**。`mkOutOfStoreSymlink` がマーカーを返す純粋関数であることと
> 引数が評価時確定の絶対パスであること自体は [mkOutOfStoreSymlink は out-of-store symlink を表すマーカーを返す](mkoutofstoresymlink-marker.md)、
> `mkRootRelativeSymlink` のマーカーとリンクの形は [mkRootRelativeSymlink は配置 root 相対の src を表すマーカーを返し、engine が target 起点の相対 symlink として張る](root-relative-symlink-marker.md)、out-of-store entry
> の配置と版管理の範囲は [out-of-store symlink は marker の絶対パスまたは root 相対パスを指し、版管理はリンク先マッピングのみとする](out-of-store-link-mapping-only.md)、root マーカーの実行時解決は [root マーカーは `kind` を運ぶ入れ物でパス解決は engine が行う](root-markers-runtime-resolution.md) の担当。
> 本 item は「out-of-store の src が取る 2 形と、それぞれが target root の解決とどの時点で
> 関係するか」を規定する。

## 出典

`docs/spec.md`「root の解決」→「ローカルパス（out-of-store）の扱い」節（本文・時点対比の
コードブロック・OS 判別の推奨）。同節の `nix` 使用例は要求ではないため規範に採らない。

決定の実体は root を marker 化して実行時解決へ寄せた [root は 3 つの marker で表す](../adr/three-root-markers.md)と、out-of-store symlink をネイティブ
機構へ委譲せず engine が扱うと定めた [out-of-store のリンクは lib の関数が包むだけにして engine が張る](../adr/out-of-store-linked-by-engine.md)。root 相対の形は [root 相対マーカーの基準点を manifest の配置 root にする](../adr/root-relative-base-is-placement-root.md)が加えた。ただし OS 判別によるローカルパス解決の推奨（`SHOULD`）は原文由来で、
対応する ADR を持たない（macOS / Linux のホーム慣例差を扱った ADR は無い）。
