---
id: "ADR-6885ec0e-01f8-46a2-a4e0-9e3580f640eb"
type: adr
name: "undo で戻せなかった target は `W_LAYAT_UNWIND_INCOMPLETE` で target ごとに報告し、`W_LAYAT_UNWOUND` は全部戻せたときだけ出す"
description: "undo の巻き戻しで戻せなかった target は W_LAYAT_UNWIND_INCOMPLETE で target ごとに --json へ報告し、W_LAYAT_UNWOUND は journal の全件を戻せたときだけ出す"
section: cli-output
status: stable
sources:
  - id: "issue-677"
    resource: "https://github.com/yasunori0418/layat/issues/677"
    title: "#677 Go のエラー処理を規約に揃える"
origin: "Issue #677 の grilling（2026-11-14）。undo が一部を戻せなくても W_LAYAT_UNWOUND を出す欠陥（#786）への対処として決めた"
justifies:
  - "REQ-2ea19863-eaa2-466b-b1ed-3f56f6417c62"
references:
  - "ADR-56325270-20bb-4959-8498-e3223c35e24b"
  - "ADR-a00bb2de-a8c5-4816-892b-4fa1f5299935"
  - "ADR-64cff3f4-5997-49f4-b23a-e6ebeced7df9"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-15T01:22:02+09:00"
---
# undo で戻せなかった target は `W_LAYAT_UNWIND_INCOMPLETE` で target ごとに報告し、`W_LAYAT_UNWOUND` は全部戻せたときだけ出す

- ステータス: stable
- 日付: 2026-11-15
- 関連: [途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](in-memory-undo-journal.md), [巻き戻しの失敗は best-effort で続行する](undo-failure-best-effort.md), [エラーコードは共通コードを再利用し固有の意味だけを E_LAYAT_* にする](error-code-scheme.md), GitHub Issue #677, #786
- 起点: Issue #677 の grilling（2026-11-14）。undo が一部を戻せなくても W_LAYAT_UNWOUND を出す欠陥（#786）への対処として決めた

## 背景

apply と rollback は、途中の段が失敗すると undo ジャーナルを逆順に戻す
（[途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](in-memory-undo-journal.md)）。逆操作が失敗しても残りを戻し続け、
戻せなかった項目を stderr へ報告する（[巻き戻しの失敗は best-effort で続行する](undo-failure-best-effort.md)）。

`--json` の subject 警告 `W_LAYAT_UNWOUND` は、差分がディスクに残っていないことを知らせる。この警告は
逆操作の成否を問わず、段が失敗すれば出ていた。journal が空で何も戻していないときにも出ていた。
逆操作が一部失敗して書き込みがディスクに残っても、`--json` の消費側には全部戻したと伝わり、
戻せなかった項目は stderr のテキストでしか分からなかった。

## 決定

戻せなかった target ごとに、subject 警告 `W_LAYAT_UNWIND_INCOMPLETE` を `detail = {target}` で出す。
target は root からの相対パスで、相対にできないときは絶対パスのまま出す。

`W_LAYAT_UNWOUND` は、journal の全件を戻せたときだけ出し、`detail` を持たない。journal が空のときは、
戻したものが無いので、どちらの警告も出さない。stderr の undo の報告も、journal が空のときは出さない。

この規範は [変更系の JSON ペイロードは engine 結果からフルインベントリと実差分を導く](../requirements/json-change-payload.md) に置く。

## 根拠

`W_LAYAT_UNWOUND` の値打ちは、出ていれば差分がディスクに残っていないと消費側が読めることにある。
一部を戻せなかったときにも出すと、残った書き込みを消費側が見落とす。

戻せなかったことは別の code で知らせる。そうすれば、`W_LAYAT_UNWOUND` を読んでいる既存の消費側は
意味を読み替えずに済む。固有の警告に `W_LAYAT_*` の code を足すのは
[エラーコードは共通コードを再利用し固有の意味だけを E_LAYAT_* にする](error-code-scheme.md) の方針どおりである。target ごとに出せば、
消費側は他の `W_LAYAT_*` と同じく `detail.target` で項目を引ける。

## 棄却した代替案

### `W_LAYAT_UNWOUND` に戻せなかった target の一覧を持たせる

新しい code を足さず、`W_LAYAT_UNWOUND` の `detail` に戻せなかった target の一覧を入れる案である。
code だけを見る消費側は、全部戻せたときと一部を戻せなかったときを区別できない。
`W_LAYAT_UNWOUND` の「差分がディスクに残っていない」という意味も成り立たなくなる。
