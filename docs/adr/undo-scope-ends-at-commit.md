---
id: "ADR-f29d1256-5e14-4b83-ac77-4717b2210193"
type: adr
name: "undo の発火を FS 反映の段に限り、世代コミットの後は対象にしない"
section: atomicity
status: stable
sources:
  - id: "issue-168"
    resource: "https://github.com/yasunori0418/layat/issues/168"
    title: "#168"
  - id: "issue-172"
    resource: "https://github.com/yasunori0418/layat/issues/172"
    title: "#172"
origin: "旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0044 §2 を書き直した"
justifies:
  - "REQ-5e75aabc-0e8f-4a6c-92bd-a712dc68a940"
  - "TP-deb05610-44bc-4962-8939-952392e5fbd0"
references:
  - "ADR-74423bb8-1453-4689-a25e-30bf46536afc"
  - "ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d"
  - "ADR-56325270-20bb-4959-8498-e3223c35e24b"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T08:26:59+09:00"
---
# undo の発火を FS 反映の段に限り、世代コミットの後は対象にしない

- ステータス: stable
- 日付: 2026-11-03
- 関連: [新規配置と張替えを先に、世代のコミットを最後に行う](placement-order-commit-last.md), [project mode では同じ link farm なら世代を積まず drift だけを直す](project-mode-generation-skip-and-drift-repair.md), [途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](in-memory-undo-journal.md), GitHub Issue #168, #172
- 起点: 旧 epic #167 から統合された epic #172 の grilling セッション（2026-07-11・2026-07-12）で確定。ADR-0044 §2 を書き直した

## 背景

[途中失敗した apply をインメモリ undo ジャーナルで巻き戻す](in-memory-undo-journal.md)。巻き戻す対象は run の FS 変更である。どこで起きた失敗を巻き戻しの対象にし、どの時点でジャーナルを捨てるかを決める
必要があった。[新規配置と張替えを先に、世代のコミットを最後に行う](placement-order-commit-last.md)。世代のコミットは FS への反映が全て成功した後の唯一のコミット点である。

## 決定

巻き戻しを発火させるのは、FS に反映する段（配置前除去・退避・配置・copy の反映・古くなった配置の除去）、
世代スキップ時の drift 修復（[project mode では同じ link farm なら世代を積まず drift だけを直す](project-mode-generation-skip-and-drift-repair.md)）、last-applied 記録の
書き込みで起きた失敗に限る。世代のコミットが成功したらジャーナルを捨てる。コミット自体が
失敗したときは巻き戻さない。

発火の範囲とジャーナルを捨てる時点の細目は [途中失敗した apply / rollback はインメモリ undo ジャーナルで全 FS 変更を巻き戻す](../requirements/undo-journal-rollback.md) に、
コミットの失敗で配置を巻き戻さないことの検証は [原子性は実 FS の条件で故障を誘発して不変条件ごとに検証する](../test-plan/fault-injection-atomicity.md) に置く。

## 根拠

コミットの成功は、その run の FS 変更が新しい世代として確定したことを意味する。その後で FS を
戻すのは、途中失敗の巻き戻しではなく別の世代への切り替えで、rollback の役割である。
ジャーナルが世代の境界を越えて FS を戻せると、両者の意味が曖昧になる。

コミットが失敗した時点では、FS への変更は全て成功している。FS は新しい manifest と整合して
いるので、戻すとかえって整合を壊す。世代が進んでいないだけなので、冪等な再実行で収束する。

## 棄却した代替案

### コミットの失敗でも巻き戻す

整合している FS を壊してまで、適用前の状態へ戻す理由が無い。

### コミット後もジャーナルを持ち越し、後から取り消せるようにする

途中失敗の巻き戻しと、コミット済みの世代からの rollback の境界が曖昧になる。
