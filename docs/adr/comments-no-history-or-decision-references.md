---
id: "ADR-77525101-dfcc-4237-941c-919b78ca5eeb"
type: adr
name: "コメントに経緯と決定の記録への参照を書かない"
description: "コメントは 3 行以内で現在の動作だけを書き、経緯と ADR・Issue・PR への参照を書かない。利用者向け文字列に ADR 番号を出さない。"
section: dev-build
status: stable
sources:
  - id: "issue-649"
    resource: "https://github.com/yasunori0418/layat/issues/649"
    title: "#649"
origin: "Issue #649 の grilling で確定"
justifies:
  - "QA-1b4601e5-33cb-435f-812e-052cc159ce8b"
generated:
  by: "claude-code/claude-opus-5-5"
  at: "2026-11-11T00:07:44+09:00"
---
# コメントに経緯と決定の記録への参照を書かない

- ステータス: stable
- 日付: 2026-11-11
- 関連: GitHub Issue #649
- 起点: Issue #649 の grilling で確定

## 背景

エージェントによる実装が積み重なり、コードのコメントが肥大して読みにくくなった。コメントには
変更の経緯や ADR・Issue・PR への参照が書かれ、ヘルプやエラーメッセージなどの利用者向け文字列にも
ADR 番号が出ていた。この規約は CLAUDE.md にだけ書かれ、決定の記録は無かった。

## 決定

- 連続するコメント行の散文は、箇条書きも数えて 3 行以内にする
- コメントには現在の動作だけを書き、経緯と ADR・Issue・PR への参照を書かない
- 利用者向け文字列に ADR 番号を出さない
- コメントにコードブロックを残すのは、判定表・利用者向けのサンプル・実行例に限る

## 根拠

経緯と判断の理由は ADR が持つ。コードは、決定を辿らずに読めるほうがよい。ADR は開発用の
文書で、利用者は読む前提に無いので、番号が出ても意味を成さない。

## 棄却した代替案

### 利用者向け文字列の規則を `requirement` にする

この規則は製品の振る舞いではなく開発の進め方を縛る規範なので、`quality` に置く。
