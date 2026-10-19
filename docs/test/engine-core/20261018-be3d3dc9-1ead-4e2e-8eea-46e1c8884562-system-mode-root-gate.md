---
id: "TC-be3d3dc9-1ead-4e2e-8eea-46e1c8884562"
type: test_condition
name: "system mode の root 解決・基底切替・euid ゲートが仕様どおりであること"
mitigates:
  - "RISK-1a8f91c8-dfaa-4598-805a-a3b71cb967f0"
---
# TC-be3d3dc9-1ead-4e2e-8eea-46e1c8884562: system mode の root 解決・基底切替・euid ゲート

`rootKind=system` が root `/` に解決され、`--root` 上書きが優先されることを検証する。

euid 0 以外を注入した system mode の `apply` / `reset` / `rollback` が、`fs.ErrPermission` を
wrap した root 権限のエラーで止まり、root・profile 基底・記録基底のどこにも何も作らないことを
見る。同じ注入の下で `apply --dryrun` と `reset --dryrun` が通ることも同じ条件の下に置く。

euid 0 を注入し `--root` を tmp に向けた system apply が、system 基底の下の
`<roothash>/<name>` に profile と backref を作ること、`SystemDir` が `Backend.SystemBase` より
優先されること、記録を `SystemStateDir` 配下に同じ鍵で書くこと、ユーザー state に何も
作らないことを見る。

`paths.Resolve` / `ResolveRecord` の配置は、home / project / fixed / system と `--root` の
有無の全組み合わせを、リテラルのパスで固定する。基底の切替が他のモードへ波及しないことは
この表が担う。

`list-generations` が検査を受けないことは CLI の担当で、条件の射程から外れる。
root = `/` を実際に使う配置は単体テストでは扱わない。
