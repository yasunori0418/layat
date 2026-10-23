---
id: "TC-1bc3e39a-116a-4f26-96b4-1d15e6f837bd"
type: test_condition
name: "status の CLI がフィルタ・ストリームの排他・--json の形・error 時の続行の契約を守る"
mitigates:
  - "RISK-e92487a5-fff9-40db-844d-6590f1f36dec"
---
# TC-1bc3e39a-116a-4f26-96b4-1d15e6f837bd: status の CLI がフィルタ・ストリームの排他・--json の形・error 時の続行の契約を守る

## テスト条件

`layat status [name]` の CLI 層が、engine の結果を契約どおりに絞り込み、名付け、出力することを
検証する。

**フィルタ**

- `[name]` は `<name>` 部分の完全一致で、基底・mode・root を問わずに絞る（前方一致や
  subject 名全体との一致にならない）
- `<name>` を列挙できない系列の error subject は、`[name]` を指定しても落とさない
- 一致が 0 件なら stdout に一覧を出さず、stderr に 1 行出して exit 0（`--json` では
  `results: []` のエンベロープを出し、exit 0）

**system 基底の受け渡し**

- `LAYAT_SYSTEM_STATE_DIR` があればその値を engine の `SystemStateDir` へ渡し、無ければ空で
  渡す（engine が既定の `/var/lib` を使う）

**subject 名とテキスト出力**

- 直キーは `<mode>:<name>`、`<roothash>` 系列は `<mode>:<name>@<root>`
- backref が読めない系列は `<mode>:<name>@<roothash>`、`<name>` を列挙できない系列は
  `<mode>:@<roothash>`。どちらも mode は基底に応じた `home` / `system`
- subject ごとに `# <subject> (applied <appliedAt>)` と `<target>\t<method>\t<dest>` の行を出し、
  subject の間を空行で区切る

**ストリームの排他と `--json` の形**

- `--json` のときは行指向の一覧を stdout に出さず、エンベロープだけを出す
- `results[i].result.info` が `{mode, root, appliedAt, linkFarm, entries}` で、entries は 0 件でも
  配列、`items: []`、`generation` スロットを出さない。エンベロープは outturn の conformance を通る

**error 時の続行**

- 不正な記録を持つ subject は `errors[]` に error を持ち、他の subject は通常どおり並ぶ。
  エンベロープの status は error、exit 1
- `--json` なしでは、不正な記録を持つ subject の error を stderr に出し、他の subject は
  通常どおり stdout に出して exit 1

上位の規範は TP-d3000054-42d9-4bac-912a-dd3abc38d3e9（エンベロープ適合と payload 意味論）と
TP-e7c25263-6d2d-4a37-8275-26906889d912（CLI 層の判断）。
