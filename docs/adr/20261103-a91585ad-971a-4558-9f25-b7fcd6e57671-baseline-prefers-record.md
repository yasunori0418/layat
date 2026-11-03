---
id: "ADR-a91585ad-971a-4558-9f25-b7fcd6e57671"
type: adr
name: "古くなった配置の除去は、比較の基準を記録から優先して読む"
status: 採用
issues:
  - "#580"
  - "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0058 §3a を書き直した"
justifies:
  - "REQ-9d622539-a1e5-4719-af51-e0cd85b66f35"
  - "REQ-1be4d678-959c-44d7-a346-44bfd95af56e"
  - "DSG-3c156bbe-e72b-43dd-8cb0-2cc105314ad5"
supersedes:
  - "ADR-0058"
references:
  - "ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93"
  - "ADR-facd3173-a5ce-4d81-90a9-370fb26428ff"
  - "ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d"
  - "ADR-d97154cd-a2e5-4853-a682-800ff9435c89"
  - "ADR-fe035038-4415-4a1e-aa5c-509fdfce9955"
---
# ADR-a91585ad-971a-4558-9f25-b7fcd6e57671: 古くなった配置の除去は、比較の基準を記録から優先して読む

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93, ADR-facd3173-a5ce-4d81-90a9-370fb26428ff, ADR-55d8ecdf-86bb-45d0-8c60-ac9c630cb62d, ADR-d97154cd-a2e5-4853-a682-800ff9435c89, ADR-fe035038-4415-4a1e-aa5c-509fdfce9955, GitHub Issue #580, #406
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0058 §3a を書き直した

## 背景

古くなった配置の除去は、前回の manifest（比較の基準）と新しい manifest
の差分で行う（ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93）。last-applied 記録を導入しても、
記録を持たない既存の環境はある。記録と profile が両方あって食い違うこともある。記録は世代の commit
より前に書く。そのため commit が失敗すると、記録が新しく、profile が古いまま残る。

## 決定

比較の基準は、記録があれば記録の manifest、無ければ profile の current 世代から読み、
どちらも無ければ空（初回）とする。記録と profile が両方あって `linkFarm` が食い違えば記録を採り、
warning を 1 行出す。どちらも無い初回は、「last-applied 記録なし」の warning を 1 行出す。

この順序を使う処理（apply の古くなった配置の除去と drift 修復・reset の撤去対象・rollback
の比較の基準）と warning の細目は REQ-9d622539-a1e5-4719-af51-e0cd85b66f35 と
DSG-3c156bbe-e72b-43dd-8cb0-2cc105314ad5 に、世代機構の側の読み方は
REQ-1be4d678-959c-44d7-a346-44bfd95af56e に置く。

## 根拠

記録は FS に最後に置いたものを表すので、profile より FS の事実に近い。commit
が失敗して両者がずれたときに記録を採れば、コミットされなかった試行が置いた entry
も古くなった配置の除去の対象に入る。

profile へ後退するのは、記録を持たない環境を止めないための互換で、記録が揃えば使われなくなる。
warning は沈黙規律の対象外で stderr に出るので、記録と profile の食い違いや、記録がまだ無いことに利用者が気付ける。

## 棄却した代替案

### profile を優先し、記録は profile が無いときだけ読む

commit の失敗後に、コミットされなかった試行が置いた entry を撤去し損ねる。profile が layat
の外へ出る終着点でも成り立たない。

### 記録が無ければ error にする

記録を持たない既存の環境で、更新した直後の apply が全部止まる。
