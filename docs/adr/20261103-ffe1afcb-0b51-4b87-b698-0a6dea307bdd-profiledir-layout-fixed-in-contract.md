---
id: "ADR-ffe1afcb-0b51-4b87-b698-0a6dea307bdd"
type: adr
name: "profileDir の内部レイアウトを世代バックエンド契約の固定部分にする"
status: 採用
issues:
  - "#406"
origin: "Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §3a を書き直した"
justifies:
  - "REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b"
  - "REQ-2aa3abbc-90b2-486e-92de-d785554bdeb3"
supersedes:
  - "ADR-0057"
references:
  - "ADR-419914ba-9896-4dd1-853b-fb514fcb9f38"
---
# ADR-ffe1afcb-0b51-4b87-b698-0a6dea307bdd: profileDir の内部レイアウトを世代バックエンド契約の固定部分にする

- ステータス: 採用
- 日付: 2026-11-03
- 関連: ADR-419914ba-9896-4dd1-853b-fb514fcb9f38
- 起点: Issue #406（epic: 世代バックエンドのインターフェイス化）の grilling（2026-10-11）。ADR-0057 §3a を書き直した

## 背景

世代操作を世代バックエンドの契約へ切り出すと、profileDir の中身（profile リンク・世代のリンク・
`.pending`）を誰が所有するかが問題になる。engine は前世代の manifest や drift の判定のために、
このレイアウトを直接読んでいる。

## 決定

profileDir の内部レイアウトは世代バックエンドの契約の固定部分とし、実装ごとに変えない。レイアウトの
定義は `internal/paths` に置き、engine とバックエンドの双方がそれに従う。

並びそのものは ADR-419914ba-9896-4dd1-853b-fb514fcb9f38 が決めており、細目は
REQ-2aa3abbc-90b2-486e-92de-d785554bdeb3 に置く。レイアウトを契約の固定部分とする規範は
REQ-833205ca-adfc-4c86-bf66-ae4d8fe4885b に置く。

## 根拠

レイアウトが固定なら、engine の読み取りは契約を経ずに動き、FS ベースの Fake もそのまま使える。
実装ごとにレイアウトを変えられるようにすると、engine の読み取りを全て契約の操作に置き換えなければならない。

## 棄却した代替案

### レイアウト全体をバックエンドの所有にする

engine の読み取りを契約に載せるか、Fake が実装ごとのレイアウトを模倣するかの二択になる。2 つ目の実装が
無い段階では、どちらも検証できない。
