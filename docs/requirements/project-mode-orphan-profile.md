---
id: "REQ-d41b1d0a-c6d5-41cc-93f9-e5cc7f152da4"
type: requirement
name: "孤児 profile は backref で逆引き可能なまま放置許容とし、`layat prune` が root 不在の系列を削除する"
status: stable
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
  対象とするか、および削除を守る条件は、それぞれを担当する 2 件の要求が定めるとおりで
  なければならない。
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T22:43:18+09:00"
---
# REQ-d41b1d0a-c6d5-41cc-93f9-e5cc7f152da4: 孤児 profile は backref で逆引き可能なまま放置許容とし、`layat prune` が root 不在の系列を削除する

## 仕様

**孤児 profile**: クローンを削除すると profile が `<state>/nix/profiles/layat/` 下に孤児として
残る。store は `nix-collect-garbage` で解放されるが profile ディレクトリは残る。`layat prune` か
手動で削除されるまでは放置許容とし、公開ドキュメントに注記する。どの root 由来の孤児かは
`<roothash>` 階層の backref ファイル（`.root`）で逆引きできる。**cleanup は `layat prune` が担う**
（実在しない root を指す孤児系列を backref で逆引きして削除する）。

> **規範は frontmatter が正**。backref `.root` を roothash 階層へ置くこと自体と
> `.pending` が config あたり最大 1 であることは [profileDir は config 専用ディレクトリとし、profile リンク・世代・pending out-link をその中に並べる](profile-on-disk-layout.md)、store 解放に
> `nix-collect-garbage` を使うことは [世代操作は世代バックエンドの契約越しに行い、GC root の間引きと store 解放は nix コマンドへ分けて委譲する](generation-gc.md)、profile を解決済み root でキーすること
> （孤児が生じる前提）は [project mode は世代を非公開にし、project / system mode は derivation 同一なら世代を積まず lstat ドリフト修復だけ行う](project-mode-generation-skip.md) の担当。`layat prune` の対象判定は
> [prune は backref の root が実在しない roothash 系列だけを系列ごと削除する](prune-target-selection.md)、削除を守る条件（dryrun・確認・try-lock・`--json`）は
> [prune は dryrun・root 一覧付き確認・try-lock skip・--json の --yes 必須で削除を守る](prune-safety-gates.md) が持つ。

## 出典

`docs/spec.md`「世代管理仕様」→「project mode の世代」節の箇条書き最終項。

決定の実体は [profile は解決済み root の roothash でキーする](../adr/profile-keyed-by-roothash-of-resolved-root.md)（孤児 profile の
放置許容）で、backref による逆引きも同じ ADR が定めている。cleanup を持たず seam に留めていた判断は、
[prune は root が実在しない roothash 系列だけを削除する](../adr/prune-only-missing-root-series.md)で `layat prune` の実装に進めた。
