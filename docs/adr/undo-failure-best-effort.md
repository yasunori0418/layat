---
id: "ADR-a00bb2de-a8c5-4816-892b-4fa1f5299935"
type: adr
name: "巻き戻しの失敗は best-effort で続行する"
section: atomicity
status: stable
sources:
  - id: "issue-168"
    resource: "https://github.com/yasunori0418/layat/issues/168"
    title: "#168"
  - id: "issue-172"
    resource: "https://github.com/yasunori0418/layat/issues/172"
    title: "#172"
origin: "旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0044 §3 を書き直した"
justifies:
  - "REQ-9fca28c9-d3b1-4ad7-8f24-13b2ec7aeab2"
references:
  - "ADR-56325270-20bb-4959-8498-e3223c35e24b"
  - "ADR-74423bb8-1453-4689-a25e-30bf46536afc"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:26:59+09:00"
---
# 巻き戻しの失敗は best-effort で続行する

- ステータス: stable
- 日付: 2026-11-03
- 関連: [途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](in-memory-undo-journal.md), [新規配置と張替えを先に、世代のコミットを最後に行う](placement-order-commit-last.md), GitHub Issue #168, #172
- 起点: 旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0044 §3 を書き直した

## 背景

[途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](in-memory-undo-journal.md)。このジャーナルの逆操作も失敗しうる。たとえば
symlink を作り直そうとした親ディレクトリを、別のプロセスが消していることがある。1 件の
逆操作が失敗したときに残りをどう扱うかを決める必要があった。

## 決定

個々の逆操作が失敗しても巻き戻し全体は止めず、その項目を飛ばして残りの巻き戻しを続ける。
全てを試し終えたら、元のエラーと巻き戻せなかった項目の一覧を全件報告して停止する。

報告の形と出力先は [巻き戻し自体の失敗は best-effort で続行し、全件を stderr へ報告して停止する](../requirements/undo-rollback-best-effort.md) に置く。

## 根拠

1 件の失敗で巻き戻しをやめると、まだ戻せた項目まで戻らないまま残り、被害が広がる。失敗を
その 1 項目に閉じ込めて残りを戻すほうが、適用前の状態に近づく。

巻き戻せなかった項目を全件並べれば、利用者は手で確かめて直せる。conflict を 1 回の実行で
全件報告するのと同じ考え方である。

## 棄却した代替案

### 最初の失敗で巻き戻しを中断する

まだ戻せた項目まで戻らないまま残る。どこまで戻ったかも利用者には分かりにくい。
