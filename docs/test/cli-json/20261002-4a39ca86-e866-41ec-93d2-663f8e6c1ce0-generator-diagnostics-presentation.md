---
id: "TC-4a39ca86-e866-41ec-93d2-663f8e6c1ce0"
type: test_condition
name: "生成器の診断が成功・失敗とも writer へ届き、失敗は分類と案内付きで出力面ごとに提示される"
mitigates:
  - "RISK-cdca94ab-ed44-4d60-a04d-4dc75d61f731"
---
# TC-4a39ca86-e866-41ec-93d2-663f8e6c1ce0: 生成器の診断が成功・失敗とも writer へ届き、失敗は分類と案内付きで出力面ごとに提示される

## 条件

nix 生成器は PATH の先頭に置いた nix のスタブで駆動し、CLI の提示は `generator.Error` を
直接与えて確かめる。実 nix での到達は e2e で確かめる。

- **素通し** — eval / build の両経路で、nix の stderr が成功時も失敗時も生成器の writer へ
  届く。writer へは行単位で渡り、行の途中で切れた書き込みが来ても行が揃うまで渡さない
- **分類と案内** — experimental-features 未有効の 3 つの文言が `PrerequisiteMissing` になり、
  `Guidance` に nix.conf と `NIX_CONFIG` の両手段が入る。属性の不在は `NotFound` で、単発は
  config 名の確認を、一括は config の定義を案内する。それ以外は `Failed`。`Message` は失敗した
  サブコマンドまでの 1 行で、生の診断は `Stderr` に残る
- **出力面ごとの提示** — 人間向けは要約 1 行と `Guidance` で生の診断を再掲しない。`--debug` では
  要約行に生成器名を後置する。`--json` の `errors[].message` は要約とキャプチャを含み、
  `--debug` に左右されず生成器名を前置きしない
