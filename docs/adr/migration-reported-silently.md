---
id: "ADR-8357b519-cb23-4984-ba56-254046779f47"
type: adr
name: "配置前除去による移行を warning にせず -v のときだけ報告する"
section: removal-migration
status: stable
sources:
  - id: "issue-170"
    resource: "https://github.com/yasunori0418/layat/issues/170"
    title: "#170"
origin: "dotfiles 側での実遭遇（#170）と、次期マイルストーンの grilling セッション（2026-07-11）で設計確定。ADR-0046 §4 を書き直した"
justifies:
  - "REQ-c9ab91c1-f778-4f87-a2ea-c66d6b3c2575"
  - "REQ-8ef34101-8150-4124-92d5-94fabe6b5d90"
references:
  - "ADR-36bfbfbd-f1dd-4581-a75a-943d80f6ccf5"
  - "ADR-3dc21349-5504-48a4-ab88-812b92c3bdf9"
  - "ADR-a19bf484-368d-4d01-b539-52de32510fa5"
  - "ADR-3b8175b9-5ebf-4123-8855-f698dc843d6b"
  - "ADR-dfe6654c-8fcd-4cf6-a053-c21ad6dc373c"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-05T00:35:05+09:00"
---
# 配置前除去による移行を warning にせず -v のときだけ報告する

- ステータス: stable
- 日付: 2026-11-03
- 関連: [指定が無いときは成功時に何も出力しない](silent-on-success.md), [-v で配置レポートを opt-in で出す](verbose-opts-into-placement-report.md), [内部コマンドの開示を --debug に分ける](debug-discloses-internal-commands.md), [配置前除去を planner が計算するデータとして表す](preremove-as-planner-data.md), [祖先 symlink は自己記録の古くなったものだけを移行する](ancestor-symlink-recorded-stale-only.md), GitHub Issue #170
- 起点: dotfiles 側での実遭遇（#170）と、次期マイルストーンの grilling セッション（2026-07-11）で設計確定。ADR-0046 §4 を書き直した

## 背景

自己記録の古くなったものを配置前に除去して移行する（[配置前除去を planner が計算するデータとして表す](preremove-as-planner-data.md)）とき、
除去したことを利用者へどう伝えるかを決める必要があった。layat は成功したときに何も出さず（[指定が無いときは成功時に何も出力しない](silent-on-success.md)）、
配置のレポートは `-v` を付けたときだけ出す（[-v で配置レポートを opt-in で出す](verbose-opts-into-placement-report.md)）。

## 決定

配置前除去で除去したものは warning にせず、通常の配置レポートと同じく、
`-v` を付けたときだけ報告する。`--dryrun` では、自己記録の古くなったものの除去を conflict ではない除去として
示し、foreign な祖先は conflict として示す。

報告の規範は [祖先 symlink は自己記録の古くなったものだけを配置前除去し、それ以外はエラーで停止する](../requirements/ancestor-symlink-nest-migration.md) と
[成功時はデフォルト沈黙とし warning と error は常時 stderr へ出す](../requirements/silent-on-success.md) に置く。

## 根拠

自己記録の古くなったものの除去は、意図された安全な操作である。自分で置いたものを自分で消している
だけで、利用者が対処すべきことは無い。成功時の沈黙という出力の規律にそのまま従う。

意図された操作を warning にすると、本当に注意すべき warning（foreign な symlink の上書きなど）が
埋もれる。

## 棄却した代替案

### 移行のたびに warning を出す

利用者に対処の要らない出力が増え、本当に注意すべき warning が埋もれる。
