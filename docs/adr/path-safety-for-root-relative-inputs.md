---
id: "ADR-d89c2ede-1aba-46fa-9f14-8dcf8a266e0e"
type: adr
name: "root 相対で指定するパスすべてにパス安全性の確認を課す"
status: stable
origin: "次期マイルストーン追加計画の grilling（2026-07-04）と Issue #362 の grilling（2026-10-12）。ADR-0040 §2b と ADR-0059 §6a を書き直した"
justifies:
  - "REQ-08c8f639-790c-4970-a954-c9567538cf9b"
  - "REQ-a77181dd-a59c-40bb-855e-3430ad970eb6"
references:
  - "ADR-00cc3497-654f-420a-928f-99f201e7efb0"
  - "ADR-400cadd4-0151-470b-92ae-9b11aa8e9bed"
  - "ADR-5e46ebd6-e44b-48dc-a93f-a8a3a3903967"
  - "ADR-6d2327f7-f513-42c3-89f8-808467db1914"
  - "ADR-b06dd8ab-317e-41ee-9c80-dd70dcd7f2fd"
  - "ADR-eeeb9dc8-9089-43ec-946a-c3ed76e98137"
  - "ADR-f40c3d37-638d-420a-a0fe-3908cbf0aec0"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:28:06+09:00"
---
# root 相対で指定するパスすべてにパス安全性の確認を課す

- ステータス: stable
- 日付: 2026-11-03
- 関連: [絶対パスと .. エスケープは eval 時に拒否する](path-safety-eval-rejection.md), [root は 3 つの marker で表す](three-root-markers.md), [copy と out-of-store の併用を eval 時に拒否する](copy-with-out-of-store-eval-error.md), [out-of-store はマッピングだけを追跡する](out-of-store-tracks-mapping-only.md), [manifest.json を schemaVersion 1 の唯一の契約として固定する](manifest-json-contract-schemaversion-v1.md), [layat.lib.mkEnv で buildEnv 合成環境を提供する](mkenv-buildenv.md), [home mode の CLI と lib の版ずれを manifest の確認に任せて許容する](home-mode-skew-guarded-by-schemaversion.md), `docs/spec.md`
- 起点: 次期マイルストーン追加計画の grilling（2026-07-04）と Issue #362 の grilling（2026-10-12）。ADR-0040 §2b と ADR-0059 §6a を書き直した

## 背景

target と subpath は、絶対パスと `..` による外への脱出を評価時に拒否している
（[絶対パスと .. エスケープは eval 時に拒否する](path-safety-eval-rejection.md)）。その後、root から相対に指定するパスが 2 つ増えた。
root マーカーの `subdir` と、root 相対マーカーの `src` である。どちらも root の下を指すつもりの
値で、確認しなければ root の外を指せる。

## 決定

root マーカーの `subdir` と root 相対マーカーの `src` にも、target と subpath と同じパス安全性の
確認を課す。絶対パスと、正規化すると `..` で root の外へ出る値を、`normalizeManifest` が評価時に
拒否する。それぞれの規則は [root マーカーの subdir 適用形は全 target への前置と等価な評価時の糖衣で、root を動かさない](../requirements/root-marker-subdir-sugar.md) と
[mkRootRelativeSymlink は配置 root 相対の src を表すマーカーを返し、engine が target 起点の相対 symlink として張る](../requirements/root-relative-symlink-marker.md) に置く。

## 根拠

root の外への書き込みや root の外を指すリンクは、保守的な古くなった配置の除去が守る「root の下にしか
触れない」という性質を破る。どれも root の実際の値に依らず静的に判定できるので、target と同じ
確認を同じ時点で行えば、指定するパスごとに安全性の基準が変わらない。

## 棄却した代替案

### `subdir` で root より上（外）へのずらしを許す

target の確認が拒否している root の外への配置を、意図して開ける抜け道になる。具体的な用途も
見つからなかった。grilling で、root の下を指す意図だと確かめた。

### root 相対の `src` は engine の実行時にだけ確認する

評価時に弾ける誤りの発見が、`--dryrun` や CI まで遅れる。
