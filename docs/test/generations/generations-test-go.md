---
id: "CASE-364ebb9d-b9f8-4c79-8509-8bbf31e73698"
type: test_case
name: "generations_test.go — rollback の再収束と失敗時の処理"
status: stable
resource: "internal/engine/generations_test.go"
covers:
  - "TC-36ea3609-d52e-42d4-975c-40fb89b23919"
  - "TC-fa7911c6-7347-47ff-b121-bec53562c063"
  - "TC-be3d3dc9-1ead-4e2e-8eea-46e1c8884562"
follows:
  - "TP-e7c25263-6d2d-4a37-8275-26906889d912"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:53:14+09:00"
---
# generations_test.go — rollback の再収束と失敗時の処理

## 対象

`internal/engine/generations_test.go`

`Rollback` は tmpdir 上の実 FS へ世代リンクと profile を手で組み立ててから駆動する
統合テスト。世代バックエンドは `generation.Fake`（FS ベース。世代一覧は sibling の
`profile-<N>-link` から読み、Switch は profile リンクを張り直す）を注入し、Switch の
失敗は `OnSwitch` フックで起こす。nix-env の世代一覧パースは
[nixprofile_test.go — 世代リンクの走査と 2 基底](nixprofile-test-go.md)（nixprofile）が持つ。

`resolveRoot` を直接呼ぶテストは分岐網羅を 1 ファイルで追えるよう engine_test.go へ集約した
（集約先の集合は多くが純引数だが、cwd を入力に取る処理も含む。
→ [internal/engine/engine_test.go — 実 tmpdir に対する apply の統合テスト](../engine-core/engine-test-go.md) / [root 解決と profile ディレクトリ準備の成功時・失敗時の処理が仕様どおりであること](../engine-core/root-resolution.md)、issue #328）。

## 検証内容

**rollback の再収束**（[rollback が FS を前世代へ再収束させてから最後にポインタを移す](rollback-reconvergence.md)）

- gen1 = {a, b}・gen2（current）= {a, c} からの再収束。c の古くなった配置の除去・b の再配置・
  a の据え置き。`RollbackResult` の世代観測が From / To を映し、Entries が戻り先世代の全
  インベントリを持つ（→ issue #130）
- 祖先 symlink 世代（N）から per-file 世代（N-1）への復帰。plan の PreRemove が祖先
  symlink を除去してから子を実ディレクトリへ置き直す（→ [配置前除去を planner が計算するデータとして表す](../../adr/preremove-as-planner-data.md)、issue #173）
- rollback は copy entries も実体化する。祖先 symlink 世代から戻るときの copy child と、
  gen N が symlink へ変えた target を N-1 の copy へ戻す method change の両方で、apply と同じ
  materializeCopies 段が place-once copy を配置する（→ issue #178）

**rollback の失敗時の処理**（[rollback は途中失敗を巻き戻し、ポインタ移動の失敗は巻き戻さず、前提不成立では停止する](rollback-partial-and-abort.md)）

- `Switch` の失敗は unwind しない。部分の `RollbackResult` が From == To == 現世代を
  表し、既に行われた配置・古くなった配置の除去はそのまま残る（→ [undo の発火を FS 反映の段に限り、世代コミットの後は対象にしない](../../adr/undo-scope-ends-at-commit.md)）
- 祖先 migration 成功後の無関係な配置失敗で、migration ごと巻き戻る（→ [配置前除去の除去も undo の対象にする](../../adr/undo-covers-preremove.md)、
  issue #168）
- copy 配置の失敗で、直前に置いた symlink ごと unwind してポインタ移動へ進まない
  （→ issue #178）
- 祖先の PreRemove が drift で失敗したときはポインタ移動へ進まず、配置を触らない
- 前世代が無い（最古世代）・profile が無い（未 apply）はエラーで停止する
- `Backend` が nil のとき profileDir の解決前にエラーで停止する
- 戻り先の target が複数 foreign 実体で塞がれているとき、全 conflict を stderr へ
  列挙してから件数付きの集約エラーを返す（→ issue #176）

**system mode の euid の条件**（[system mode の root 解決・基底切替・euid の条件が仕様どおりであること](../engine-core/system-mode-root-gate.md)）

- euid 0 以外を注入した system mode の rollback が `fs.ErrPermission` を wrap したエラーで止まり、
  root・profile 基底・記録基底に何も作らない
- euid 0 を注入した system mode の rollback は `SystemDir` の世代を辿り、記録を `SystemStateDir`
  配下で戻り先世代の link-farm に書き直し、ユーザー state と `Backend.SystemBase` に何も作らない
