---
id: "SDR-ba95c0c6-ea15-4e97-95a9-925ac04df58d"
type: sara_adr
name: "sara-id を廃止し sara init を包む起票ラッパーで slug のファイル名を適用する"
description: "採番を `sara init` に任せ、薄い起票ラッパー `sara-new` が `<slug>.md` のファイル名規約の適用、`--description` の必須チェック、`generated` の付与を担う。"
section: ids-and-filenames
status: stable
sources:
  - id: "issue-366"
    resource: "https://github.com/yasunori0418/layat/issues/366"
    title: "#366"
  - id: "issue-364"
    resource: "https://github.com/yasunori0418/layat/issues/364"
    title: "#364"
  - id: "issue-692"
    resource: "https://github.com/yasunori0418/layat/issues/692"
    title: "epic: 散文リンク移行の後始末 — 表示崩れ・重複表記・文の破綻・frontmatter のフル ID"
origin: "epic #692 の grilling（2026-11-08）で確定。[sara-id を廃止し sara init を包む起票ラッパーを使う](sara-new-wrapper-replaces-sara-id.md) を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
  - "QA-1dd63789-24ed-4a86-bf12-b62d8c526ba5"
supersedes:
  - "SDR-f5443db0-6407-41c0-9ab5-55f22dc4de5e"
references:
  - "SDR-722ce265-2aaf-4a90-915d-32d26b7a38bb"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T09:15:32+09:00"
---
# sara-id を廃止し sara init を包む起票ラッパーで slug のファイル名を適用する

- ステータス: stable
- 日付: 2026-11-08
- 関連: [ID はフル UUID に統一し 8 文字省略形を使わない](full-uuid-ids-no-short-form.md), GitHub Issue #366, #364, #692
- 起点: epic #692 の grilling（2026-11-08）で確定。[sara-id を廃止し sara init を包む起票ラッパーを使う](sara-new-wrapper-replaces-sara-id.md) を書き直した

## 背景

`sara-id` は devShell 同梱の採番コマンドで、UUIDv4 の採番・8 文字の重複の確認・省略形の出力・型名から
prefix への写像・ファイル名素材の組み立てを担っていた。item ファイルの作成は `sara init` が担当した。

sara 0.10.0 から、`sara init` が `id_format` に従って UUIDv4 を採番するようになった。
[ID はフル UUID に統一し 8 文字省略形を使わない](full-uuid-ids-no-short-form.md) ので、8 文字の重複の確認と
省略形の出力も要らない。

## 決定

`sara-id` を廃止する。後継として、`sara init` の呼び出しを前後から包む薄い起票ラッパー `sara-new` を
使う。採番は `sara init` に任せ、ラッパーは ID を生成しない。ラッパーが担うのは、ファイル名規約
（`<slug>.md`）の適用、`--description` の必須チェック、`generated` の付与である。

ラッパーは `dev/scripts/` の独立したファイルに置き、契約テストを持つ。test_plan は起こさない。

## 根拠

採番を sara に一本化すれば、採番のロジックを二重に持たずに済む。ラッパーに残る責務はファイル名の
組み立て、`--description` の必須チェック、`generated` の付与である。いずれも別の決定が定めた規約を起票時に
適用するもので、ラッパー自身は守るべき規範を持たない。逆算の階層（test_plan）を起こすほどのものではない。

devShell の定義に埋め込まず独立したファイルにしておけば、別のリポジトリへ持ち出しやすい。

## 棄却した代替案

### `sara-id` を残し、責務を重複の確認へ寄せる

省略形を使わない以上、重複を確認する対象が無い。

### ラッパーを持たず、`sara init` の後に手で改名する

ファイル名規約の適用が手作業になり、間違えやすい。
