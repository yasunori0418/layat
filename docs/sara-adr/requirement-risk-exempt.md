---
id: "SDR-bb13e2a5-b2ce-4af9-9148-37f4e92ef050"
type: sara_adr
name: "requirement に risk_exempt を設け、理由を書いた requirement を未カバーから外す"
description: "requirement に任意フィールド risk_exempt を設け、requirement 水準の risk を持たない理由を 1 文で書いた requirement を未カバーの一覧から外して免除として列挙する。"
status: stable
sources:
  - id: "issue-648"
    resource: "https://github.com/yasunori0418/layat/issues/648"
    title: "epic: 要求からテストまでの連鎖を sara のモデルでつなぐ — test_plan とテストケース・設計とリスクの切れ目"
origin: "epic #648 の grilling セッション（2026-11-08）で確定"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-09T22:59:08+09:00"
---
# requirement に risk_exempt を設け、理由を書いた requirement を未カバーから外す

- ステータス: stable
- 日付: 2026-11-09
- 関連: [保証する連鎖は requirement → risk → test_condition → test_case とし、design は任意の中継点にする](guaranteed-chain-design-optional.md), GitHub Issue #648, `docs/model.yaml`, `docs/model.md`
- 起点: epic #648 の grilling セッション（2026-11-08）で確定

## 背景

2026-11-06 の監査（sara 0.12.0）で、どの `risk` にも `threatens` されていない `requirement` が 15 件見つかった。このうち 8 件は、フラグの不在、依存や構造の制約、配布形態の宣言で、テストで確かめる振る舞いを持たない。

モデルにはこの区別を書く場所が無い。`sara-gap` の未カバーの一覧では、リスク識別に未着手の `requirement` と、`risk` を持たないと判断した `requirement` を見分けられない。

## 決定

`requirement` に任意の text フィールド `risk_exempt` を設け、`requirement` 水準の `risk` を持たない理由を 1 文で書く。

値を持つ `requirement` は `sara-gap` の未カバーの一覧から外し、免除として別の節に列挙する。

`design` には設けない。[risk を持たない design は正常と定めた](guaranteed-chain-design-optional.md)ので、免除を書く必要が無い。

## 根拠

理由が item に残るので、免除の妥当性をレビューで確かめられる。真偽値では「免除した」という事実しか残らない。

免除の一覧が別の節に出るので、免除した `requirement` が増えても見えなくならない。

任意フィールドの追加なので、既存の item は書き換えなくても `sara check` を通る。

## 棄却した代替案

### すべての `requirement` に `risk` を起票する

確かめる振る舞いを持たない `requirement` に `risk` を起こすと、その `risk` を `mitigates` する `test_condition` と、それを `covers` する `test_case` を書けない。未カバーの一覧の項目が `requirement` から `risk` へ移るだけで、着手できる作業は増えない。
