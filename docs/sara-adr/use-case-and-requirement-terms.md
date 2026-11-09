---
id: "SDR-78d55d19-9e0f-4121-ae86-8e14095e4eba"
type: sara_adr
name: "use_case を要求、requirement を要件に対応させる"
description: "型名 use_case に日本語の要求を、requirement に日本語の要件を対応させ、モデルの読み方に書く。"
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
# use_case を要求、requirement を要件に対応させる

- ステータス: stable
- 日付: 2026-11-09
- 関連: GitHub Issue #648, `docs/model.yaml`, `docs/model.md`
- 起点: epic #648 の grilling セッション（2026-11-08）で確定

## 背景

Issue #500 の基本方針は「要求と要件、それに関わる設計、そこで発生するリスクを、テストプロセスでカバーできるドキュメントの生態系」と述べ、要求と要件を分けて使っている。`docs/model.yaml` の型名は英語で、どちらの語がどの型に当たるかが定まっていなかった。

## 決定

`use_case` を要求、`requirement` を要件に対応させ、`docs/model.md` の型構成の表に書く。型名は変えない。

## 根拠

要求は利用者が何をしたいかを述べ、要件はそれに応えるためにプロダクトが満たす条件を述べる。`use_case` は「ユーザーがどう使うか」を、`requirement` は「何を満たすべきか」を書く型なので、それぞれの語と一致する。

型名は他のプロジェクトでも使える汎用名に保っているので、日本語の語との対応は型の読み方として書く。

## 棄却した代替案

### `requirement` を要求のまま呼び、`use_case` には別の語を当てる

基本方針の「要求と要件」の 2 語のうち、要件に当たる型が無くなる。
