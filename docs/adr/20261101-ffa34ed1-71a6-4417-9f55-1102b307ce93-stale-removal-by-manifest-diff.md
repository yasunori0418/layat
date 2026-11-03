---
id: "ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93"
type: adr
name: "stale 除去は manifest の差分で行う"
status: 採用
origin: "ADR-0002 §決定（stale 除去は世代由来の store マニフェストを diff して行う）を書き直した"
justifies:
  - "REQ-1be4d678-959c-44d7-a346-44bfd95af56e"
  - "REQ-16aef46b-7bb8-4ca1-b962-e9f3ed1fd1d2"
supersedes:
  - "ADR-0002"
---
# ADR-ffa34ed1-71a6-4417-9f55-1102b307ce93: stale 除去は manifest の差分で行う

- ステータス: 採用
- 日付: 2026-11-01
- 起点: ADR-0002 §決定（stale 除去は世代由来の store マニフェストを diff して行う）を書き直した

## 背景

config から entry が消えたら、前回置いた symlink を除去しなければならない。そのためには
「前回何を置いたか」を何かに記録しておき、今回の内容と比べる必要がある。

## 決定

stale 除去は、前回の適用の manifest（baseline）と新しい manifest の差分から、消えた entry を
判定して行う。「配置したもの」の manifest は link farm derivation の一部として store に埋め込む。
baseline は last-applied 記録があれば記録の manifest、無ければ自 profile の current 世代の
store manifest とし、host の世代は読まない。

manifest の置き場所と baseline の選び方は REQ-1be4d678-959c-44d7-a346-44bfd95af56e、baseline を
使った除去の条件は REQ-16aef46b-7bb8-4ca1-b962-e9f3ed1fd1d2 に置く。

## 根拠

store に埋め込んだ manifest は変わらず、世代が GC root なので世代とともに残る。last-applied 記録が
無いときも、前回の内容を自 profile の世代から読める。記録を優先する理由と優先順は ADR-a91585ad-971a-4558-9f25-b7fcd6e57671 が
決めている。home-manager も前世代と新世代の manifest を比べて除去しており、実績がある。

## 棄却した代替案

### store に manifest を持たず、store 外の可変 JSON だけで配置を記録する

世代に manifest が無いので、rollback で戻る世代の内容を世代から読めない。last-applied 記録は
最後に置いたものしか持たず、rollback の戻り先は世代から読むので、世代ごとの store manifest を
置き換えない（ADR-be5865a8-000a-491f-8438-6308076ea727）。

### 新しい manifest に無い target を全て除去する

前回の記録と比べないので、layat が置いていないファイルまで消しうる。
