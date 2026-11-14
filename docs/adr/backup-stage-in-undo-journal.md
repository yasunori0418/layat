---
id: "ADR-3f97fc1d-188d-4013-bff0-16b67250f68e"
type: adr
name: "退避を配置前除去の後の段として undo ジャーナルに載せる"
section: removal-migration
status: stable
sources:
  - id: "issue-169"
    resource: "https://github.com/yasunori0418/layat/issues/169"
    title: "#169"
  - id: "issue-172"
    resource: "https://github.com/yasunori0418/layat/issues/172"
    title: "#172"
origin: "旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §4（undo ジャーナルへの統合）を書き直した"
justifies:
  - "REQ-9b0046e0-8ddc-4c0b-940e-3fe6f36d0e98"
  - "REQ-5e75aabc-0e8f-4a6c-92bd-a712dc68a940"
references:
  - "ADR-56325270-20bb-4959-8498-e3223c35e24b"
  - "ADR-c42a51ff-fc51-4d19-b8c6-2b79704dd30f"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:26:59+09:00"
---
# 退避を配置前除去の後の段として undo ジャーナルに載せる

- ステータス: stable
- 日付: 2026-11-03
- 関連: [途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](in-memory-undo-journal.md), [--backup は配置を塞ぐ記録外の実体だけを退避する](backup-scope-unrecorded-entities.md), GitHub Issue #169, #172
- 起点: 旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §4（undo ジャーナルへの統合）を書き直した

## 背景

[--backup は配置を塞ぐ記録外の実体だけを退避する](backup-scope-unrecorded-entities.md)。退避は rename で行う。
[途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](in-memory-undo-journal.md)。退避を apply のどの段で行い、途中失敗
したときにどう扱うかを決める必要があった。

## 決定

退避は、plan の独立した段として、配置前除去の後・配置の前に engine が実行する。退避の rename は
undo ジャーナルに記録し、後続の段が失敗したら退避物を元の場所へ rename で戻す。

段の位置と対象の括り方は [backup 退避は配置前除去の後・配置の前に置き、drift 修復の処理でも同じく実施する](../requirements/backup-stage-position.md) に、退避の逆操作は
[途中失敗した apply / rollback はインメモリ undo ジャーナルで全 FS 変更を巻き戻す](../requirements/undo-journal-rollback.md) に置く。

## 根拠

退避は配置より前に済んでいなければならない。配置前除去が、自己記録の古くなったものを
片付けた後に置けば、退避するのは片付けた後もなお配置を塞ぐ記録外の実体だけになる。

退避も apply の中で行う FS 変更である。ジャーナルに載せれば、後続の段で失敗したときに
利用者のファイルが退避先に移ったまま残ることがない。既存の巻き戻しの機構でそのまま巻き戻される。

## 棄却した代替案

### 退避をジャーナルの外で行う

後続の段が失敗すると、利用者のファイルが退避先に移ったまま apply が失敗で終わる。失敗した
apply は FS に痕跡を残さないという意味論が崩れる。
