---
id: "TC-527b5034-715a-4df1-871a-072dc9062704"
type: test_condition
name: "実 nix を通した home mode で世代がコミットされ、rollback が前世代の配置へ戻る"
status: stable
mitigates:
  - "RISK-fbf029f6-866c-4a08-a4eb-f09e3c7e907e"
  - "RISK-cdcc6faf-9164-409b-b584-2921fa036d10"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:49+09:00"
---
# 実 nix を通した home mode で世代がコミットされ、rollback が前世代の配置へ戻る

## テスト条件

Go テストが double へ差し替えている層（`nix build` / `nix-env --set` / 世代一覧）を
実物に戻し、home mode の世代機構が一気通貫で動くことを検証する。仮 `$HOME` と
`$XDG_STATE_HOME` で隔離した環境に fixture flake を書き、entry を入れ替えながら
apply を重ねる。

- **配置と世代コミット** — apply が `$HOME` 配下へ配置し、`<state>/nix/profiles/layat/
  home/profile` に profile リンクが作られ、`list-generations` が世代を返す
- **世代の積み上がり** — entry を入れ替えて再 apply すると、旧 entry が古くなった配置として除去され、
  新 entry が配置され、世代が 2 つ以上になる
- **rollback の往復** — `rollback` が前世代の配置状態（旧 entry あり・新 entry なし）へ
  戻す
- **成功時沈黙** — 指定が無いときは配置レポートを出さず、`-v` で出す（→ [指定が無いときは成功時に何も出力しない](../../adr/silent-on-success.md) / [-v で配置レポートを opt-in で出す](../../adr/verbose-opts-into-placement-report.md)）
- **撤去の非遷移** — `reset` は配置を撤去するが profile リンクと世代の本数は動かさない
  （FS のみの撤去 → [reset は profile を触らない FS-only teardown で配置物を無い状態へ戻す](../../requirements/reset-fs-only-teardown.md)）
- **`--json` の世代観測** — `apply --dryrun --json` の generation が before == after
  （非 null）を含み、`list-generations --json` が `{number, date, current}` の 2 世代を
  返し、current が rollback 先の 1 世代だけであること。`items` は空で generation スロットを
  持たず `dryRun` が false であること（→ issue #132）。`rollback --json` は generation が
  後退する向きの遷移を含み、`reset --json` は generation スロット自体を持たないこと
  （→ issue #285）
- **`--all` の読み取り専用** — 全 config を対象とする列挙が観測に留まり、profile の
  リンク先・世代の本数・`$HOME` 配下の配置のいずれも動かさないこと（→ [rollback と list-generations は home / system mode 限定にする](../../requirements/rollback-list-generations-home-only.md) の
  `SHALL be read-only`）。JSON の処理とテキストの処理は印字の分岐が別なので、両方の処理を
  またいだ前後比較で見る。観測するのは列挙に現れる home 側だけで、除外された
  project mode の profile が動かないことは本条件の範囲外（未被覆）

## 対応する CASE

[e2e 02-home — 実 nix での home mode 世代コミットと rollback / reset の往復](e2e-02-home.md)（`tests/e2e/scenarios/02-home.sh`）。E2E ハーネス全体のスコープは
[非 NixOS で動く主張を実 nix の一気通貫 E2E で検証する](../../test-plan/e2e-harness-scope.md)、`--json` の検証範囲は [エンベロープの outturn 適合を Go テストと E2E の両方で検証する](../../test-plan/json-conformance-verification.md) が定める。
