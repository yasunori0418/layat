---
id: "ADR-63ec1446-d328-42d2-b8d2-e8a11d752ed7"
type: adr
name: "prune は新設時から --json に対応する"
section: cli-output
status: stable
sources:
  - id: "issue-580"
    resource: "https://github.com/yasunori0418/layat/issues/580"
    title: "#580"
origin: "次期マイルストーン計画の grilling（2026-07-04）。公開後の実運用で孤児系列が溜まり始めたことを消費側の要求と判断した。ADR-0034 §2c を書き直した"
justifies:
  - "REQ-a47fbfb0-c808-42e9-b2ad-767c933494d6"
references:
  - "ADR-ad1c2507-1f42-4cd3-ab56-9183892f8988"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-03T22:55:03+09:00"
---
# prune は新設時から --json に対応する

- ステータス: stable
- 日付: 2026-11-03
- 関連: [profile は解決済み root の roothash でキーする](profile-keyed-by-roothash-of-resolved-root.md), GitHub Issue #580
- 起点: 次期マイルストーン計画の grilling（2026-07-04）。公開後の実運用で孤児系列が溜まり始めたことを消費側の要求と判断した。ADR-0034 §2c を書き直した

## 背景

layat は `--json` を全コマンドに一律で持たせている。prune を新設するとき、`--json`
を最初から持たせるか、後から足すかを決める必要がある。

## 決定

prune は新設時から `--json` に対応し、削除した系列（`--dryrun` では削除予定の系列）と、
残した系列を機械可読な形で出す。エンベロープへの載せ方と系列の形は
[prune --json はトップレベル info の removed に削除系列を、skipped に判定か削除に至れず残した系列を載せる](../requirements/prune-json-payload.md) に置く。

## 根拠

全コマンドが `--json` を持つという一律の方針から、prune だけを外す理由が無い。後から足すと、
テキスト出力を解析するスクリプトがその間に生まれ、移行の手間が残る。

## 棄却した代替案

### テキスト出力だけで新設し、要求が出てから `--json` を足す

prune だけが一律の方針の例外になり、その間に書かれた自動化はテキストの解析に頼ることになる。
