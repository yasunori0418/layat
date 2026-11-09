---
id: "CASE-b2a424d5-991d-4534-b52f-62c9704a0b8b"
type: test_case
name: "e2e 13-status — 実 apply が書いた記録を status が列挙し、[name] で絞り、system 基底の記録も拾う"
status: stable
resource: "tests/e2e/scenarios/13-status.sh"
covers:
  - "TC-076bd7b4-6427-4a35-9bab-86daaaada504"
  - "TC-1bc3e39a-116a-4f26-96b4-1d15e6f837bd"
follows:
  - "TP-229b69c0-cf5e-4fb6-a353-27e5064d93e9"
  - "TP-d3000054-42d9-4bac-912a-dd3abc38d3e9"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T09:27:49+09:00"
---
# e2e 13-status — 実 apply が書いた記録を status が列挙し、[name] で絞り、system 基底の記録も拾う

## 対象

`tests/e2e/scenarios/13-status.sh`

`e2e_isolate` で仮 `$HOME`・`$XDG_STATE_HOME` と system の記録基底（`LAYAT_SYSTEM_STATE_DIR`）を
隔離し、homeRoot の `home` と projectRoot の `proj` を持つ fixture flake を実 nix 経由で apply する
E2E シナリオ。

走査と読み取りの分岐は [status_test.go — 記録基底の列挙順・基底の扱い・読めない系列・読み取り専用](engine-status-test-go.md)（tmpdir に組んだ記録）、
出力契約は [status_test.go — status の subject 名・フィルタ・一致 0 件・ストリームの排他・--json の形・error 時の続行](../cli-json/status-cli-test-go.md)（engine の差し替え）が覆う。**このシナリオの
持ち分は「実 apply が書いた本物の記録を status が読み取り、CLI の実際の処理で出せるか」**である。
記録の書式や置き場が apply と status の間でずれる退行は、手組みの記録だけでは気付けない。

## 検証内容

- **両方の subject の列挙** — `layat status` が exit 0 で、見出しが `# home:home`、
  `# project:proj@<root>`（root は git toplevel の実パス）の順に並ぶ。home の entry 行が
  `.cfg/a\tsymlink\t/nix/store/...` の形で出る
- **`[name]` の絞り込み** — `layat status proj` が `project:proj@<root>` だけを出す
- **一致 0 件** — `layat status nope` が exit 0・stdout 空で、stderr に通知を出す
- **`--json` の info** — subject 名の並び、home の `mode: home`・`root: ""`・entries の target
  `.cfg/a`、proj の `mode: project`・`root` が git toplevel・entries の target
  `.layat-out/proj` を `assert_json` で確かめる。`items: []`・`generation` スロット無し・
  `status: success`・`dryRun: false` で、エンベロープは `outturn-validate` を通る
- **system 基底** — 隔離した system の記録基底の `<name>` 直キーへ home の記録を複製して置くと、
  user 基底の 2 件の後に `system:sys`（`info.mode: system`）として並び、`layat status sys` で
  それだけに絞れる。system 基底へ apply するには root 権限が要るので、記録は apply ではなく
  複製で用意する
