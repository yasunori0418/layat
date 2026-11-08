---
id: "ADR-facd3173-a5ce-4d81-90a9-370fb26428ff"
type: adr
name: "apply は世代の commit の前に記録を書く"
status: stable
sources:
  - id: "issue-580"
    resource: "https://github.com/yasunori0418/layat/issues/580"
    title: "#580"
  - id: "issue-406"
    resource: "https://github.com/yasunori0418/layat/issues/406"
    title: "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0058 §4a を書き直した"
justifies:
  - "REQ-9d622539-a1e5-4719-af51-e0cd85b66f35"
references:
  - "ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d"
  - "ADR-d97154cd-a2e5-4853-a682-800ff9435c89"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
  - "ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-05T00:40:25+09:00"
---
# ADR-facd3173-a5ce-4d81-90a9-370fb26428ff: apply は世代の commit の前に記録を書く

- ステータス: stable
- 日付: 2026-11-03
- 関連: ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d, ADR-d97154cd-a2e5-4853-a682-800ff9435c89, ADR-fe035038-4415-4a1e-aa5c-509fdfce9955, ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93, GitHub Issue #580, #406
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0058 §4a を書き直した

## 背景

apply は FS を変えてから世代を commit する。この順序のどこで記録を書くかによって、
途中で失敗したときに記録と FS がどう食い違うかが決まる。

## 決定

apply は、全ての FS 操作が成功した直後、世代の commit より前に、記録を一時ファイルと rename で
atomic に書く。書けなければ apply を失敗とし、undo journal で FS を巻き戻す。世代スキップ
の処理でも同じ時機に記録を更新し、書けなかったときの扱いも同じにする。

## 根拠

記録は FS の事実である。FS が変わった時点で書き、commit の成否とは切り離す。commit
が失敗しても記録は FS と一致しているので、次の apply は記録を比較の基準にして収束できる。

書けなかったときに FS を戻せば、記録と FS がずれた状態を残さない。世代スキップの処理も drift
修復で FS を変える。だから同じ規則に揃える。

## 棄却した代替案

### 世代の commit の後に書く

commit が成功して記録だけが書けないと、記録が前回のまま FS だけが進み、次の apply
が古い比較の基準で古くなった配置を除去する。

### 書き込みの失敗を warning にとどめ、apply は成功させる

記録と FS がずれたまま残り、次の古くなった配置の除去を誤らせる。
