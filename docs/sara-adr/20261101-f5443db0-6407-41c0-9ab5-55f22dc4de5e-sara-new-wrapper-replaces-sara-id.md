---
id: "SDR-f5443db0-6407-41c0-9ab5-55f22dc4de5e"
type: sara_adr
name: "sara-id を廃止し sara init を包む起票ラッパーを使う"
status: stable
issues:
  - "#366"
  - "#364"
origin: "epic #364 の着手前 grilling（2026-08-17）で確定。ADR-0053 §3 を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
references:
  - "SDR-722ce265-2aaf-4a90-915d-32d26b7a38bb"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:32+09:00"
---
# SDR-f5443db0-6407-41c0-9ab5-55f22dc4de5e: sara-id を廃止し sara init を包む起票ラッパーを使う

- ステータス: stable
- 日付: 2026-11-01
- 関連: SDR-722ce265-2aaf-4a90-915d-32d26b7a38bb, GitHub Issue #366, #364
- 起点: epic #364 の着手前 grilling（2026-08-17）で確定。ADR-0053 §3 を書き直した

## 背景

`sara-id` は devShell 同梱の採番コマンドで、UUIDv4 の採番・8 文字の重複の確認・省略形の出力・型名から
prefix への写像・ファイル名素材の組み立てを担っていた。item ファイルの作成は `sara init` が担当した。

sara 0.10.0 から、`sara init` が `id_format` に従って UUIDv4 を採番するようになった。ID をフル UUID に
統一した（SDR-722ce265-2aaf-4a90-915d-32d26b7a38bb）ことで、8 文字の重複の確認と省略形の出力も
要らなくなった。

## 決定

`sara-id` を廃止する。後継として、`sara init` の呼び出しを前後から包む薄い起票ラッパー `sara-new` を
使う。採番は `sara init` に任せ、ラッパーは ID を生成しない。ラッパーが担うのは、ファイル名規約
（`<YYYYMMDD>-<フル UUIDv4>-<slug>.md`）の適用だけである。

ラッパーは `dev/scripts/` の独立したファイルに置き、契約テストを持つ。test_plan は起こさない。

## 根拠

採番を sara に一本化すれば、採番のロジックを二重に持たずに済む。ラッパーに残る責務はファイル名の
組み立てだけで、守るべき規範を持たない。逆算の階層（test_plan）を起こすほどのものではない。

devShell の定義に埋め込まず独立したファイルにしておけば、別のリポジトリへ持ち出しやすい。

## 棄却した代替案

### `sara-id` を残し、責務を重複の確認へ寄せる

省略形を使わない以上、重複を確認する対象が無い。

### ラッパーを持たず、`sara init` の後に手で改名する

ファイル名規約の適用が手作業になり、間違えやすい。
