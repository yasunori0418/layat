---
id: "ADR-97248450-c19e-4857-94fb-2a10c2b1d80f"
type: adr
name: "世代スキップのときも退避を行う"
status: stable
sources:
  - id: "issue-169"
    resource: "https://github.com/yasunori0418/layat/issues/169"
    title: "#169"
  - id: "issue-172"
    resource: "https://github.com/yasunori0418/layat/issues/172"
    title: "#172"
origin: "旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §4（世代スキップ時）を書き直した"
justifies:
  - "REQ-9b0046e0-8ddc-4c0b-940e-3fe6f36d0e98"
references:
  - "ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d"
  - "ADR-3f97fc1d-188d-4013-bff0-16b67250f68e"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:26:59+09:00"
---
# ADR-97248450-c19e-4857-94fb-2a10c2b1d80f: 世代スキップのときも退避を行う

- ステータス: stable
- 日付: 2026-11-03
- 関連: [project mode では同じ link farm なら世代を積まず drift だけを直す](project-mode-generation-skip-and-drift-repair.md), [退避を配置前除去の後の段として undo ジャーナルに載せる](backup-stage-in-undo-journal.md), GitHub Issue #169, #172
- 起点: 旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0045 §4（世代スキップ時）を書き直した

## 背景

[project mode では同じ link farm なら世代を積まず drift だけを直す](project-mode-generation-skip-and-drift-repair.md)。世代スキップでは配置前除去が構造的に空になる。
配置前除去の発生源は全て manifest の差分を伴い、link farm を変えるからである。退避の段
（[退避を配置前除去の後の段として undo ジャーナルに載せる](backup-stage-in-undo-journal.md)）も世代スキップで空とみなしてよいかを決める必要が
あった。

## 決定

世代スキップのときも `--backup` を有効にし、通常の apply と同じく退避してから配置する。

世代スキップの有無ごとの段の規範は [backup 退避は配置前除去の後・配置の前に置き、drift 修復の処理でも同じく実施する](../requirements/backup-stage-position.md) に置く。

## 根拠

退避が起きる条件は、target に記録外の実体があることである。これは manifest や link farm の
変化を前提としない。shell に入り直す間に他のツールが target にファイルを置く、といった FS の
出来事だけで起こりうる。配置前除去が空になる不変条件は、退避には当てはまらない。

drift を直すたびに退避が行われなくなると、shell に入り直したら安全のための機構が
働かなかった、という驚きを生む。

## 棄却した代替案

### 世代スキップのときは退避の段を省く

配置前除去と同じく空とみなせば処理が単純になるが、記録外の実体の出現は manifest と無関係に
起こる。退避されずに conflict で止まるか、`--backup` を付けたのに退避されない結果になる。
