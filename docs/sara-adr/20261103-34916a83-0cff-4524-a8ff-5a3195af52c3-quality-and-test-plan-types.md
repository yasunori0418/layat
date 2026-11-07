---
id: "SDR-34916a83-0cff-4524-a8ff-5a3195af52c3"
type: sara_adr
name: "テスト計画の型を品質方針の型から分けて設ける"
status: stable
sources:
  - id: "issue-583"
    resource: "https://github.com/yasunori0418/layat/issues/583"
    title: "#583"
  - id: "issue-237"
    resource: "https://github.com/yasunori0418/layat/issues/237"
    title: "#237"
  - id: "issue-203"
    resource: "https://github.com/yasunori0418/layat/issues/203"
    title: "#203"
origin: "epic #203 の grilling セッション（2026-08-08）で確定。ADR-0049 §2a・§2c を書き直した"
justifies:
  - "INF-659b139d-0cf8-4c65-b30d-93c5ee2dfc71"
references:
  - "SDR-2457aa00-fab7-4c99-af53-2e6643c21c79"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-04T22:44:41+09:00"
---
# SDR-34916a83-0cff-4524-a8ff-5a3195af52c3: テスト計画の型を品質方針の型から分けて設ける

- ステータス: stable
- 日付: 2026-11-03
- 関連: SDR-2457aa00-fab7-4c99-af53-2e6643c21c79, GitHub Issue #583, #237, #203, `docs/model.yaml`
- 起点: epic #203 の grilling セッション（2026-08-08）で確定。ADR-0049 §2a・§2c を書き直した

## 背景

テスト計画や品質方針にあたる文書は、受け皿の型が無いため `requirement` として置かれていた。
`requirement` の親は use_case だが、これらの文書は「ユーザーがどう使うか」から導かれないので親を持てず、
接続漏れのまま残っていた。

テスト系の型（risk / test_condition / test_case）は ISTQB のテスト分析以降に対応する。テスト計画に
対応する型は空いていた。

## 決定

テスト計画の型 test_plan（`docs/test-plan/`）と、品質方針の型 quality（`docs/quality/`）を別々に設ける。
test_plan はテスト計画活動の成果物（スコープ・テストレベル・アプローチ・テスト容易性の要求）を、quality は
品質方針・規約・プロセス横断のガバナンスを受ける。

どちらも solution を親に `derives_from` で張り、`requirement` と同じ `specification` /
`specification_ja` を持つ。`specification` は英語の規範文で書く。sara の RFC2119 キーワード検証は
`requirement` だけを対象にするので、この 2 型の様式はレビューで守る。

## 根拠

受け皿の無さを抑止で隠さず、型を足して実接続で warning を消す。文書は既に存在しており、型が無いのは
構造の欠落である。

分け方の切り口は ISTQB のテストプロセスに置いた。test_plan は工程上の位置で定義でき、テスト分析以降の
型と連続する。quality はその工程の外にある関心事で、工程上の位置を持たない。

## 棄却した代替案

### 1 つの型（例: `policy`）にまとめる

型は 1 つ少なくて済む。ただし、infrastructure が抱えていた「品質系と基盤系の同居」と同じ混在を、
新しい型の中で最初から作ることになる。
