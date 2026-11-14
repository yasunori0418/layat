---
id: "ADR-a6a1ef1b-676c-484a-9c8a-3e6d49c1b444"
type: adr
name: "世代スキップには profile に加え記録の一致も要する"
section: generations
status: stable
sources:
  - id: "issue-580"
    resource: "https://github.com/yasunori0418/layat/issues/580"
    title: "#580"
  - id: "issue-406"
    resource: "https://github.com/yasunori0418/layat/issues/406"
    title: "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0058 §3b を書き直した"
justifies:
  - "REQ-46fccb80-4bae-4d37-bc19-dded88e9a9c0"
  - "REQ-9d622539-a1e5-4719-af51-e0cd85b66f35"
references:
  - "ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d"
  - "ADR-facd3173-a5ce-4d81-90a9-370fb26428ff"
  - "ADR-d97154cd-a2e5-4853-a682-800ff9435c89"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
  - "ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-05T00:40:25+09:00"
---
# 世代スキップには profile に加え記録の一致も要する

- ステータス: stable
- 日付: 2026-11-03
- 関連: [project mode では同じ link farm なら世代を積まず drift だけを直す](project-mode-generation-skip-and-drift-repair.md), [apply は世代の commit の前に記録を書く](apply-writes-record-before-commit.md), [rollback は FS を収束させてからポインタを動かす](rollback-baseline-and-pointer-last.md), [1 config = 1 profile を全モードで自前に持つ](one-config-one-profile-all-modes.md), [古くなった配置の除去は manifest の差分で行う](stale-removal-by-manifest-diff.md), GitHub Issue #580, #406
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0058 §3b を書き直した

## 背景

新しい link-farm が前回と同じなら、apply は世代を積まずに drift
だけを直す（[project mode では同じ link farm なら世代を積まず drift だけを直す](project-mode-generation-skip-and-drift-repair.md)）。「前回と同じ」を何と比べるかで、commit
が失敗した後に収束するかどうかが変わる。記録は commit より前に書く。そのため commit が失敗すると、
記録は新しく、profile は古いまま残る。

## 決定

世代スキップは、新しい link-farm が profile の current 世代に一致し、記録があればその
`linkFarm` にも一致するときだけ行う。profile が無ければ、記録だけがあっても skip しない。

## 根拠

記録だけと比べると、同じ設定での再実行が skip に入り、失敗した commit をやり直さない。profile
だけと比べると、古い設定での再実行が skip に入り、コミットされなかった試行が置いた entry
を撤去し損ねる。

両方の一致を条件にすれば、
食い違っている間は通常の処理（記録を比較の基準にした古くなった配置の除去から commit まで）
を通るので、どちらの場合も収束する。

## 棄却した代替案

### 記録の `linkFarm` だけと比べる

commit が失敗した後、同じ設定での再実行が commit をやり直さない。

### profile の current だけと比べる

commit が失敗した後、古い設定での再実行が、コミットされなかった試行の entry を撤去しない。
