---
id: "TC-36ea3609-d52e-42d4-975c-40fb89b23919"
type: test_condition
name: "rollback が FS を前世代へ再収束させてから最後にポインタを移す"
status: stable
mitigates:
  - "RISK-fbf029f6-866c-4a08-a4eb-f09e3c7e907e"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:53:33+09:00"
---
# TC-36ea3609-d52e-42d4-975c-40fb89b23919: rollback が FS を前世代へ再収束させてから最後にポインタを移す

## テスト条件

`Rollback` を tmpdir 上の実 FS で駆動し、現世代 N から N-1 への FS 再収束を検証する。
world は「gen1 と gen2 の link farm を書き、`profile-N-link` 世代リンクを張り、profile を
gen2 へ向け、FS を gen2 の配置状態にする」形で組む。

**素の再収束** — gen1 = {a, b}・gen2（current）= {a, c} のとき、c が古くなった配置として除去され、
b が再配置され、a は据え置かれる。ポインタ移動は FS 収束の後に起きる
（先に動かすと基準世代が N-2 へずれる → [rollback は FS を先に収束させてから profile ポインタを最後に移す](../../requirements/rollback-refit-then-pointer.md)）。

**copy 反映を含む再収束** — 再収束は symlink の配置だけでなく copy 反映段
（`materializeCopies` の place-once）を含む。戻り先 N-1 の copy entry は配置先が不在なら
実体化され、既存の copy 先は触られない（`rollback` に `--recopy` 相当は無い →
[rollback は FS を先に収束させてから profile ポインタを最後に移す](../../requirements/rollback-refit-then-pointer.md)）。祖先 migration 越しの copy child と、
symlink → copy への method change の 2 形がこの条件に属する（→ issue #178）。

**祖先 symlink からの復帰** — 世代 N が配下の per-file entries を祖先 symlink 1 本へ
移行していた場合、N-1 へ戻るには plan の PreRemove がその祖先 symlink を先に除去し、
配下の子 entries を実ディレクトリへ置き直す必要がある。PreRemove が無いと、
`ensureParentDir` の `MkdirAll` が symlink 上で no-op になり、子の `os.Symlink` が
symlink 越しに store を指して EEXIST / EROFS で失敗する（→ [配置前除去を planner が計算するデータとして表す](../../adr/preremove-as-planner-data.md)、issue #173）。
収束後に当該パスが実ディレクトリへ戻っていることを確認する。

`rootKind` から絶対 root を得る側は本条件の担当から外れる。root 解決そのものは
[root 解決と profile ディレクトリ準備の成功時・失敗時の処理が仕様どおりであること](../engine-core/root-resolution.md)（engine-core）へ一本化した（→ issue #328）。

上位の規範は [エンジンとコマンド層の Go テストは nix を介さない実 FS 統合テストを主戦力とする](../../test-plan/go-test-layering.md)（`internal/engine/` を実 FS の tmpdir で駆動する統合レベル）。

## 対応する CASE

[generations_test.go — rollback の再収束と失敗時の処理](generations-test-go.md)（`internal/engine/generations_test.go`）。
