---
id: "REQ-16aef46b-7bb8-4ca1-b962-e9f3ed1fd1d2"
type: requirement
name: "古くなった配置の除去は前世代の記録通りを指す symlink のみに限り、copy は消さず孤児を警告する"
status: stable
derives_from:
  - "UC-0b6f60cb-3e98-4ee7-8929-4d94a29f0af6"
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
  - "UC-403fbe32-b146-401b-8b53-fe67c1e169c5"
specification: |
  Removal SHALL be conservative. The baseline manifest SHALL be the last-applied record
  when one exists and otherwise the manifest of the current generation of the profile
  ([apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](last-applied-record.md)). Only a symlink that the baseline manifest
  records as having been placed by layat, and that still points at the destination so
  recorded — the store path that the baseline records, or the recorded out-of-store path — SHALL
  be removed. Regular files and links not managed by layat SHALL NOT be touched. Where the
  record and the actual entity disagree, the engine SHALL NOT remove it and SHALL warn.
  On a first run, where neither a last-applied record nor a generation exists and the
  baseline is therefore empty, nothing SHALL be removed. When an entry
  disappears, a symlink — whether a store link or an out-of-store one — SHALL be removed
  under those invariants, whereas a copy SHALL NOT be removed, being data owned by the
  user, and its being orphaned SHALL be notified by a warning.
specification_ja: |
  削除は保守的に行わなければならない。比較の基準のマニフェストは、last-applied 記録があれば
  記録、無ければ profile の current 世代のマニフェストとしなければならない
  （[apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](last-applied-record.md)）。比較の基準のマニフェストが「layat が配置した」と記録し、かつ
  現状もその記録通りの先（比較の基準が記録する store パス / 記録された out-of-store パス）を指す symlink
  のみを削除しなければならない。通常ファイルや layat 非管理の link には触れてはならない。
  記録と実体が不一致なら削除してはならず、警告しなければならない。初回（last-applied 記録も
  世代も無く比較の基準が空）では何も消してはならない。entry が消えたとき、symlink（store / out-of-store）は上記の不変条件を
  満たすもののみ除去しなければならず、copy は除去してはならない（ユーザー所有データの
  ため）。ただし copy が孤児になったことは警告で通知しなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T22:43:21+09:00"
---
# REQ-16aef46b-7bb8-4ca1-b962-e9f3ed1fd1d2: 古くなった配置の除去は前世代の記録通りを指す symlink のみに限り、copy は消さず孤児を警告する

## 仕様

削除は保守的に行う。前世代マニフェストが「layat が配置した」と記録し、**かつ現状もその記録通りの
先（その世代の store パス／記録された out-of-store パス）を指す symlink** のみ削除する。
通常ファイルや layat 非管理の link には触れない。記録と実体が不一致なら削除せず警告する。
初回／記録なしは何も消さない。

| 配置種別 | entry が消えたとき |
|---|---|
| symlink（store / out-of-store）| **除去する**（ただし上記の保守的不変条件を満たすもののみ）|
| copy | **除去しない**（ユーザー所有データ）。ただし孤児を警告で通知する |

> **上は原文の写しで、規範は frontmatter が正**。除去後の空親ディレクトリ剪定は
> [target 除去後は空の親ディレクトリチェーンを root 境界まで保守的に剪定する](empty-parent-dir-pruning.md)、除去に途中失敗したときの巻き戻しは [途中失敗した apply / rollback はインメモリ undo ジャーナルで全 FS 変更を巻き戻す](undo-journal-rollback.md) の担当。この保守的不変条件は
> `reset` の symlink 除去にも適用されるが（[reset は profile を触らない FS-only teardown で配置物を無い状態へ戻す](reset-fs-only-teardown.md)）、copy を明示的に消す唯一の手段が
> `reset` であることも同 item の担当。
>
> **2026-10-11 追記（[古くなった配置の除去は、比較の基準を記録から優先して読む](../adr/baseline-prefers-record.md)）**: 上の写しの「前世代マニフェスト」は比較の基準と読み替える。比較の基準は
> last-applied 記録があれば記録の manifest、無ければ profile の current 世代の manifest で、「初回／記録なし」は
> どちらも無く比較の基準が空のときを指す。記録の規範は [apply は最後に配置した manifest を last-applied 記録に書き、古くなった配置の除去の比較の基準は記録を優先して読む](last-applied-record.md) の担当。

## 出典

`docs/spec.md`「世代管理仕様」→「stale 除去の対象と安全不変条件」節の導入文と表。

決定の実体は [古くなった配置の除去は記録どおりの symlink だけに限る](../adr/conservative-removal-invariant.md)。比較の基準を
last-applied 記録から優先して読むのは [古くなった配置の除去は、比較の基準を記録から優先して読む](../adr/baseline-prefers-record.md)。
