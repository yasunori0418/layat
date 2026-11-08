---
id: "REQ-840b3641-6e76-46da-82e9-680cabd65abe"
type: requirement
name: "失敗時に残る pending gcroot は config あたり最大 1 個に有界とし回収処理を持たない"
status: stable
derives_from:
  - "UC-0b6f60cb-3e98-4ee7-8929-4d94a29f0af6"
specification: |
  When an apply fails before reaching the generation commit, the `<profileDir>/.pending`
  out-link (a gcroot under the default nix generator) remains and keeps holding a built
  but unused link-farm. Because the next apply overwrites it under the same name
  (`.pending`), it SHALL be bounded to at most one per config, and no reclamation process
  SHALL be provided; this SHALL be accepted. After a successful commit the engine SHALL
  remove `.pending` as a plain filesystem operation.
specification_ja: |
  世代 commit 到達前に apply が失敗すると `<profileDir>/.pending` out-link（既定の nix
  生成器では gcroot）が残り、ビルド済み未使用 link-farm を掴み続ける。次回 apply が同名
  （`.pending`）で上書きするため config あたり最大 1 個に有界でなければならず、回収処理を
  持ってはならない。これは許容しなければならない。commit 成功後は engine が `.pending` を
  素の FS 操作で除去しなければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T23:05:37+09:00"
---
# REQ-840b3641-6e76-46da-82e9-680cabd65abe: 失敗時に残る pending gcroot は config あたり最大 1 個に有界とし回収処理を持たない

## 仕様

世代 commit（実行フローの f）到達前に apply が失敗すると `<profileDir>/.pending` out-link
（既定の nix 生成器では gcroot）が残り、ビルド済み未使用 link-farm を掴み続けるが、次回 apply
が**同名**（`.pending`）で上書きするため config あたり最大 1 個に有界。回収処理は持たず
許容する。commit 成功後の `.pending` の除去は build 出力の一時リンクの片付けで、engine の
素の FS 操作（gcroot の概念を持たない）。

pending out-link を張る段そのものは [実行フローの順序は eval 先行 → flock → build とし build をロック内に閉じる](exec-flow-eval-lock-build.md) の担当。

## 出典

`docs/spec.md`「CLI 仕様」→「実行フロー」の `.pending` gcroot 残存の箇条書き。

決定の実体は [.pending の残留を許容する](../adr/pending-leftover-tolerated.md)。commit 後の片付けを engine の FS 操作として
位置づけたのは [.pending の後片付けを gcroot を知らない engine の FS 操作として行う](../adr/pending-cleanup-without-gcroot.md)。
