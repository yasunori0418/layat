---
id: "ADR-08a6e1be-e424-4cfe-a529-e123a50cd6ca"
type: adr
name: "CLI が叩く attr path を legacy entrypoint の書き方で分岐させない"
status: 採用
issues:
  - "#112"
  - "#107"
origin: "ADR-0032 §2 を書き直した"
justifies:
  - "REQ-c890ce4a-6528-4ab3-ac86-23d7aebff7da"
  - "REQ-496b1a07-5b74-416b-9e5f-3952b4c03737"
  - "DSG-92f54490-872a-42ac-bbd7-d06e9ee381c6"
supersedes:
  - "ADR-0032"
---
# ADR-08a6e1be-e424-4cfe-a529-e123a50cd6ca: CLI が叩く attr path を legacy entrypoint の書き方で分岐させない

- ステータス: 採用
- 日付: 2026-11-03
- 関連: GitHub Issue #112, #107
- 起点: ADR-0032 §2 を書き直した

## 背景

legacy entrypoint には、`mkShell` の `passthru` に named manifest を載せる正準形と、トップレベルの
attrset を返す代替形の 2 つの書き方がある。CLI がこの 2 つを見分けて attr path を変えるかどうかを
決める必要があった。

## 決定

CLI は entrypoint の種別（flake / legacy）でだけ attr path を変え、legacy の 2 つの書き方は区別
しない。どちらの書き方でも、CLI が評価する attr path は `layat.<name>` になる。

種別ごとの attr path は REQ-c890ce4a-6528-4ab3-ac86-23d7aebff7da と
REQ-496b1a07-5b74-416b-9e5f-3952b4c03737 に、分岐を attr path の組み立てに閉じる設計は
DSG-92f54490-872a-42ac-bbd7-d06e9ee381c6 に置く。

## 根拠

derivation は attrset で、`passthru` の各フィールドはその attrset にマージされる。Nix の属性選択は
`passthru` にもトップレベルの属性と同じように届くので、2 つの書き方で同じ attr path が通る。

CLI に書き方を判定する処理を持ち込む必要が無く、利用者がどちらで書いても同じコードパスを通る。

## 棄却した代替案

### 書き方を判定して attr path を切り替える

属性選択が `passthru` にそのまま届く以上、判定は何も得ない。判定を誤るという新しい失敗の
原因と、それを確かめるテストが増えるだけになる。
