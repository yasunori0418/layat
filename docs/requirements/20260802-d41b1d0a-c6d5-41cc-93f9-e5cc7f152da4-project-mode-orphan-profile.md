---
id: "REQ-d41b1d0a-c6d5-41cc-93f9-e5cc7f152da4"
type: requirement
name: "孤児 profile は backref で逆引き可能なまま放置許容とし、`layat prune` が root 不在の系列を削除する"
derives_from:
  - "UC-19a90989-0ae3-438f-8a75-4e1e2637f81c"
specification: |
  When a clone is deleted, its profile SHALL remain as an orphan under
  `<state>/nix/profiles/layat/`. The store SHALL be freed by `nix-collect-garbage` but the
  profile directory SHALL remain; that SHALL be tolerated until it is removed by
  `layat prune` or by hand, and SHALL be noted in the public documentation. Which root an
  orphan originated from SHALL be recoverable from the backref file `.root` at the
  `<roothash>` level. `layat prune` SHALL resolve orphan series through that backref and
  SHALL remove those whose root no longer exists; which series it selects, and the gates
  that guard the deletion, SHALL be as the two requirements devoted to them state.
specification_ja: |
  クローンを削除すると profile は `<state>/nix/profiles/layat/` 下に孤児として残らなければ
  ならない。store は `nix-collect-garbage` で解放されるが profile ディレクトリは残らなければ
  ならない。これは `layat prune` または手動で削除されるまで放置許容としなければならず、
  公開ドキュメントに注記しなければならない。どの root 由来の孤児かは `<roothash>` 階層の
  backref ファイル `.root` で逆引きできなければならない。`layat prune` はこの backref で
  孤児系列を逆引きし、root がもはや実在しない系列を削除しなければならない。どの系列を
  対象とするか、および削除を守るゲートは、それぞれを担当する 2 件の要求が定めるとおりで
  なければならない。
---
# REQ-d41b1d0a-c6d5-41cc-93f9-e5cc7f152da4: 孤児 profile は backref で逆引き可能なまま放置許容とし、`layat prune` が root 不在の系列を削除する

## 仕様

**orphan profile**: クローンを削除すると profile が `<state>/nix/profiles/layat/` 下に孤児として
残る。store は `nix-collect-garbage` で解放されるが profile ディレクトリは残る。`layat prune` か
手動で削除されるまでは放置許容とし、公開ドキュメントに注記する。どの root 由来の孤児かは
`<roothash>` 階層の backref ファイル（`.root`）で逆引きできる。**cleanup は `layat prune` が担う**
（実在しない root を指す孤児系列を backref で逆引きして削除する）。

> **規範は frontmatter が正**。backref `.root` を roothash 階層へ置くこと自体と
> `.pending` が config あたり最大 1 であることは REQ-2aa3abbc-90b2-486e-92de-d785554bdeb3、store 解放に
> `nix-collect-garbage` を使うことは REQ-706de717-4e47-471a-a1c0-448635be159c、profile を解決済み root でキーすること
> （孤児が生じる前提）は REQ-46fccb80-4bae-4d37-bc19-dded88e9a9c0 の担当。`layat prune` の対象判定は
> REQ-c44433a1-7ee7-459a-9aae-7cc42166876f、削除を守るゲート（dryrun・確認・try-lock・`--json`）は
> REQ-42fe312c-927c-4da3-9346-f7ca2f3a58ed が持つ。

## 出典

`docs/spec.md`「世代管理仕様」→「project mode の世代」節の箇条書き最終項。

決定の実体は ADR-0005「project mode（プロジェクト相対配置）と ephemeral 配置原則」（孤児 profile の
放置許容）で、backref による逆引きは ADR-0013 が定めている。cleanup を持たず seam に留める判断は
ADR-0024 にあったが、ADR-0034 がこれを改訂し `layat prune` の実装決定へ進めた。
