---
id: "REQ-a8a923ad-07fb-4582-b90a-07a6e0c41baa"
type: requirement
name: "out-of-store symlink は marker の絶対パスまたは root 相対パスを指し、版管理はリンク先マッピングのみとする"
description: "out-of-store entry の symlink は marker の絶対パスか、root 相対の marker では target の親からの相対パスを指し、世代はどのパスを指すかだけを版管理し、指す先の内容を版管理してはならない"
section: placement-copy
status: stable
derives_from:
  - "UC-01b896b4-04b9-40d0-bf9e-966eaf64c3d4"
specification: |
  For an out-of-store entry, which is placed as a symlink pointing either at the absolute
  path held by the marker or, for a root-relative marker, at the relative path from the
  directory of the target to `<src>/<subpath>` that realises the marker's root-relative
  path ([symlink 配置は親 dir を作り配置元/subpath を指すリンクを張り、foreign symlink は警告して後勝ちする](symlink-placement-procedure.md), [mkRootRelativeSymlink は配置 root 相対の src を表すマーカーを返し、engine が target 起点の相対 symlink として張る](root-relative-symlink-marker.md)),
  generations SHALL version only the mapping of which
  path is pointed at, and SHALL NOT version the content of the destination, which is live
  by design and SHALL never be snapshotted.
specification_ja: |
  out-of-store entry は symlink として配置され、marker の絶対パスを指す。root 相対の
  marker では、target の親から `<src>/<subpath>` への相対パスを指す（marker の root 相対
  パスを相対リンクとして実現したもの。[symlink 配置は親 dir を作り配置元/subpath を指すリンクを張り、foreign symlink は警告して後勝ちする](symlink-placement-procedure.md)・[mkRootRelativeSymlink は配置 root 相対の src を表すマーカーを返し、engine が target 起点の相対 symlink として張る](root-relative-symlink-marker.md)）。
  この entry について、世代が版管理するのは「どのパスを指すか」というリンク先
  マッピングのみとし、指す先の内容は版管理してはならない（設計上ライブであり、永遠に
  スナップショットしない）。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:28:26+09:00"
---
# out-of-store symlink は marker の絶対パスまたは root 相対パスを指し、版管理はリンク先マッピングのみとする

## 仕様

- symlink として配置する。指す先は marker の絶対パスまたは root 相対パス（ローカル FS）。
  root 相対の張り方は [mkRootRelativeSymlink は配置 root 相対の src を表すマーカーを返し、engine が target 起点の相対 symlink として張る](root-relative-symlink-marker.md) に従う
- 世代では「どのパスを指すか」のリンク先マッピングのみ版管理する。指す先の内容は設計上
  ライブで、永遠にスナップショットしない

> **上は原文の写しに root 相対（→ [root 相対マーカーのリンクを target 起点の相対パスで張る](../adr/root-relative-link-as-relative-path.md)）を加えたもので、規範は frontmatter が正**。marker の生成と識別は
> [mkOutOfStoreSymlink は out-of-store symlink を表すマーカーを返す](mkoutofstoresymlink-marker.md) / [marker は判別タグで識別し manifest.json には漏らさない](marker-tag-discrimination.md)、method と src の対応表は [method は配置方法を選び symlink は世代管理下・copy は世代管理外になる](method-src-matrix.md)、
> symlink 配置の手順そのものは [symlink 配置は親 dir を作り配置元/subpath を指すリンクを張り、foreign symlink は警告して後勝ちする](symlink-placement-procedure.md) の担当。

## 出典

`docs/spec.md`「配置動作仕様」→「out-of-store symlink」節。

決定の実体は [out-of-store はマッピングだけを追跡する](../adr/out-of-store-tracks-mapping-only.md)で確定した、out-of-store のリンク先
マッピングのみを版管理する方針。
