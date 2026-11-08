---
id: "SDR-e093c6b2-75fd-4c91-a8f0-5a89e005e4ed"
type: sara_adr
name: "ADR も UUID で採番し、連番の旧 ADR は legacy_adr 型に凍結する"
status: deprecated
sources:
  - id: "issue-499"
    resource: "https://github.com/yasunori0418/layat/issues/499"
    title: "#499"
origin: "Issue #499 の grilling（2026-11-01）で確定。ADR-0048 §4・ADR-0053 §4 を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
references:
  - "SDR-1afa3a31-224d-492d-96d9-d3f6c9cb20cd"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:24:32+09:00"
---
# SDR-e093c6b2-75fd-4c91-a8f0-5a89e005e4ed: ADR も UUID で採番し、連番の旧 ADR は legacy_adr 型に凍結する

> **廃止（2026-11-04）**: 本 SDR の決定は次の後継へ移った。以下の本文は当時の記録として読む。
>
> - 全体 → [ADR は UUID だけで採番する](adr-uuid-only-numbering.md)

- ステータス: deprecated
- 日付: 2026-11-01
- 関連: [ADR は単独で覆せる 1 つの決定だけを書く](adr-one-decision-granularity.md), GitHub Issue #499, `docs/adr/README.md`
- 起点: Issue #499 の grilling（2026-11-01）で確定。ADR-0048 §4・ADR-0053 §4 を書き直した

## 背景

ADR だけは連番（`ADR-NNNN`）で採番してきた。既存 ADR の相互参照やコミットメッセージでの言及を
無効にしないためである。一方、並列レーンで ADR を起こすと、複数のレーンが同じ「次の番号」を取る構造が
残っていた。

1 決定 1 ADR の粒度（[ADR は単独で覆せる 1 つの決定だけを書く](adr-one-decision-granularity.md)）で旧 ADR を書き直すと、新しい ADR は
100 本を超え、多くのレーンが並行して起票する。sara は型ごとに 1 つの `id_format` しか持てず、連番の
ID は UUID 形式の検証を通らない。

## 決定

新しく起こす ADR は他の item と同じく UUID で採番する（`ADR-<フル UUIDv4>`）。既存の連番 ADR は、
`id_format` を連番のまま持つ別の型 `legacy_adr` へ付け替えて凍結する。`legacy_adr` は新規に起こさず、
廃止にするときを除いて本文も書き換えない。

型・ファイル名・起票の手順と `legacy_adr` の扱いは `docs/adr/README.md` に置く。

## 根拠

UUID なら並列レーンで同じ ID を取り合わない。マージ後の振り直しと参照の書き換えが構造的に起きない。

旧 ADR を連番のまま別の型に残せば、既存の ID・ファイル名・コミットメッセージでの言及はそのまま
有効になる。連番から UUID への読み替え表を持たずに済む。

## 棄却した代替案

### 旧 ADR も UUID へ移行する

コミットメッセージに残る連番の言及は書き換えられない。読み替え表を恒久的に持つことになるので
採らない。

### ADR だけ連番を続け、衝突はマージ後の `sara check` で検出する

レーン数が増えると衝突のたびに振り直しが起きる。採番の時点で衝突しない方式を選ぶ。
