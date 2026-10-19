---
id: "TC-3f8d24a7-4e43-4998-b47f-a9e03cbcbfee"
type: test_condition
name: "実 nix を通した system mode で世代がコミット・往復し、非 root の変更系は FS に触れずに止まる"
mitigates:
  - "RISK-1a8f91c8-dfaa-4598-805a-a3b71cb967f0"
  - "RISK-fbf029f6-866c-4a08-a4eb-f09e3c7e907e"
  - "RISK-cdcc6faf-9164-409b-b584-2921fa036d10"
---
# TC-3f8d24a7-4e43-4998-b47f-a9e03cbcbfee: 実 nix を通した system mode で世代がコミット・往復し、非 root の変更系は FS に触れずに止まる

## テスト条件

Go テストが double へ差し替えている層（euid・`nix build` / `nix-env --set`・世代一覧）を
実物に戻し、system mode の世代機構と権限ゲートが一気通貫で動くことを検証する。root は
user namespace で取り、system の profile 基底と記録基底は隔離先へ向ける。

- **配置と世代コミット** — root の apply が `--root` の配下へ配置し、profile 基底の
  `<roothash>/<name>` に profile・世代リンクと backref `.root`、記録基底の
  `layat/<roothash>/<name>/last-applied.json` が立つ
- **世代の往復** — entry を入れ替えた再 apply で世代 2 になり、`rollback` が世代 1 の
  配置へ戻し、`reset --yes` が配置を撤去する
- **列挙** — `list-generations <name>` が世代を `info.mode == "system"` で返し、
  `--all` は `<roothash>` 系列を列挙に含めない
- **権限ゲート** — 非 root の `apply` / `rollback` / `reset --yes` が exit 1 で止まり、
  当該 subject の `errors[]` が `E_PERMISSION` を運び、配置先と profile を変えない。
  読み取り系の `apply --dryrun` と `list-generations` は非 root でも通る

## 対応する CASE

CASE-99803cb0-6fd0-4004-af78-83e301320ce0（`tests/e2e/scenarios/12-system.sh`）。E2E ハーネス全体のスコープは
TP-229b69c0-cf5e-4fb6-a353-27e5064d93e9 が定める。
