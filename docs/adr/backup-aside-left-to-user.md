---
id: "ADR-18267548-27d0-430b-8149-b6124d02ce56"
type: adr
name: "退避物を成功後も残して利用者の所有物にする"
status: stable
sources:
  - id: "issue-169"
    resource: "https://github.com/yasunori0418/layat/issues/169"
    title: "#169"
  - id: "issue-172"
    resource: "https://github.com/yasunori0418/layat/issues/172"
    title: "#172"
origin: "旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §5 と §4（成功時の扱いの非対称）を書き直した"
justifies:
  - "REQ-5dd5a4e9-6162-4fa5-b295-66844f5a4f3b"
  - "REQ-5e75aabc-0e8f-4a6c-92bd-a712dc68a940"
references:
  - "ADR-3f97fc1d-188d-4013-bff0-16b67250f68e"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:26:59+09:00"
---
# 退避物を成功後も残して利用者の所有物にする

- ステータス: stable
- 日付: 2026-11-03
- 関連: [退避を配置前除去の後の段として undo ジャーナルに載せる](backup-stage-in-undo-journal.md), GitHub Issue #169, #172
- 起点: 旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §5 と §4（成功時の扱いの非対称）を書き直した

## 背景

[退避を配置前除去の後の段として undo ジャーナルに載せる](backup-stage-in-undo-journal.md)。`--backup` の退避は途中失敗すれば元へ戻る。`--recopy` の上書きも rename で退避するが、
こちらの退避物は apply が成功すると削除する。`--backup` の退避物を成功後にどう扱うかを決める
必要があった。

## 決定

`--backup` の退避物は apply が成功しても削除しない。以後は layat の管理対象にせず、`reset` も
検出・復元しない。戻したいときは利用者が手で `mv` する。

成功時に残す規範は [途中失敗した apply / rollback はインメモリ undo ジャーナルで全 FS 変更を巻き戻す](../requirements/undo-journal-rollback.md) に、`reset` が復元しない規範は
[apply --backup は配置を塞ぐ記録外実体を rename 退避してから配置する](../requirements/apply-backup-rename-aside.md) に置く。

## 根拠

`--recopy` の退避は apply の中で完結する一時的な上書きの手順で、退避物は途中失敗に備えた
控えにすぎない。`--backup` の退避は、利用者へ成果物を引き渡すことに当たる。成功と同時に
消せば、データを失わずに確かめてから消せるという `--backup` の存在意義がなくなる。

layat が動かした事実は必ず見せるが、動かした後の所有権は利用者に戻す、という単純な分け方に
する。layat が退避物を追い続けると、記録外の実体を管理下に取り込むことになる。

## 棄却した代替案

### 成功時に退避物を自動で削除する（`--recopy` と同じ扱い）

apply の成功と同時に退避物が消え、確かめてから消すという目的を果たせない。

### `reset` で退避物を元へ戻す

退避物を追跡する状態を layat が持つことになり、記録外の実体を管理下に取り込む。
