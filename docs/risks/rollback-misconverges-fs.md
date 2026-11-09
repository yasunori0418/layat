---
id: "RISK-fbf029f6-866c-4a08-a4eb-f09e3c7e907e"
type: risk
name: "rollback が FS を前世代へ収束させ損ね、世代表示と実配置が食い違う"
status: stable
threatens:
  - "REQ-0e341430-17f0-498b-9439-65491652163a"
likelihood: high
impact: high
level: high
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:53:31+09:00"
---
# rollback が FS を前世代へ収束させ損ね、世代表示と実配置が食い違う

## リスク

layat は profile ディレクトリ自体ではなく任意 root へ配置するため、profile ポインタを
戻しただけでは FS は何も変わらない。`rollback` は「現世代 N の manifest を基準世代、
戻る世代 N-1 を target」として planner を回し、apply と同順（配置前除去 → 配置 / 張替え →
copy 反映 → 古くなった配置の除去）で FS を収束させてから、最後にポインタを移す
（→ [rollback は FS を先に収束させてから profile ポインタを最後に移す](../requirements/rollback-refit-then-pointer.md)）。

この収束が欠けると、`list-generations` は N-1 を current と表示するのに FS には N の
配置が残る、という嘘の状態になる。ポインタを先に動かすと更に悪く、次の apply が比較の基準を
N-2 と読んで古くなった配置の除去を誤る。

収束の失敗形は場面ごとに異なる。**祖先 symlink の migration** — 世代 N が配下 entries を
まとめて祖先 symlink へ移行していた場合、戻り先 N-1 の配下 entries を置くには先にその
祖先 symlink を除去する必要がある。plan の PreRemove を落とすと、symlink 越しに解決した
store パスへ書こうとして EEXIST / EROFS で失敗する（→ [配置前除去を planner が計算するデータとして表す](../adr/preremove-as-planner-data.md)、issue #173）。**途中失敗の
巻き戻し** — rollback も apply と同じ undo journal を使う必要があり、migration だけ済んで
配置が失敗した状態を残してはならない。**ポインタ移動の失敗** — 世代バックエンドの
`Switch` の失敗時点では FS の書き込みが全て成功しているため巻き戻さない（apply の commit 失敗と同じ
非対称性 → [undo の発火を FS 反映の段に限り、世代コミットの後は対象にしない](../adr/undo-scope-ends-at-commit.md)）。ここで unwind すると整合状態を壊す。

前提が満たされない場合にエラーを出さずに進むのも本リスクに含む。前世代が無い（最古世代）・profile が
そもそも無い（未 apply）・戻り先の target が foreign な実体で塞がれている、のいずれも
エラーで停止すべきであり、部分的に配置してから諦めるのが最悪の結果になる。

同じ FS 操作系でも `reset` の teardown は脅威が別で、あちらは「消してはいけないものを
消す」側にある（→ [reset の teardown が管理外の実体まで消し、preview と同意の条件も機能しない](reset-teardown-overreach.md)）。

## 実現性

**likelihood: high** — rollback は apply エンジンの再利用で成立しており、apply 側の
変更（PreRemove の追加・undo journal の導入）のたびに rollback 側の配線の張り忘れが実際に
発生している（→ issue #173、#168）。

**impact: high** — 誤収束が壊すのは、他の失敗から戻るための手段そのものである。ポインタを
先に動かして次の apply が比較の基準を N-2 と読めば古くなった配置の除去が記録外の実体へ届き、途中失敗で
祖先 symlink の migration だけが済んだ状態は元の実ディレクトリを失ったまま残る。どちらも
rollback をもう一度回しても戻らない。

## 緩和

[rollback が FS を前世代へ再収束させてから最後にポインタを移す](../test/generations/rollback-reconvergence.md)（rollback の再収束と祖先 migration）・[rollback は途中失敗を巻き戻し、ポインタ移動の失敗は巻き戻さず、前提不成立では停止する](../test/generations/rollback-partial-and-abort.md)（rollback の途中失敗・
前提不成立・conflict 全件報告）・[実 nix を通した home mode で世代がコミットされ、rollback が前世代の配置へ戻る](../test/generations/home-generation-e2e.md)（e2e での世代往復）が緩和する。
